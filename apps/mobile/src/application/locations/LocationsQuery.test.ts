import { describe, expect, it } from 'vitest';
import {
  inventoryId,
  tenantId,
  type InventorySummary
} from '../../domain/inventories/InventorySummary';
import type { InventoryWorkspace } from '../home/InventorySummaryRepository';
import { LocationsQuery } from './LocationsQuery';

class FakeInventoryWorkspaceRepository {
  constructor(private readonly inventory: InventorySummary) {}

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return {
      defaultInventoryId: this.inventory.id,
      inventories: [this.inventory],
      tenants: [{ id: this.inventory.tenantId, name: 'Household' }]
    };
  }
}

describe('LocationsQuery', () => {
  it.each([
    { permissions: ['view', 'create_asset'] as const, canAdd: true },
    { permissions: ['view'] as const, canAdd: false }
  ])('maps create permission to canAdd=$canAdd', async ({ permissions, canAdd }) => {
    const query = new LocationsQuery(
      new FakeInventoryWorkspaceRepository(inventory(permissions))
    );

    await expect(query.execute()).resolves.toMatchObject({ canAdd });
  });
});

function inventory(permissions: InventorySummary['permissions']): InventorySummary {
  return {
    id: inventoryId('inventory-home'),
    tenantId: tenantId('tenant-household'),
    name: 'Home',
    role: permissions.includes('create_asset') ? 'editor' : 'viewer',
    permissions,
    description: 'Household inventory.',
    updatedAtLabel: 'Updated today',
    locationCount: 0,
    locations: [],
    assets: []
  };
}
