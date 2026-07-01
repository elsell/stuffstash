import type { AssetKind, AssetSummary } from '../../domain/assets/AssetSummary';
import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

export type ParentLookupResult = {
  readonly id: string;
  readonly title: string;
  readonly kind: AssetKind;
  readonly subtitle: string;
  readonly pathLabel: string;
  readonly selectionHint: string;
  readonly willPromoteToContainer: boolean;
};

export class ParentLookupQuery {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(query: string): Promise<readonly ParentLookupResult[]> {
    const trimmed = query.trim();
    if (trimmed.length === 0) {
      const inventory = await this.inventories.getDefaultInventorySummary();
      return inventory.assets.slice(0, 5).map(toParentLookupResult);
    }

    const assets = await this.inventories.searchAssets(trimmed);
    return prioritizeExactTitleMatches(assets, trimmed).slice(0, 6).map(toParentLookupResult);
  }
}

function prioritizeExactTitleMatches(
  assets: readonly AssetSummary[],
  query: string
): readonly AssetSummary[] {
  const normalizedQuery = normalizeParentName(query);
  const exactMatches = assets.filter((asset) => normalizeParentName(asset.title) === normalizedQuery);
  const fuzzyMatches = assets.filter((asset) => normalizeParentName(asset.title) !== normalizedQuery);

  return [...exactMatches, ...fuzzyMatches];
}

function toParentLookupResult(asset: AssetSummary): ParentLookupResult {
  const willPromoteToContainer = asset.kind === 'item';

  return {
    id: asset.id,
    title: asset.title,
    kind: asset.kind,
    subtitle: parentSubtitle(asset),
    pathLabel: asset.locationTrail.length > 1
      ? asset.locationTrail.slice(1).join(' / ')
      : asset.title,
    selectionHint: willPromoteToContainer
      ? 'Will become a container for this item'
      : parentKindLabel(asset.kind),
    willPromoteToContainer
  };
}

function parentSubtitle(asset: AssetSummary): string {
  if (asset.locationTrail.length > 1) {
    return asset.locationTrail.slice(0, -1).join(' / ');
  }

  return asset.locationLabel === 'Inventory root' ? 'No parent' : asset.locationLabel;
}

function parentKindLabel(kind: AssetKind): string {
  switch (kind) {
    case 'location':
      return 'Location';
    case 'container':
      return 'Container';
    case 'item':
      return 'Item';
  }
}

function normalizeParentName(value: string): string {
  return value.trim().toLocaleLowerCase();
}
