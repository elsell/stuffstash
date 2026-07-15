import type { Asset, AssetAttachment, AssetCheckout, WorkspaceData } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import { isAuthenticationRequiredError } from './authenticationRequired';
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
    if (isAuthenticationRequiredError(caught)) throw caught;
    return {
      loaded: false,
      asset: null,
      attachments: [],
      checkoutHistory: [],
      error: assetDetailFailureMessage(caught)
    };
  }
}

export function assetDetailFailureMessage(caught: unknown): string {
  const safeForUser = typeof caught === 'object' && caught !== null &&
    (caught as { safeForUser?: unknown }).safeForUser === true;
  if (safeForUser && caught instanceof Error && caught.message.trim() && !isGenericAdapterMessage(caught)) {
    return caught.message.trim();
  }
  return 'Asset details could not be loaded. Try again.';
}

function isGenericAdapterMessage(caught: Error): boolean {
  const status = 'status' in caught && typeof caught.status === 'number' ? caught.status : 0;
  if (status !== 400 && status !== 422) return false;
  return /^(bad request|invalid request|unprocessable entity|validation failed|request failed)\.?$/i.test(caught.message.trim());
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
