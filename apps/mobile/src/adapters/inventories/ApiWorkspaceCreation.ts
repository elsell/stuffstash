import type { StuffStashClient } from '@stuff-stash/api-client';
import type { WorkspaceCreationPort } from '../../application/inventories/CreateWorkspace';
export class ApiWorkspaceCreation implements WorkspaceCreationPort {
  constructor(private readonly client: Pick<StuffStashClient, 'getTenant' | 'createTenant' | 'createInventory'>) {}
  async canCreateInventory(tenantId: string) {
    const household = await this.client.getTenant(tenantId);
    return household.id === tenantId && household.access.permissions.includes('create_inventory');
  }
  async createHousehold(name: string) {
    const household = await this.client.createTenant(name);
    return { id: household.id, name: household.name, canCreateInventory: household.access.permissions.includes('create_inventory') };
  }
  async createInventory(tenantId: string, name: string) {
    const inventory = await this.client.createInventory(tenantId, name);
    return { id: inventory.id, tenantId: inventory.tenantId, name: inventory.name };
  }
}
