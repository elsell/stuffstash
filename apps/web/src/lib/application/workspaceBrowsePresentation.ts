import type { Asset } from '$lib/domain/inventory';

export interface PlaceBrowseSummary {
  asset: Asset;
  containedCount: number;
  recentContainedNames: string[];
}

export type BrowseFailurePhase = 'initial' | 'replacement' | 'append' | 'map';

const browseFailureFallbacks: Record<BrowseFailurePhase, string> = {
  initial: 'Browse could not be loaded. Try again.',
  replacement: 'Browse could not be refreshed. Try again.',
  append: 'More results could not be loaded. Try again.',
  map: 'Map could not be loaded. Try again.'
};

export function browseFailureMessage(error: unknown, phase: BrowseFailurePhase): string {
  const safeForUser = typeof error === 'object' && error !== null &&
    'safeForUser' in error && error.safeForUser === true;
  if (safeForUser && error instanceof Error && error.message.trim()) {
    return error.message.trim();
  }
  return browseFailureFallbacks[phase];
}

export function buildPlaceBrowseSummaries(places: Asset[], allAssets: Asset[]): PlaceBrowseSummary[] {
  const children = new Map<string, Asset[]>();
  for (const asset of allAssets) {
    if (!asset.parentAssetId) continue;
    children.set(asset.parentAssetId, [...(children.get(asset.parentAssetId) ?? []), asset]);
  }
  return places.map((place) => {
    const descendants: Asset[] = [];
    const pending = [...(children.get(place.id) ?? [])];
    while (pending.length) {
      const asset = pending.shift()!;
      descendants.push(asset);
      pending.push(...(children.get(asset.id) ?? []));
    }
    descendants.sort((left, right) =>
      (Date.parse(right.updatedAt ?? '') || 0) - (Date.parse(left.updatedAt ?? '') || 0) || right.id.localeCompare(left.id)
    );
    return { asset: place, containedCount: descendants.length, recentContainedNames: descendants.slice(0, 3).map((asset) => asset.title) };
  });
}
