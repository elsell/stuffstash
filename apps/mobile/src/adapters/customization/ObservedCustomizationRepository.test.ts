import { expect, it } from 'vitest';
import type { StuffStashClient } from '@stuff-stash/api-client';
import { ApiCustomizationRepository } from './ApiCustomizationRepository';
import { ObservedCustomizationRepository } from './ObservedCustomizationRepository';
import type { CustomizationMutation } from '../../application/customization/CustomizationMutationObserver';

it('reports successful customization changes and preserves cache freshness after rejection', async () => {
  const changes: CustomizationMutation[] = [];
  const api = new ApiCustomizationRepository({ createAssetTag: async () => ({ id: 'tag', key: 'tag', displayName: 'Tag' }), archiveAssetTag: async () => { throw new Error('unavailable'); } } as unknown as StuffStashClient);
  const repository = new ObservedCustomizationRepository(api, { onCustomizationChanged: (change) => changes.push(change) });
  const context = { tenantId: 'tenant', tenantName: 'Home', inventoryId: 'inventory', inventoryName: 'Home', tenantPermissions: [], inventoryPermissions: [] };
  await repository.createTag(context, { displayName: 'Tag' });
  await expect(repository.archiveTag(context, 'tag')).rejects.toThrow();
  expect(changes).toEqual([{ tenantId: 'tenant', inventoryId: 'inventory', kind: 'tag' }]);
});
