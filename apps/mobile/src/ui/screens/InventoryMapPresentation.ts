import type {
  InventoryMapAssetViewModel,
  InventoryMapViewModel
} from '../../application/assets/InventoryMapQuery';

export type InventoryMapSurface = 'list' | 'map';

export type InventoryMapColumnViewModel = {
  readonly key: string;
  readonly level: number;
  readonly title: string;
  readonly parentId?: string;
  readonly assets: readonly InventoryMapAssetViewModel[];
  readonly emptyLabel: string;
};

export type InventoryMapBreadcrumbViewModel = {
  readonly key: string;
  readonly level: number;
  readonly title: string;
  readonly assetId?: string;
};

export type InventoryMapSearchMatch = {
  readonly assetId: string;
  readonly openPath: readonly string[];
};

export type InventoryMapRowInteractionState = {
  readonly expanded: boolean;
  readonly highlighted: boolean;
};

export type InventoryMapEmptyColumnActionViewModel = {
  readonly label: 'Add item here';
};

export const inventoryMapGestureConfig = {
  branchSwipeActivationDistance: -2,
  branchSwipeAxisDominanceRatio: 1.05,
  branchSwipeRevealWidth: 58,
  branchSwipeSelectDistance: -58,
  mapPanActivationDistance: 8,
  mapPanAxisDominanceRatio: 1.2,
  mapPanVelocityProjection: 0.18
} as const;

export function buildBrowseSurfaceOptions(): readonly {
  readonly label: string;
  readonly value: InventoryMapSurface;
}[] {
  return [
    { label: 'List', value: 'list' },
    { label: 'Map', value: 'map' }
  ];
}

export function buildInventoryMapColumns(
  map: InventoryMapViewModel,
  openPath: readonly string[]
): readonly InventoryMapColumnViewModel[] {
  const childrenByParent = groupAssetsByParent(map.assets);
  const assetsById = new Map(map.assets.map((asset) => [asset.id, asset]));
  const safePath = sanitizeOpenPath(openPath, assetsById);
  const columns: InventoryMapColumnViewModel[] = [{
    key: 'root',
    level: 0,
    title: map.inventoryName,
    assets: childrenByParent.get(rootKey) ?? [],
    emptyLabel: 'No active assets yet'
  }];

  safePath.forEach((assetIdValue, index) => {
    const asset = assetsById.get(assetIdValue);
    if (!asset || !asset.canContainAssets) {
      return;
    }

    columns.push({
      key: asset.id,
      level: index + 1,
      title: asset.title,
      parentId: asset.id,
      assets: childrenByParent.get(asset.id) ?? [],
      emptyLabel: `${asset.title} is empty`
    });
  });

  return columns;
}

export function buildInventoryMapBreadcrumbs(
  map: InventoryMapViewModel,
  openPath: readonly string[]
): readonly InventoryMapBreadcrumbViewModel[] {
  const assetsById = new Map(map.assets.map((asset) => [asset.id, asset]));
  const safePath = sanitizeOpenPath(openPath, assetsById);

  return [
    { key: 'root', level: 0, title: map.inventoryName },
    ...safePath.map((assetIdValue, index) => ({
      key: assetIdValue,
      level: index + 1,
      title: assetsById.get(assetIdValue)?.title ?? 'Unknown',
      assetId: assetIdValue
    }))
  ];
}

export function selectInventoryMapBranch(
  map: InventoryMapViewModel,
  openPath: readonly string[],
  assetIdValue: string
): readonly string[] {
  const assetsById = new Map(map.assets.map((asset) => [asset.id, asset]));
  const selectedAsset = assetsById.get(assetIdValue);

  if (!selectedAsset || !selectedAsset.canContainAssets) {
    return openPath;
  }

  const selectedPath = pathToAsset(assetIdValue, assetsById);
  return selectedPath ?? openPath;
}

export function pathForBreadcrumbLevel(
  openPath: readonly string[],
  level: number
): readonly string[] {
  return openPath.slice(0, Math.max(0, level));
}

export function preserveInventoryMapHighlightForPath(
  openPath: readonly string[],
  highlightedAssetId: string | undefined
): string | undefined {
  if (!highlightedAssetId) {
    return undefined;
  }

  return openPath.includes(highlightedAssetId) ? highlightedAssetId : undefined;
}

export function findInventoryMapSearchMatch(
  map: InventoryMapViewModel,
  query: string
): InventoryMapSearchMatch | undefined {
  const trimmed = query.trim().toLowerCase();
  if (trimmed.length === 0) {
    return undefined;
  }

  const assetsById = new Map(map.assets.map((asset) => [asset.id, asset]));
  const match = map.assets.find((asset) =>
    asset.title.toLowerCase().includes(trimmed)
    || asset.kindLabel.toLowerCase().includes(trimmed)
    || asset.placementLabel.toLowerCase().includes(trimmed)
  );

  if (!match) {
    return undefined;
  }

  const path = pathToAsset(match.id, assetsById);
  if (!path) {
    return undefined;
  }

  return {
    assetId: match.id,
    openPath: match.canContainAssets ? path : path.slice(0, -1)
  };
}

export function mapOverviewLabel(map: InventoryMapViewModel): string {
  const rootCount = map.assets.filter((asset) => asset.parentAssetId === undefined).length;
  const assetLabel = map.assets.length === 1 ? '1 active asset' : `${map.assets.length.toString()} active assets`;
  const rootLabel = rootCount === 1 ? '1 root item' : `${rootCount.toString()} root items`;

  return `${assetLabel} · ${rootLabel}`;
}

export function buildInventoryMapRowInteractionState(
  openPath: readonly string[],
  assetIdValue: string,
  highlightedAssetId: string | undefined
): InventoryMapRowInteractionState {
  const expanded = openPath.includes(assetIdValue);
  return {
    expanded,
    highlighted: expanded || highlightedAssetId === assetIdValue
  };
}

export function buildInventoryMapEmptyColumnAction(
  parentAsset: InventoryMapAssetViewModel | undefined
): InventoryMapEmptyColumnActionViewModel | undefined {
  if (!parentAsset?.canAddContainedAssets) {
    return undefined;
  }

  return { label: 'Add item here' };
}

export function shouldActivateInventoryMapBranchSwipe(input: {
  readonly canContainAssets: boolean;
  readonly dx: number;
  readonly dy: number;
}): boolean {
  return (
    input.canContainAssets
    && input.dx < inventoryMapGestureConfig.branchSwipeActivationDistance
    && Math.abs(input.dx) > Math.abs(input.dy) * inventoryMapGestureConfig.branchSwipeAxisDominanceRatio
  );
}

export function shouldSuppressInventoryMapScrollForBranchSwipe(input: {
  readonly canContainAssets: boolean;
  readonly dx: number;
  readonly dy: number;
}): boolean {
  return shouldActivateInventoryMapBranchSwipe(input);
}

export function shouldSelectInventoryMapBranchDuringSwipe(input: {
  readonly dx: number;
}): boolean {
  return input.dx <= inventoryMapGestureConfig.branchSwipeSelectDistance;
}

export function clampInventoryMapOffset(input: {
  readonly offset: number;
  readonly maxOffset: number;
}): number {
  return Math.min(Math.max(0, input.offset), Math.max(0, input.maxOffset));
}

export function nearestInventoryMapColumnForOffset(input: {
  readonly offset: number;
  readonly snapInterval: number;
  readonly maxLevel: number;
}): number {
  if (input.snapInterval <= 0) {
    return 0;
  }

  return Math.min(
    Math.max(0, Math.round(input.offset / input.snapInterval)),
    Math.max(0, input.maxLevel)
  );
}

export function inventoryMapBranchSwipeOffset(input: {
  readonly dragX: number;
  readonly fromLevel: number;
  readonly snapInterval: number;
  readonly maxLevel: number;
}): number {
  const selectDistance = Math.abs(inventoryMapGestureConfig.branchSwipeSelectDistance);
  const extraPull = Math.max(0, -input.dragX - selectDistance);
  const progress = Math.min(1, extraPull / Math.max(1, input.snapInterval - selectDistance));
  return clampInventoryMapOffset({
    offset: (input.fromLevel + progress) * input.snapInterval,
    maxOffset: input.maxLevel * input.snapInterval
  });
}

export function shouldActivateInventoryMapPagerPan(input: {
  readonly dx: number;
  readonly dy: number;
}): boolean {
  return (
    Math.abs(input.dx) > inventoryMapGestureConfig.mapPanActivationDistance
    && Math.abs(input.dx) > Math.abs(input.dy) * inventoryMapGestureConfig.mapPanAxisDominanceRatio
  );
}

function groupAssetsByParent(
  assets: readonly InventoryMapAssetViewModel[]
): Map<string, readonly InventoryMapAssetViewModel[]> {
  const groups = new Map<string, InventoryMapAssetViewModel[]>();

  for (const asset of assets) {
    const key = asset.parentAssetId ?? rootKey;
    const group = groups.get(key) ?? [];
    group.push(asset);
    groups.set(key, group);
  }

  return groups;
}

function sanitizeOpenPath(
  openPath: readonly string[],
  assetsById: ReadonlyMap<string, InventoryMapAssetViewModel>
): readonly string[] {
  const safePath: string[] = [];
  let expectedParentId: string | undefined;

  for (const assetIdValue of openPath) {
    const asset = assetsById.get(assetIdValue);
    if (!asset || asset.parentAssetId !== expectedParentId) {
      break;
    }
    safePath.push(asset.id);
    expectedParentId = asset.id;
  }

  return safePath;
}

function pathToAsset(
  assetIdValue: string,
  assetsById: ReadonlyMap<string, InventoryMapAssetViewModel>
): readonly string[] | undefined {
  const path: string[] = [];
  const seen = new Set<string>();
  let cursor = assetsById.get(assetIdValue);

  while (cursor) {
    if (seen.has(cursor.id)) {
      return undefined;
    }
    seen.add(cursor.id);
    path.unshift(cursor.id);
    cursor = cursor.parentAssetId ? assetsById.get(cursor.parentAssetId) : undefined;
  }

  return path.length > 0 ? path : undefined;
}

const rootKey = '__inventory_root__';
