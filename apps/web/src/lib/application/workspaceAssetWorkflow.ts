import type {
  AddAssetDraft,
  AddAssetSaveResult,
  AddAssetSubmission,
  Asset,
  AssetAttachment,
  Inventory,
  SelectedPhoto,
  WorkspaceData,
  WorkspaceMode
} from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import type { WorkspaceRouteState } from './workspaceRoute';

type AssetCreateRepository = Pick<InventoryRepository, 'createAsset' | 'selectAssetLifecycle' | 'uploadAssetPhoto'>;

interface PhotoUploadResult {
  uploaded: UploadedPhoto[];
  failures: number;
}

export interface CreateAssetWorkflowResult {
  data: WorkspaceData;
  saveResult: AddAssetSaveResult;
  message?: string;
  error?: string;
  closeAdd: boolean;
  mode?: WorkspaceMode;
  clearDetail?: boolean;
  selectedAsset?: Asset;
  route?: Partial<WorkspaceRouteState>;
}

export async function createAssetWorkflow(
  repository: AssetCreateRepository,
  data: WorkspaceData,
  inventory: Inventory,
  draft: AddAssetSubmission
): Promise<CreateAssetWorkflowResult> {
  let createdParent: Asset | null = null;
  let createdAsset: Asset | null = null;
  let savedAsset: Asset | null = null;
  let uploadResult: PhotoUploadResult = { uploaded: [], failures: 0 };

  try {
    createdParent = draft.parentQuickCreate
      ? await repository.createAsset(data.context.selectedTenantId, inventory.id, {
          kind: draft.parentQuickCreate.kind,
          title: draft.parentQuickCreate.title,
          description: '',
          parentAssetId: draft.parentAssetId,
          customFields: {},
          photos: []
        })
      : null;

    const { parentQuickCreate: _parentQuickCreate, ...assetDraft } = draft;
    const childDraft: AddAssetDraft = {
      ...assetDraft,
      parentAssetId: createdParent?.id ?? draft.parentAssetId
    };
    createdAsset = await repository.createAsset(data.context.selectedTenantId, inventory.id, childDraft);
    uploadResult = await uploadPhotos(repository, createdAsset, draft.photos);
    savedAsset = assetWithPrimaryPhoto(createdAsset, uploadResult.uploaded[0]);

    if (data.context.assetLifecycleState !== 'active') {
      return {
        data: await repository.selectAssetLifecycle(createdAsset.tenantId, createdAsset.inventoryId, 'active'),
        saveResult: { saved: true },
        message: createAssetMessage(createdAsset, uploadResult, createdParent),
        closeAdd: true,
        mode: createdAsset.kind === 'location' ? 'location' : 'asset',
        selectedAsset: savedAsset,
        route: createdAssetRoute(createdAsset)
      };
    }

    const nextData = prependCreatedAssets(data, savedAsset, createdParent);
    const message = createAssetMessage(createdAsset, uploadResult, createdParent);

    if (uploadResult.failures > 0) {
      return {
        data: nextData,
        saveResult: { saved: true },
        message,
        closeAdd: true,
        mode: savedAsset.kind === 'location' ? 'location' : 'asset',
        selectedAsset: savedAsset,
        route: createdAssetRoute(savedAsset)
      };
    }

    return {
      data: nextData,
      saveResult: { saved: true },
      message,
      closeAdd: true,
      mode: savedAsset.kind === 'location' ? 'location' : 'asset',
      selectedAsset: savedAsset,
      route: createdAssetRoute(savedAsset)
    };
  } catch (caught) {
    if (createdAsset) {
      const selectedAsset = savedAsset ?? createdAsset;
      const failure = caught instanceof Error ? caught.message : 'Action failed.';
      return {
        data,
        saveResult: { saved: true },
        message: createAssetMessage(createdAsset, uploadResult, createdParent),
        error: `Saved ${createdAsset.title}, but could not refresh the active view. ${failure}`,
        closeAdd: true,
        mode: selectedAsset.kind === 'location' ? 'location' : 'asset',
        selectedAsset,
        route: createdAssetRoute(selectedAsset)
      };
    }
    const nextData =
      createdParent && data.context.assetLifecycleState === 'active' && !data.assets.some((asset) => asset.id === createdParent?.id)
        ? { ...data, assets: [createdParent, ...data.assets] }
        : data;
    const failure = caught instanceof Error ? caught.message : 'Action failed.';
    return {
      data: nextData,
      saveResult: createdParent ? { saved: false, createdParentId: createdParent.id } : { saved: false },
      error: createdParent ? `Created ${createdParent.title}, but could not save ${draft.title}. ${failure}` : failure,
      closeAdd: false
    };
  }
}

export function replaceWorkspaceAsset(data: WorkspaceData, asset: Asset): WorkspaceData {
  if (asset.tenantId !== data.context.selectedTenantId || asset.inventoryId !== data.context.selectedInventoryId) {
    return data;
  }
  if (asset.lifecycleState !== data.context.assetLifecycleState) {
    return data;
  }
  const existing = data.assets.some(
    (candidate) =>
      candidate.tenantId === asset.tenantId && candidate.inventoryId === asset.inventoryId && candidate.id === asset.id
  );
  return {
    ...data,
    assets: existing
      ? data.assets.map((candidate) =>
          candidate.tenantId === asset.tenantId && candidate.inventoryId === asset.inventoryId && candidate.id === asset.id
            ? asset
            : candidate
        )
      : [asset, ...data.assets]
  };
}

function prependCreatedAssets(data: WorkspaceData, asset: Asset, createdParent: Asset | null): WorkspaceData {
  return { ...data, assets: createdParent ? [asset, createdParent, ...data.assets] : [asset, ...data.assets] };
}

function createdAssetRoute(asset: Asset): Partial<WorkspaceRouteState> {
  return asset.kind === 'location'
    ? {
        mode: 'location',
        tenantId: asset.tenantId,
        inventoryId: asset.inventoryId,
        locationId: asset.id
      }
    : {
        mode: 'asset',
        tenantId: asset.tenantId,
        inventoryId: asset.inventoryId,
        assetId: asset.id
      };
}

interface UploadedPhoto {
  attachment: AssetAttachment;
  photo: SelectedPhoto;
}

function assetWithPrimaryPhoto(asset: Asset, uploaded: UploadedPhoto | undefined): Asset {
  return uploaded
    ? {
        ...asset,
        photo: {
          id: uploaded.attachment.id,
          assetId: asset.id,
          url: uploaded.photo.previewUrl ?? uploaded.attachment.thumbnailUrl ?? '',
          alt: asset.title
        }
      }
    : asset;
}

async function uploadPhotos(
  repository: Pick<InventoryRepository, 'uploadAssetPhoto'>,
  asset: Asset,
  photos: SelectedPhoto[]
): Promise<PhotoUploadResult> {
  let failures = 0;
  const uploaded: UploadedPhoto[] = [];
  for (const photo of photos) {
    try {
      uploaded.push({
        attachment: await repository.uploadAssetPhoto(asset.tenantId, asset.inventoryId, asset.id, photo),
        photo
      });
    } catch {
      failures += 1;
    }
  }
  return { uploaded, failures };
}

function createAssetMessage(asset: Asset, uploadResult: PhotoUploadResult, createdParent: Asset | null): string {
  const locationSuffix = createdParent ? ` in ${createdParent.title}` : '';
  const uploadSuffix =
    uploadResult.uploaded.length > 0 ? ` with ${photoUploadCountLabel(uploadResult.uploaded.length)}` : '';
  const savedMessage = `Saved ${asset.title}${locationSuffix}${uploadSuffix}.`;
  if (uploadResult.failures > 0) {
    return `${savedMessage} ${photoUploadCountLabel(uploadResult.failures)} failed.`;
  }
  return savedMessage;
}

function photoUploadCountLabel(count: number): string {
  return `${count} ${count === 1 ? 'photo upload' : 'photo uploads'}`;
}
