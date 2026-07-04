import type { Asset, AssetViewModel, CustomAssetType, LocationAsset, LocationSummary } from '$lib/domain/inventory';

export function topLevelLocations(assets: Asset[]): LocationSummary[] {
  return assets
    .filter((asset) => asset.kind === 'location' && asset.parentAssetId === null && asset.lifecycleState === 'active')
    .map((location) => ({
      location: location as LocationAsset,
      assetCount: containedAssets(assets, location.id).length
    }));
}

export function recentlyAddedAssets(assets: Asset[]): AssetViewModel[] {
  return assets
    .filter((asset) => asset.kind !== 'location' && asset.lifecycleState === 'active')
    .slice(0, 6)
    .map((asset) => withTrail(asset, assets));
}

export function containedAssets(assets: Asset[], parentAssetId: string): AssetViewModel[] {
  return assets
    .filter((asset) => asset.parentAssetId === parentAssetId && asset.lifecycleState === 'active')
    .map((asset) => withTrail(asset, assets));
}

export function parentTargets(assets: Asset[]): AssetViewModel[] {
  return assets
    .filter((asset) => (asset.kind === 'container' || asset.kind === 'location') && asset.lifecycleState === 'active')
    .map((asset) => withTrail(asset, assets));
}

export function moveParentTargets(assets: Asset[], movingAssetId: string): AssetViewModel[] {
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

export function withTrail(asset: Asset, assets: Asset[]): AssetViewModel {
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
