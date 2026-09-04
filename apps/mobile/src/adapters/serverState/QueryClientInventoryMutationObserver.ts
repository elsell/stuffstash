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
    void this.client.invalidateQueries({
      predicate: (query) => isAffectedInventoryQuery(query.queryKey, inventoryKey, mutation.kind)
    });
  }
}

function isAffectedInventoryQuery(
  queryKey: QueryKey,
  inventoryKey: QueryKey,
  kind: InventoryMutation['kind']
): boolean {
  if (!hasPrefix(queryKey, inventoryKey)) {
    return false;
  }
  const resource = queryKey[inventoryKey.length];
  if (kind === 'asset_tag_created') {
    return resource === 'asset-tags' || resource === 'add-context' || resource === 'browse';
  }
  if (kind === 'asset_checkout_changed') {
    return resource === 'home'
      || resource === 'assets'
      || resource === 'location'
      || resource === 'browse'
      || resource === 'asset';
  }
  return resource === 'home'
    || resource === 'assets'
    || resource === 'locations'
    || resource === 'location'
    || resource === 'browse'
    || resource === 'asset';
}

function hasPrefix(value: QueryKey, prefix: QueryKey): boolean {
  return prefix.every((part, index) => value[index] === part);
}
