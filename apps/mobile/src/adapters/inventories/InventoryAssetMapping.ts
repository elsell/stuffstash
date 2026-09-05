import type {
  Asset,
  AssetTag,
  Inventory,
  Tenant
} from '@stuff-stash/api-client';
import {
  AssetBrowsePageInput
} from '../../application/home/InventorySummaryRepository';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  AccessRole,
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';

type AssetDetailWorkspaceSelection = {
  readonly assets: readonly Asset[];
  readonly photoAssetIds: ReadonlySet<string>;
};

export function selectAssetDetailWorkspace(
  root: Asset,
  assets: readonly Asset[]
): AssetDetailWorkspaceSelection {
  const childrenByParent = new Map<string, Asset[]>();
  const indexedIds = new Set<string>();
  for (const asset of assets) {
    if (indexedIds.has(asset.id)) {
      continue;
    }
    indexedIds.add(asset.id);
    if (!asset.parentAssetId) {
      continue;
    }
    const children = childrenByParent.get(asset.parentAssetId) ?? [];
    children.push(asset);
    childrenByParent.set(asset.parentAssetId, children);
  }
  for (const children of childrenByParent.values()) {
    children.sort((left, right) => left.id.localeCompare(right.id));
  }

  if (root.kind === 'container') {
    const directChildren = childrenByParent.get(root.id) ?? [];
    const workspaceAssets = [root, ...directChildren];
    return {
      assets: workspaceAssets,
      photoAssetIds: new Set(workspaceAssets.map((asset) => asset.id))
    };
  }

  const subtree: Asset[] = [];
  const photoAssetIds = new Set<string>([root.id]);
  const visited = new Set<string>();
  const pending = [root];
  while (pending.length > 0) {
    const asset = pending.pop();
    if (!asset || visited.has(asset.id)) {
      continue;
    }
    visited.add(asset.id);
    subtree.push(asset);
    if (asset.kind === 'item' || asset.parentAssetId === root.id) {
      photoAssetIds.add(asset.id);
    }
    const children = childrenByParent.get(asset.id) ?? [];
    for (let index = children.length - 1; index >= 0; index -= 1) {
      const child = children[index];
      if (child) {
        pending.push(child);
      }
    }
  }
  return { assets: subtree, photoAssetIds };
}

export function emptyInventorySummary(tenant: Tenant, inventory: Inventory): InventorySummary {
  return {
    id: inventoryId(inventory.id),
    tenantId: tenantId(tenant.id),
    name: inventory.name,
    role: mapAccessRole(inventory.access.relationship),
    permissions: [...inventory.access.permissions],
    description: '',
    updatedAtLabel: 'Loaded from API',
    locationCount: 0,
    locations: [],
    assets: [],
    assetTags: []
  };
}

export function mapAccessRole(relationship: string): AccessRole {
  switch (relationship) {
    case 'owner':
    case 'editor':
    case 'viewer':
      return relationship;
    default:
      return 'viewer';
  }
}

export function mapTenant(tenant: Tenant) {
  return {
    id: tenantId(tenant.id),
    name: tenant.name
  };
}

export function mapAssetTag(tag: AssetTag) {
  return {
    id: tag.id,
    key: tag.key,
    displayName: tag.displayName,
    color: tag.color
  };
}

export function mapLocation(
  location: Asset,
  assets: readonly Asset[],
  photo?: AssetSummary['photo']
): LocationSummary {
  const children = assets.filter((asset) => asset.parentAssetId === location.id);
  const recentChildren = sortAssetsByUpdatedDesc(children).slice(0, 3);

  return {
    id: assetId(location.id),
    inventoryId: inventoryId(location.inventoryId),
    title: location.title,
    description: location.description || 'Location asset',
    containedAssetCount: children.length,
    recentAssetTitles: recentChildren.map((asset) => asset.title),
    hasPhoto: photo !== undefined,
    photo
  };
}

export function filterAssetsByKind(
  assets: readonly Asset[],
  kind: AssetBrowsePageInput['kind']
): readonly Asset[] {
  if (kind === 'all') {
    return assets;
  }

  return assets.filter((asset) => asset.kind === kind);
}

export function assetMatchesKind(asset: Asset, kind: AssetBrowsePageInput['kind']): boolean {
  return kind === 'all' || asset.kind === kind;
}

export function filterAssetsByCheckoutState(
  assets: readonly Asset[],
  checkoutState: AssetBrowsePageInput['checkoutState']
): readonly Asset[] {
  if (checkoutState === 'checked_out') {
    return assets.filter((asset) => asset.currentCheckout !== undefined);
  }
  if (checkoutState === 'available') {
    return assets.filter((asset) => asset.currentCheckout === undefined);
  }
  return assets;
}

export function sortAssetsByUpdatedDesc(assets: readonly Asset[]): readonly Asset[] {
  return [...assets].sort((left, right) => {
    const rightTime = Date.parse(right.updatedAt || right.createdAt || '');
    const leftTime = Date.parse(left.updatedAt || left.createdAt || '');
    const timeComparison = safeTimestamp(rightTime) - safeTimestamp(leftTime);

    if (timeComparison !== 0) {
      return timeComparison;
    }

    return right.id.localeCompare(left.id);
  });
}

export function safeTimestamp(timestamp: number): number {
  return Number.isNaN(timestamp) ? 0 : timestamp;
}

export function mapAsset(
  inventoryName: string,
  asset: Asset,
  assets: readonly Asset[],
  photos: readonly NonNullable<AssetSummary['photo']>[] = [],
  options: MapAssetOptions = {}
): AssetSummary {
  const knownAsset = assets.find((candidate) => candidate.id === asset.id);
  const shouldResolveMissingParent =
    options.resolveMissingParentFromKnownAsset &&
    (asset.parentAssetId === undefined || asset.parentAssetId === null) &&
    knownAsset?.parentAssetId;
  const parentAssetID = shouldResolveMissingParent
    ? knownAsset.parentAssetId
    : asset.parentAssetId === undefined
      ? knownAsset?.parentAssetId
      : asset.parentAssetId;
  const parent = parentAssetID
    ? assets.find((candidate) => candidate.id === parentAssetID)
    : undefined;
  const assetWithResolvedParent = parentAssetID === asset.parentAssetId
    ? asset
    : { ...asset, parentAssetId: parentAssetID ?? null };
  const ancestors = ancestorTrail(assetWithResolvedParent, assets);
  const photo = photos[0];

  return {
    id: assetId(asset.id),
    title: asset.title,
    kind: asset.kind,
    lifecycleState: asset.lifecycleState,
    parentAssetId: parentAssetID ? assetId(parentAssetID) : undefined,
    locationLabel: parent?.title ?? 'Inventory root',
    locationTrail: [inventoryName, ...ancestors.map((ancestor) => ancestor.title), asset.title].filter(isString),
    parentLocationTrail: ancestors.map((ancestor) => ({
      id: assetId(ancestor.id),
      title: ancestor.title
    })),
    description: asset.description,
    updatedAtLabel: updatedAtLabel(asset),
    hasPhoto: photo !== undefined,
    photos,
    photo,
    currentCheckout: asset.currentCheckout,
    tags: asset.tags,
    undoableOperationId: asset.undoableOperationId
  };
}

export type MapAssetOptions = {
  readonly resolveMissingParentFromKnownAsset?: boolean;
};

export function placementAssetsFromFullTree(
  ancestryAssets: readonly Asset[],
  selectedAssets: readonly Asset[]
): readonly Asset[] {
  const merged = new Map<string, Asset>();
  for (const asset of ancestryAssets) {
    merged.set(asset.id, asset);
  }
  for (const asset of selectedAssets) {
    if (!merged.has(asset.id)) {
      merged.set(asset.id, asset);
    }
  }
  return [...merged.values()];
}

export function placementAssetsWithSelectedOverrides(
  ancestryAssets: readonly Asset[],
  selectedAssets: readonly Asset[]
): readonly Asset[] {
  const merged = new Map<string, Asset>();
  for (const asset of ancestryAssets) {
    merged.set(asset.id, asset);
  }
  for (const asset of selectedAssets) {
    merged.set(asset.id, asset);
  }
  return [...merged.values()];
}

export function searchMatchLabels(matches: readonly { readonly field: string }[]): readonly string[] {
  const labels: string[] = [];
  const seen = new Set<string>();
  for (const match of matches) {
    const label = searchMatchFieldLabel(match.field);
    if (seen.has(label)) {
      continue;
    }
    labels.push(label);
    seen.add(label);
  }
  return labels;
}

export function searchMatchFieldLabel(field: string): string {
  switch (field) {
    case 'tag_display_name':
    case 'tag_key':
      return 'Tag';
    case 'title':
      return 'Title';
    case 'description':
      return 'Description';
    case 'location':
    case 'path':
      return 'Location';
    case 'custom_field':
      return 'Custom field';
    default:
      return humanizeSearchMatchField(field);
  }
}

export function humanizeSearchMatchField(field: string): string {
  const label = field.trim().replace(/[_-]+/g, ' ');
  if (label.length === 0) {
    return 'Match';
  }
  return label.charAt(0).toUpperCase() + label.slice(1);
}

export function ancestorTrail(asset: Asset, assets: readonly Asset[]): readonly Asset[] {
  const byID = new Map(assets.map((candidate) => [candidate.id, candidate]));
  const ancestors: Asset[] = [];
  const seen = new Set<string>([asset.id]);
  let parentID = asset.parentAssetId ?? undefined;

  while (parentID && !seen.has(parentID)) {
    seen.add(parentID);
    const parent = byID.get(parentID);
    if (!parent) {
      break;
    }
    ancestors.unshift(parent);
    parentID = parent.parentAssetId ?? undefined;
  }

  return ancestors;
}

export function isString(value: string | undefined): value is string {
  return typeof value === 'string' && value.length > 0;
}

export function updatedAtLabel(asset: Asset): string {
  const timestamp = asset.updatedAt || asset.createdAt;
  if (!timestamp) {
    return 'Loaded from API';
  }
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) {
    return 'Loaded from API';
  }
  return `Updated ${date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })}`;
}

export function summaryToApiAsset(
  tenantID: string,
  inventoryID: string,
  asset: AssetSummary
): Asset {
  return {
    id: asset.id,
    tenantId: tenantID,
    inventoryId: inventoryID,
    kind: asset.kind,
    title: asset.title,
    description: asset.description,
    parentAssetId: asset.parentAssetId ?? null,
    lifecycleState: asset.lifecycleState,
    tags: [...(asset.tags ?? [])],
    customFields: {},
    createdAt: '',
    updatedAt: '',
    currentCheckout: asset.currentCheckout
  };
}

export async function mapWithConcurrency<Input, Output>(
  items: readonly Input[],
  concurrency: number,
  mapper: (item: Input) => Promise<Output>
): Promise<readonly Output[]> {
  const results = new Array<Output>(items.length);
  let nextIndex = 0;
  const workerCount = Math.min(concurrency, items.length);

  await Promise.all(Array.from({ length: workerCount }, async () => {
    while (nextIndex < items.length) {
      const index = nextIndex;
      nextIndex += 1;
      results[index] = await mapper(items[index]);
    }
  }));

  return results;
}
