import type { Asset, AssetViewModel, CustomAssetType, LocationAsset, LocationSummary, ParentTargetViewModel } from '$lib/domain/inventory';
import { compareNaturalText } from './textCollation';

export function topLevelLocations(assets: Asset[]): LocationSummary[] {
  return assets
    .filter((asset) => asset.kind === 'location' && asset.parentAssetId === null && asset.lifecycleState === 'active')
    .map((location) => ({
      location: location as LocationAsset,
      assetCount: containedAssets(assets, location.id).length
    }))
    .sort((left, right) => compareNaturalText(left.location.title, right.location.title));
}

export function homeLocationPreview(locations: LocationSummary[]): LocationSummary[] {
  return locations.slice(0, 9);
}

export function recentlyChangedAssets(assets: Asset[]): AssetViewModel[] {
  return assets
    .filter((asset) => asset.lifecycleState === 'active')
    .sort(compareRecentlyChangedAssets)
    .slice(0, 6)
    .map((asset) => withTrail(asset, assets));
}

function compareRecentlyChangedAssets(left: Asset, right: Asset): number {
  const leftUpdatedAt = left.updatedAt ? Date.parse(left.updatedAt) : Number.NaN;
  const rightUpdatedAt = right.updatedAt ? Date.parse(right.updatedAt) : Number.NaN;
  const leftHasTimestamp = Number.isFinite(leftUpdatedAt);
  const rightHasTimestamp = Number.isFinite(rightUpdatedAt);
  if (leftHasTimestamp !== rightHasTimestamp) {
    return leftHasTimestamp ? -1 : 1;
  }
  if (leftHasTimestamp && rightHasTimestamp && leftUpdatedAt !== rightUpdatedAt) {
    return rightUpdatedAt - leftUpdatedAt;
  }
  return right.id.localeCompare(left.id);
}

export function containedAssets(assets: Asset[], parentAssetId: string): AssetViewModel[] {
  return assets
    .filter((asset) => asset.parentAssetId === parentAssetId && asset.lifecycleState === 'active')
    .map((asset) => withTrail(asset, assets));
}

export function parentTargets(assets: Asset[]): ParentTargetViewModel[] {
  return assets
    .filter(isActiveParentTarget)
    .map((asset) => withTrail(asset, assets));
}

export function moveParentTargets(assets: Asset[], movingAssetId: string): ParentTargetViewModel[] {
  const movingAsset = assets.find((asset) => asset.id === movingAssetId);
  if (!movingAsset) {
    return [];
  }
  const sameInventoryAssets = assets.filter(
    (asset) => asset.tenantId === movingAsset.tenantId && asset.inventoryId === movingAsset.inventoryId
  );
  const excluded = descendantIds(sameInventoryAssets, movingAssetId);
  excluded.add(movingAssetId);
  return parentTargets(sameInventoryAssets).filter((asset) => !excluded.has(asset.id));
}

export function labelAssets(items: Asset[], customAssetTypes: CustomAssetType[]): Asset[] {
  return items.map((asset) => labelAsset(asset, customAssetTypes));
}

export function labelAsset(asset: Asset, customAssetTypes: CustomAssetType[]): Asset {
  if (asset.customAssetTypeLabel || !asset.customAssetTypeId) {
    return asset;
  }
  return {
    ...asset,
    customAssetTypeLabel: customAssetTypes.find((assetType) => assetType.id === asset.customAssetTypeId)?.displayName
  };
}

export function detailAssetList(assets: Asset[], loadedAssetDetail: Asset | null, customAssetTypes: CustomAssetType[]): Asset[] {
  if (!loadedAssetDetail || assets.some((asset) => asset.id === loadedAssetDetail.id)) {
    return assets;
  }
  return [labelAsset(loadedAssetDetail, customAssetTypes), ...assets];
}

export function selectedAssetForDetail(
  selectedAssetId: string | null,
  assets: Asset[],
  loadedAssetDetail: Asset | null,
  customAssetTypes: CustomAssetType[]
): Asset | null {
  if (!selectedAssetId) {
    return null;
  }
  if (loadedAssetDetail?.id === selectedAssetId) {
    return labelAsset(loadedAssetDetail, customAssetTypes);
  }
  return assets.find((asset) => asset.id === selectedAssetId) ?? null;
}

export function withTrail<T extends Asset>(asset: T, assets: Asset[]): T & { containmentTrail: string } {
  return {
    ...asset,
    containmentTrail: containmentTrail(asset, assets)
  };
}

export function containmentTrail(asset: Asset, assets: Asset[]): string {
  const trail: string[] = [];
  let parentId = asset.parentAssetId;
  const seen = new Set<string>();
  while (parentId && !seen.has(parentId)) {
    seen.add(parentId);
    const parent = assets.find((candidate) => candidate.id === parentId);
    if (!parent) {
      break;
    }
    trail.unshift(parent.title);
    parentId = parent.parentAssetId;
  }
  return trail.length > 0 ? trail.join(' / ') : 'Inventory root';
}

export function filterAssets(assets: Asset[], query: string): Asset[] {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return [];
  }
  return assets
    .map((asset, index) => ({ asset, index, score: searchMatchScore(asset, normalized) }))
    .filter((match) => match.score !== null)
    .sort((left, right) => left.score! - right.score! || left.index - right.index)
    .map((match) => match.asset);
}

function searchMatchScore(asset: Asset, query: string): number | null {
  const title = asset.title.toLowerCase();
  if (title === query) {
    return 0;
  }
  if (title.startsWith(query)) {
    return 1;
  }
  if (title.includes(query)) {
    return 2;
  }
  if (asset.description.toLowerCase().includes(query)) {
    return 3;
  }
  if (asset.customAssetTypeLabel?.toLowerCase().includes(query)) {
    return 4;
  }
  return null;
}

function isActiveParentTarget(asset: Asset): asset is Asset & Pick<ParentTargetViewModel, 'kind' | 'lifecycleState'> {
  return (asset.kind === 'container' || asset.kind === 'location') && asset.lifecycleState === 'active';
}

function descendantIds(assets: Asset[], assetId: string): Set<string> {
  const descendants = new Set<string>();
  const queue = [assetId];
  while (queue.length > 0) {
    const parentId = queue.shift();
    for (const asset of assets) {
      if (asset.parentAssetId === parentId && !descendants.has(asset.id)) {
        descendants.add(asset.id);
        queue.push(asset.id);
      }
    }
  }
  return descendants;
}
