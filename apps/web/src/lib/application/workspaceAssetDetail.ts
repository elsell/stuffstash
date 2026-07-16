import type { Asset, AssetAttachment, AssetCheckout, CustomFieldDefinition, WorkspaceData } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import { replaceWorkspaceAsset } from './workspaceAssetWorkflow';

type AssetDetailRepository = Pick<InventoryRepository, 'getAsset' | 'listAssetAttachments' | 'listAssetCheckoutHistory'>;

export interface LoadWorkspaceAssetDetailResult {
  loaded: boolean;
  asset: Asset | null;
  attachments: AssetAttachment[];
  checkoutHistory: AssetCheckout[];
  error: string;
}

export interface AssetDetailIdentity {
  tenantId: string;
  inventoryId: string;
  assetId: string;
}

export interface WorkspaceAssetDetailState {
  data: WorkspaceData;
  loadedAssetDetail: Asset;
  selectedAssetId: string;
  selectedAssetAttachments: AssetAttachment[];
  selectedAssetCheckoutHistory: AssetCheckout[];
  mode: 'asset';
}

export interface AssetDetailStatusPresentation {
  kind: 'edit-unavailable' | 'files-empty';
  message: string;
}

export interface AssetDetailFieldGroups {
  populated: CustomFieldDefinition[];
  unset: CustomFieldDefinition[];
}

export function partitionAssetDetailFields(
  definitions: CustomFieldDefinition[],
  values: Record<string, unknown> = {}
): AssetDetailFieldGroups {
  return definitions.reduce<AssetDetailFieldGroups>((groups, definition) => {
    const value = values[definition.key];
    const target = value === undefined || value === null || value === '' ? groups.unset : groups.populated;
    target.push(definition);
    return groups;
  }, { populated: [], unset: [] });
}

export async function loadWorkspaceAssetDetail(
  repository: AssetDetailRepository,
  tenantId: string,
  inventoryId: string,
  assetId: string
): Promise<LoadWorkspaceAssetDetailResult> {
  try {
    const asset = await repository.getAsset(tenantId, inventoryId, assetId);
    const attachments = await repository.listAssetAttachments(tenantId, inventoryId, assetId);
    const checkoutHistory = await repository.listAssetCheckoutHistory(tenantId, inventoryId, assetId);
    return {
      loaded: true,
      asset,
      attachments,
      checkoutHistory,
      error: ''
    };
  } catch (caught) {
    return {
      loaded: false,
      asset: null,
      attachments: [],
      checkoutHistory: [],
      error: caught instanceof Error ? caught.message : 'Asset could not be loaded.'
    };
  }
}

export function refreshWorkspaceAssetAttachments(
  repository: AssetDetailRepository,
  identity: AssetDetailIdentity
): Promise<AssetAttachment[]> {
  return repository.listAssetAttachments(identity.tenantId, identity.inventoryId, identity.assetId);
}

export function applyLoadedWorkspaceAssetDetail(data: WorkspaceData, result: LoadedWorkspaceAssetDetail): WorkspaceAssetDetailState {
  return {
    data: replaceWorkspaceAsset(data, result.asset),
    loadedAssetDetail: result.asset,
    selectedAssetId: result.asset.id,
    selectedAssetAttachments: result.attachments,
    selectedAssetCheckoutHistory: result.checkoutHistory,
    mode: 'asset'
  };
}

type LoadedWorkspaceAssetDetail = LoadWorkspaceAssetDetailResult & {
  loaded: true;
  asset: Asset;
};

export function assetDescriptionText(description: string): string {
  return description || 'No description.';
}

export function assetEditUnavailableStatus(canEdit: boolean): AssetDetailStatusPresentation | null {
  if (canEdit) {
    return null;
  }
  return {
    kind: 'edit-unavailable',
    message: 'Edit actions require asset edit access.'
  };
}

export function assetFilesStatus(fileCount: number): AssetDetailStatusPresentation | null {
  if (fileCount > 0) {
    return null;
  }
  return {
    kind: 'files-empty',
    message: 'No active files.'
  };
}
