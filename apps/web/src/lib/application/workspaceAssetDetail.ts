import type { Asset, AssetAttachment } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';

type AssetDetailRepository = Pick<InventoryRepository, 'getAsset' | 'listAssetAttachments'>;

export interface LoadWorkspaceAssetDetailResult {
  loaded: boolean;
  asset: Asset | null;
  attachments: AssetAttachment[];
  error: string;
}

export interface AssetDetailIdentity {
  tenantId: string;
  inventoryId: string;
  assetId: string;
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
    return {
      loaded: true,
      asset,
      attachments,
      error: ''
    };
  } catch (caught) {
    return {
      loaded: false,
      asset: null,
      attachments: [],
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
