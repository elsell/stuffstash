import type { QueryClient, QueryKey } from '@tanstack/react-query';
import type {
  InventoryMutation,
  InventoryMutationObserver
} from '../../application/home/InventoryMutationObserver';
import { mobileQueryKeys } from './MobileQueryClient';

export class QueryClientInventoryMutationObserver implements InventoryMutationObserver {
  constructor(
    private readonly client: QueryClient,
    private readonly compositionScopeId: string
  ) {}

  onInventoryMutation(mutation: InventoryMutation): void {
    const inventoryKey = mobileQueryKeys.inventory(
      this.compositionScopeId,
      mutation.tenantId,
      mutation.inventoryId
    );
    const relatedAssetIds = new Set(mutation.relatedAssetIds ?? []);
    const cachedParentId = mutation.assetId
      ? cachedAssetParentId(this.client, inventoryKey, mutation.assetId)
      : undefined;
    if (cachedParentId) relatedAssetIds.add(cachedParentId);
    const dependencies = new Set(relatedAssetIds);
    if (mutation.assetId) dependencies.add(mutation.assetId);
    // Follow old and destination ancestry, including recursive containment views.
    let previousSize = -1;
    while (previousSize !== dependencies.size) {
      previousSize = dependencies.size;
      for (const assetId of [...dependencies]) {
        const parentId = cachedAssetParentId(this.client, inventoryKey, assetId);
        if (parentId) dependencies.add(parentId);
        for (const containingId of cachedContainingAssetIds(this.client, inventoryKey, assetId)) {
          dependencies.add(containingId);
        }
      }
    }
    for (const id of dependencies) {
      if (id !== mutation.assetId) relatedAssetIds.add(id);
    }
    void this.client.invalidateQueries({
      predicate: (query) => isAffectedInventoryQuery(
        query.queryKey,
        inventoryKey,
        mutation,
        relatedAssetIds
      ) || (
        hasPrefix(query.queryKey, inventoryKey)
        && query.queryKey[inventoryKey.length] === 'asset'
        && query.queryKey[inventoryKey.length + 2] === 'contents'
        && mutation.kind !== 'asset_checkout_changed'
        && mutation.kind !== 'asset_photo_changed'
        && Boolean(mutation.assetId && referencesAncestor(query.state.data, mutation.assetId))
      )
    });
  }
}

function isAffectedInventoryQuery(
  queryKey: QueryKey,
  inventoryKey: QueryKey,
  mutation: InventoryMutation,
  relatedAssetIds: ReadonlySet<string>
): boolean {
  if (!hasPrefix(queryKey, inventoryKey)) {
    return false;
  }
  const resource = queryKey[inventoryKey.length];
  const kind = mutation.kind;
  if (kind === 'asset_tag_created') {
    return resource === 'asset-tags' || resource === 'add-context' || resource === 'browse';
  }
  if (resource === 'asset') {
    return isAffectedAssetQuery(queryKey, inventoryKey, mutation, relatedAssetIds);
  }
  if (kind === 'asset_checkout_changed') {
    return resource === 'map' || resource === 'home'
      || resource === 'assets'
      || resource === 'location'
      || resource === 'browse';
  }
  return resource === 'map' || resource === 'home'
    || resource === 'assets'
    || resource === 'locations'
    || resource === 'location'
    || resource === 'browse';
}

function isAffectedAssetQuery(
  queryKey: QueryKey,
  inventoryKey: QueryKey,
  mutation: InventoryMutation,
  relatedAssetIds: ReadonlySet<string>
): boolean {
  if (!mutation.assetId) return true;
  const queryAssetId = queryKey[inventoryKey.length + 1];
  const region = queryKey[inventoryKey.length + 2];
  if (queryAssetId === mutation.assetId) {
    if (mutation.kind === 'asset_photo_changed') {
      return region === undefined || region === 'core' || region === 'photos';
    }
    if (mutation.kind === 'asset_checkout_changed') {
      return region === undefined || region === 'core';
    }
    if (
      mutation.kind === 'asset_updated'
      || mutation.kind === 'asset_lifecycle_changed'
      || mutation.kind === 'asset_created'
    ) {
      return region === undefined || region === 'core' || region === 'contents';
    }
    return true;
  }
  return typeof queryAssetId === 'string'
    && relatedAssetIds.has(queryAssetId)
    && region === 'contents';
}

function cachedAssetParentId(
  client: QueryClient,
  inventoryKey: QueryKey,
  assetId: string
): string | undefined {
  const coreQuery = client.getQueryCache().findAll({ queryKey: inventoryKey }).find((query) => {
    const key = query.queryKey;
    return key[inventoryKey.length] === 'asset'
      && key[inventoryKey.length + 1] === assetId
      && key[inventoryKey.length + 2] === 'core';
  });
  const data = coreQuery?.state.data;
  if (typeof data !== 'object' || data === null || !('snapshot' in data)) return undefined;
  const snapshot = data.snapshot;
  if (typeof snapshot !== 'object' || snapshot === null || !('asset' in snapshot)) return undefined;
  const asset = snapshot.asset;
  if (typeof asset !== 'object' || asset === null || !('parentAssetId' in asset)) return undefined;
  return typeof asset.parentAssetId === 'string' ? asset.parentAssetId : undefined;
}

function cachedContainingAssetIds(
  client: QueryClient,
  inventoryKey: QueryKey,
  assetId: string
): readonly string[] {
  const containingIds: string[] = [];
  for (const query of client.getQueryCache().findAll({ queryKey: inventoryKey })) {
    const key = query.queryKey;
    if (key[inventoryKey.length] !== 'asset' || key[inventoryKey.length + 2] !== 'contents') {
      continue;
    }
    const data = query.state.data;
    const containingId = key[inventoryKey.length + 1];
    if (typeof containingId !== 'string') continue;
    if (containsAsset(data, assetId)) containingIds.push(containingId);
    if (containingId === assetId) {
      containingIds.push(...rowIds(data, 'parentLocationTrail'));
    }
  }
  return containingIds;
}

function rowIds(data: unknown, field: string): readonly string[] {
  if (typeof data !== 'object' || data === null) return [];
  const rows = (data as Record<string, unknown>)[field];
  if (!Array.isArray(rows)) return [];
  return rows.flatMap((row: unknown) => (
    typeof row === 'object' && row !== null && 'id' in row && typeof row.id === 'string'
      ? [row.id]
      : []
  ));
}

function containsAsset(data: unknown, assetId: string): boolean {
  return ['containedAssets', 'containedSpaces', 'containedItems']
    .some((field) => rowIds(data, field).includes(assetId));
}

function referencesAncestor(data: unknown, assetId: string): boolean {
  if (rowIds(data, 'parentLocationTrail').includes(assetId)) return true;
  if (typeof data !== 'object' || data === null) return false;
  return ['containedAssets', 'containedSpaces', 'containedItems'].some((field) => {
    const rows = (data as Record<string, unknown>)[field];
    return Array.isArray(rows) && rows.some((row: unknown) => rowIds(row, 'parentLocationTrail').includes(assetId));
  });
}

function hasPrefix(value: QueryKey, prefix: QueryKey): boolean {
  return prefix.every((part, index) => value[index] === part);
}
