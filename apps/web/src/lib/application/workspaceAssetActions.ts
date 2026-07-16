import type { AssetRouteAction } from './workspaceRoute';
import { workspaceRouteHref } from './workspaceRoute';
import type { Asset, AssetAttachment, AssetViewModel } from '$lib/domain/inventory';

export function assetActionHref(asset: Asset, action: Exclude<AssetRouteAction, null>): string {
  return workspaceRouteHref(assetActionRoute(asset, action), asset.tenantId, asset.inventoryId);
}

export function assetDetailHref(asset: AssetViewModel): string {
  return workspaceRouteHref(
    asset.kind === 'location'
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
        },
    asset.tenantId,
    asset.inventoryId
  );
}

export function attachmentDeleteHref(asset: AssetViewModel, attachment: AssetAttachment): string {
  return workspaceRouteHref(
    {
      mode: 'asset',
      tenantId: asset.tenantId,
      inventoryId: asset.inventoryId,
      assetId: asset.id,
      attachmentId: attachment.id,
      attachmentAction: 'delete'
    },
    asset.tenantId,
    asset.inventoryId
  );
}

export function assetActionIsAvailable(
  asset: AssetViewModel,
  action: Exclude<AssetRouteAction, null>,
  state: { canEdit: boolean; saving: boolean }
): boolean {
  if (!state.canEdit || state.saving) {
    return false;
  }
  if (action === 'delete') {
    return true;
  }
  if (action === 'checkout') {
    return asset.lifecycleState === 'active' && !asset.currentCheckout;
  }
  if (action === 'return') {
    return !!asset.currentCheckout;
  }
  if (action === 'restore') {
    return asset.lifecycleState === 'archived';
  }
  return asset.lifecycleState === 'active';
}

function assetActionRoute(asset: Asset, action: Exclude<AssetRouteAction, null>) {
  return {
    mode: 'asset' as const,
    tenantId: asset.tenantId,
    inventoryId: asset.inventoryId,
    locationId: action === 'edit' && asset.kind === 'location' ? asset.id : null,
    assetId: asset.id,
    assetAction: action,
    action: action === 'edit' ? 'edit' as const : null
  };
}
