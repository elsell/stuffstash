import type { QueryClient } from '@tanstack/react-query';
import type { CustomizationMutation, CustomizationMutationObserver } from '../../application/customization/CustomizationMutationObserver';
import { mobileQueryKeys } from './MobileQueryClient';

export class QueryClientCustomizationMutationObserver implements CustomizationMutationObserver {
  constructor(private readonly client: QueryClient, private readonly scopeId: string) {}
  onCustomizationChanged(mutation: CustomizationMutation): void {
    const prefix = mutation.inventoryId ? mobileQueryKeys.inventory(this.scopeId, mutation.tenantId, mutation.inventoryId) : mobileQueryKeys.tenant(this.scopeId, mutation.tenantId);
    void this.client.invalidateQueries({ predicate: ({ queryKey }) => {
      if (!prefix.every((part, index) => queryKey[index] === part)) return false;
      const offset = queryKey[4] === 'inventory' ? 6 : 4;
      const resource = queryKey[offset];
      if (resource === 'customization') return queryKey[offset + 1] === mutation.kind || mutation.kind === 'asset-type';
      if (resource === 'asset') return ['core', 'contents'].includes(String(queryKey[offset + 2]));
      return ['asset-tags', 'add-context', 'home', 'assets', 'locations', 'location', 'browse', 'map', 'parent-candidates'].includes(String(resource));
    } });
  }
}
