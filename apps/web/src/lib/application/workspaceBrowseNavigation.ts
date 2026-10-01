import { t } from '$lib/presentation/localization';
import type { Asset, AssetLifecycleFilter, LocationAsset } from '$lib/domain/inventory';
import { workspaceRouteHref } from './workspaceRoute';

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
  return t('assets.visibleCount', { count });
}

export function homeHeadingPresentation(lifecycleState: AssetLifecycleFilter): HomeHeadingPresentation {
  if (lifecycleState === 'archived') {
    return {
      title: t('web.workspaceBrowseNavigation.archivedAssets'),
      description: t('web.workspaceBrowseNavigation.assetsRemovedFromActiveBrowsing')
    };
  }
  return {
    title: t('web.workspaceBrowseNavigation.home'),
    description: t('web.workspaceBrowseNavigation.recentlyChangedAndThePlacesWhereYourThingsLive')
  };
}

export function homeLifecycleOptions(tenantId: string | null, inventoryId: string | null): HomeLifecycleOption[] {
  return [
    { value: 'active', label: t('web.workspaceBrowseNavigation.active'), href: homeLifecycleHref(tenantId, inventoryId, 'active') },
    { value: 'archived', label: t('web.workspaceBrowseNavigation.archived'), href: homeLifecycleHref(tenantId, inventoryId, 'archived') }
  ];
}

export function homeRecentEmptyState(): HomeEmptyStatePresentation {
  return { message: t('web.workspaceBrowseNavigation.noItemsOrContainersYet') };
}

export function homeArchivedEmptyState(): HomeEmptyStatePresentation {
  return { title: t('web.workspaceBrowseNavigation.noArchivedAssets') };
}

export function homeLocationsEmptyState(): HomeEmptyStatePresentation {
  return {
    title: t('web.workspaceBrowseNavigation.noLocationsYet'),
    message: t('web.workspaceBrowseNavigation.locationsMakeBrowsingEasierButYouCanCaptureAn'),
    actionLabel: t('web.workspaceBrowseNavigation.addFirstLocation'),
    secondaryActionLabel: t('web.workspaceBrowseNavigation.addItem')
  };
}

export function homeCreateLocationDenied(): HomeDeniedPresentation {
  return {
    id: 'home-add-location-denied',
    message: t('web.workspaceBrowseNavigation.creatingLocationsIsUnavailableForThisInventory')
  };
}

export function locationEmptyState(canCreateAsset: boolean): LocationEmptyStatePresentation {
  return {
    title: t('web.workspaceBrowseNavigation.noStuffHereYet'),
    message: canCreateAsset ? t('web.workspaceBrowseNavigation.addAnItemOrMoveExistingStuffIntoThis') : t('web.workspaceBrowseNavigation.thisLocationIsEmpty'),
    actionLabel: t('web.workspaceBrowseNavigation.addItemHere'),
    deniedMessage: t('web.workspaceBrowseNavigation.addingItemsIsUnavailableForThisInventory')
  };
}
