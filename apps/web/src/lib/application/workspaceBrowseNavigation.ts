import type { Asset, AssetLifecycleFilter, LocationAsset } from '$lib/domain/inventory';
import { workspaceRouteHref } from './workspaceRoute';

export function homeAddLocationHref(tenantId: string | null, inventoryId: string | null): string {
  return workspaceRouteHref({ action: 'add', addKind: 'location' }, tenantId, inventoryId);
}

export function homeLifecycleHref(
  tenantId: string | null,
  inventoryId: string | null,
  lifecycleState: AssetLifecycleFilter
): string {
  return workspaceRouteHref({ mode: 'home', tenantId, inventoryId, lifecycleState }, tenantId, inventoryId);
}

export function browseAssetHref(asset: Asset): string {
  return workspaceRouteHref({ mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id }, asset.tenantId, asset.inventoryId);
}

export function browseLocationHref(location: LocationAsset): string {
  return workspaceRouteHref(
    { mode: 'location', tenantId: location.tenantId, inventoryId: location.inventoryId, locationId: location.id },
    location.tenantId,
    location.inventoryId
  );
}

export function locationBackHref(location: LocationAsset): string {
  return workspaceRouteHref({ mode: 'locations' }, location.tenantId, location.inventoryId);
}

export function locationEditHref(location: LocationAsset): string {
  return workspaceRouteHref(
    { mode: 'asset', locationId: location.id, assetId: location.id, action: 'edit', assetAction: 'edit' },
    location.tenantId,
    location.inventoryId
  );
}

export function locationAddItemHref(location: LocationAsset): string {
  return workspaceRouteHref(
    { action: 'add', addKind: 'item', addParentAssetId: location.id },
    location.tenantId,
    location.inventoryId
  );
}

export function locationRowHref(asset: Asset): string {
  return asset.kind === 'location' ? browseLocationHref(asset as LocationAsset) : browseAssetHref(asset);
}
