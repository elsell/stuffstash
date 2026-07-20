import { describe, expect, it } from 'vitest';
import { StuffStashAPIError, type StuffStashClient } from '@stuff-stash/api-client';
import { CustomizationFailure } from '../../application/customization/CustomizationErrors';
import type { CustomizationContext } from '../../application/customization/CustomizationRepository';
import { ApiCustomizationRepository } from './ApiCustomizationRepository';

const context: CustomizationContext = {
  tenantId: 'tenant-one', tenantName: 'Home', tenantPermissions: ['configure'],
  inventoryId: 'inventory-one', inventoryName: 'Household', inventoryPermissions: ['view', 'configure', 'edit_asset']
};

describe('ApiCustomizationRepository', () => {
  it.each([
    [403, 'forbidden', 'permission-denied'],
    [404, 'not_found', 'not-found'],
    [409, 'resource_in_use', 'conflict'],
    [422, 'invalid_request', 'invalid'],
    [503, 'unavailable', 'unavailable']
  ] as const)('maps API status %s to a sanitized typed failure', async (status, code, kind) => {
    const repository = new ApiCustomizationRepository({
      listAssetTags: async () => { throw new StuffStashAPIError(status, code, 'unsafe token and request id'); }
    } as unknown as StuffStashClient);
    const failure = await repository.listTags(context).catch((error: unknown) => error);
    expect(failure).toBeInstanceOf(CustomizationFailure);
    expect((failure as CustomizationFailure).kind).toBe(kind);
    expect((failure as Error).message).not.toContain('unsafe');
  });

  it('maps lifecycle fields and keeps tenant and inventory operations correctly scoped', async () => {
    const calls: string[][] = [];
    const repository = new ApiCustomizationRepository({
      listInventoryCustomFieldDefinitions: async () => ({ items: [fieldResponse('tenant', null, 'active')], pagination: {} }),
      restoreTenantCustomFieldDefinition: async (...args: string[]) => { calls.push(args); return fieldResponse('tenant', null, 'archived'); },
      deleteInventoryCustomFieldDefinition: async (...args: string[]) => { calls.push(args); }
    } as unknown as StuffStashClient);

    await expect(repository.listFields(context, 'inventory', 'active')).resolves.toMatchObject({ items: [{ kind: 'field', scope: 'tenant', lifecycle: 'active' }] });
    await repository.restoreField({ scope: 'tenant', tenantId: 'tenant-one', id: 'field-one' });
    await repository.deleteField({ scope: 'inventory', tenantId: 'tenant-one', inventoryId: 'inventory-one', id: 'field-one' });
    expect(calls).toEqual([['tenant-one', 'field-one'], ['tenant-one', 'inventory-one', 'field-one']]);
  });
});

function fieldResponse(scope: 'tenant' | 'inventory', inventoryId: string | null, lifecycleState: 'active' | 'archived') {
  return { id: 'field-one', tenantId: 'tenant-one', inventoryId, scope, key: 'priority', displayName: 'Priority', type: 'text' as const, enumOptions: [], applicability: 'all_assets' as const, customAssetTypeIds: [], lifecycleState };
}
