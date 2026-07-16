import type { AssetId, AssetKind, AssetSummary } from '../../domain/assets/AssetSummary';
import type { InventoryId, TenantId } from '../../domain/inventories/InventorySummary';
import { toAssetCardViewModel } from './AssetViewModels';

export type InventoryMapAssetViewModel = {
  readonly id: string;
  readonly parentAssetId?: string;
  readonly title: string;
  readonly kind: AssetKind;
  readonly kindLabel: string;
  readonly customTypeLabel?: string;
  readonly description: string;
  readonly placementLabel: string;
  readonly parentPlacementLabel: string;
  readonly updatedAtLabel: string;
  readonly childCount: number;
  readonly childCountLabel: string;
  readonly canContainAssets: boolean;
  readonly canAddContainedAssets: boolean;
  readonly imagePlaceholderLabel: string;
  readonly photoLabel: string;
  readonly checkedOutLabel?: string;
  readonly photo?: {
    readonly uri: string;
    readonly headers?: Readonly<Record<string, string>>;
  };
};

export type InventoryMapViewModel = {
  readonly sessionScopeId: string;
  readonly tenantId: TenantId;
  readonly inventoryId: InventoryId;
  readonly inventoryName: string;
  readonly canCreateAsset: boolean;
  readonly assets: readonly InventoryMapAssetViewModel[];
};

export interface InventoryMapAssetRepository {
  listActiveInventoryMapAssets(): Promise<{
    readonly sessionScopeId: string;
    readonly tenantId: TenantId;
    readonly inventoryId: InventoryId;
    readonly inventoryName: string;
    readonly permissions: readonly string[];
    readonly assets: readonly AssetSummary[];
  }>;
}

export class InventoryMapQuery {
  constructor(private readonly inventoryMapAssets: InventoryMapAssetRepository) {}

  async execute(): Promise<InventoryMapViewModel> {
    const inventory = await this.inventoryMapAssets.listActiveInventoryMapAssets();
    const activeAssets = inventory.assets;
    const childCounts = countChildrenByParent(activeAssets);
    const canCreateAsset = inventory.permissions.includes('create_asset');
    const canEditAsset = inventory.permissions.includes('edit_asset');

    const assets = activeAssets.map((asset) => toInventoryMapAssetViewModel(asset, {
      childCount: childCounts.get(asset.id) ?? 0,
      canCreateAsset,
      canEditAsset
    }));

    return {
      sessionScopeId: inventory.sessionScopeId,
      tenantId: inventory.tenantId,
      inventoryId: inventory.inventoryId,
      inventoryName: inventory.inventoryName,
      canCreateAsset,
      assets: flattenMapAssets(assets)
    };
  }
}

function toInventoryMapAssetViewModel(
  asset: AssetSummary,
  options: {
    readonly childCount: number;
    readonly canCreateAsset: boolean;
    readonly canEditAsset: boolean;
  }
): InventoryMapAssetViewModel {
  const card = toAssetCardViewModel(asset);
  const canContainAssets = asset.kind === 'container' || asset.kind === 'location';

  return {
    id: card.id,
    parentAssetId: asset.parentAssetId,
    title: card.title,
    kind: asset.kind,
    kindLabel: card.kindLabel,
    customTypeLabel: card.customTypeLabel,
    description: card.description,
    placementLabel: card.locationTrailLabel,
    parentPlacementLabel: labelParentPlacement(asset),
    updatedAtLabel: card.updatedAtLabel,
    childCount: options.childCount,
    childCountLabel: labelChildCount(options.childCount),
    canContainAssets,
    canAddContainedAssets: canContainAssets && options.canCreateAsset && options.canEditAsset,
    imagePlaceholderLabel: card.imagePlaceholderLabel,
    photoLabel: card.photoLabel,
    checkedOutLabel: card.checkedOutLabel,
    photo: card.photo
  };
}

function labelParentPlacement(asset: AssetSummary): string {
  if (asset.parentLocationTrail.length === 0) {
    return 'Inventory root';
  }

  return asset.parentLocationTrail.map((segment) => segment.title).join(' / ');
}

function countChildrenByParent(assets: readonly AssetSummary[]): Map<AssetId, number> {
  const counts = new Map<AssetId, number>();

  for (const asset of assets) {
    if (!asset.parentAssetId) {
      continue;
    }
    counts.set(asset.parentAssetId, (counts.get(asset.parentAssetId) ?? 0) + 1);
  }

  return counts;
}

function flattenMapAssets(assets: readonly InventoryMapAssetViewModel[]): readonly InventoryMapAssetViewModel[] {
  const childrenByParent = new Map<string, InventoryMapAssetViewModel[]>();
  const rootKey = '__inventory_root__';

  for (const asset of assets) {
    const key = asset.parentAssetId ?? rootKey;
    const children = childrenByParent.get(key) ?? [];
    children.push(asset);
    childrenByParent.set(key, children);
  }

  for (const children of childrenByParent.values()) {
    children.sort(compareSiblingMapAssets);
  }

  const flattened: InventoryMapAssetViewModel[] = [];
  const visit = (parentKey: string) => {
    for (const asset of childrenByParent.get(parentKey) ?? []) {
      flattened.push(asset);
      visit(asset.id);
    }
  };

  visit(rootKey);
  return flattened;
}

function compareSiblingMapAssets(left: InventoryMapAssetViewModel, right: InventoryMapAssetViewModel): number {
  const kindOrder = mapKindRank(left.kind) - mapKindRank(right.kind);
  if (kindOrder !== 0) {
    return kindOrder;
  }

  const titleOrder = stableTextSortKey(left.title).localeCompare(stableTextSortKey(right.title));
  if (titleOrder !== 0) {
    return titleOrder;
  }

  return left.id.localeCompare(right.id);
}

function mapKindRank(kind: AssetKind): number {
  return kind === 'item' ? 1 : 0;
}

function stableTextSortKey(value: string): string {
  return value.trim().normalize('NFKD').toLowerCase();
}

function labelChildCount(count: number): string {
  if (count === 0) {
    return 'Empty';
  }
  if (count === 1) {
    return '1 inside';
  }
  return `${count.toString()} inside`;
}
