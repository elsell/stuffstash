import type { Asset, AssetViewModel, LocationSummary } from '$lib/domain/inventory';

export function topLevelLocations(assets: Asset[]): LocationSummary[] {
  return assets
    .filter((asset) => asset.kind === 'location' && asset.parentAssetId === null && asset.lifecycleState === 'active')
    .map((location) => ({
      location,
      assetCount: containedAssets(assets, location.id).length
    }));
}

export function recentlyAddedAssets(assets: Asset[]): Asset[] {
  return assets
    .filter((asset) => asset.kind !== 'location' && asset.lifecycleState === 'active')
    .slice(0, 6);
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
  return assets.filter((asset) => {
    return (
      asset.title.toLowerCase().includes(normalized) ||
      asset.description.toLowerCase().includes(normalized) ||
      asset.customAssetTypeLabel?.toLowerCase().includes(normalized)
    );
  });
}
