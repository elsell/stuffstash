import type { Asset, AssetLifecycleFilter, LocationAsset } from '$lib/domain/inventory';
import { workspaceRouteHref } from './workspaceRoute';

export type HomeBrowseMode = 'home' | 'locations';

export interface HomeHeadingPresentation {
  title: string;
  description: string;
}

export interface HomeEmptyStatePresentation {
  title?: string;
  message?: string;
  actionLabel?: string;
  secondaryActionLabel?: string;
}

export interface HomeDeniedPresentation {
  id: string;
  message: string;
}

export interface LocationEmptyStatePresentation {
  title: string;
  message: string;
  actionLabel: string;
  deniedMessage: string;
}

export interface HomeLifecycleOption {
  value: AssetLifecycleFilter;
  label: string;
  href: string;
}

export function homeAddLocationHref(tenantId: string | null, inventoryId: string | null): string {
  return workspaceRouteHref({ action: 'add', addKind: 'location' }, tenantId, inventoryId);
}

export function homeAddItemHref(tenantId: string | null, inventoryId: string | null): string {
  return workspaceRouteHref({ action: 'add', addKind: 'item' }, tenantId, inventoryId);
}

export function homeLocationsHref(tenantId: string | null, inventoryId: string | null): string {
  return workspaceRouteHref({ mode: 'browse', browseScope: 'places' }, tenantId, inventoryId);
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
  return workspaceRouteHref({ mode: 'browse', browseScope: 'places' }, location.tenantId, location.inventoryId);
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

export function visibleAssetCountLabel(count: number): string {
  return `${count} visible ${count === 1 ? 'asset' : 'assets'}`;
}

export function homeHeadingPresentation(lifecycleState: AssetLifecycleFilter, browseMode: HomeBrowseMode): HomeHeadingPresentation {
  if (lifecycleState === 'archived') {
    return {
      title: 'Archived assets',
      description: 'Assets removed from active browsing.'
    };
  }
  if (browseMode === 'locations') {
    return {
      title: 'Locations',
      description: 'The places where your things live.'
    };
  }
  return {
    title: 'Home',
    description: 'Recently changed and the places where your things live.'
  };
}

export function homeLifecycleOptions(tenantId: string | null, inventoryId: string | null): HomeLifecycleOption[] {
  return [
    { value: 'active', label: 'Active', href: homeLifecycleHref(tenantId, inventoryId, 'active') },
    { value: 'archived', label: 'Archived', href: homeLifecycleHref(tenantId, inventoryId, 'archived') }
  ];
}

export function homeRecentEmptyState(): HomeEmptyStatePresentation {
  return { message: 'No items or containers yet.' };
}

export function homeArchivedEmptyState(): HomeEmptyStatePresentation {
  return { title: 'No archived assets' };
}

export function homeLocationsEmptyState(browseMode: HomeBrowseMode = 'home'): HomeEmptyStatePresentation {
  if (browseMode === 'locations') {
    return {
      title: 'No locations yet',
      message: 'Add a location to start browsing by place.',
      actionLabel: 'Add first location'
    };
  }
  return {
    title: 'No locations yet',
    message: 'Locations make browsing easier, but you can capture an item now.',
    actionLabel: 'Add first location',
    secondaryActionLabel: 'Add item'
  };
}

export function homeCreateLocationDenied(): HomeDeniedPresentation {
  return {
    id: 'home-add-location-denied',
    message: 'Creating locations is unavailable for this inventory.'
  };
}

export function locationEmptyState(canCreateAsset: boolean): LocationEmptyStatePresentation {
  return {
    title: 'No stuff here yet',
    message: canCreateAsset ? 'Add an item or move existing stuff into this location.' : 'This location is empty.',
    actionLabel: 'Add item here',
    deniedMessage: 'Adding items is unavailable for this inventory.'
  };
}
