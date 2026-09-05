import { SettingsScopeUnavailableError } from '../../application/settings/SettingsQuery';
import type { ReadRequest } from '../../application/shared/ReadRequest';
import type { StuffStashClient } from '@stuff-stash/api-client';
import type {
  SettingsInventoryScope,
  SettingsScopeRepository,
  SettingsTenantScope
} from '../../application/settings/SettingsQuery';

type SettingsScopeApiClient = Pick<StuffStashClient, 'getTenant' | 'listInventories'>;

type CurrentTenantScope = {
  getCurrentSettingsScope(request?: ReadRequest): Promise<{
    readonly tenantId: string;
    readonly inventory: SettingsInventoryScope;
  }>;
};

export class ApiSettingsScopeRepository implements SettingsScopeRepository {
  constructor(
    private readonly client: SettingsScopeApiClient,
    private readonly currentTenant: CurrentTenantScope
  ) {}

  async getSelectedScope(request: ReadRequest = {}): Promise<{ readonly tenant: SettingsTenantScope; readonly inventory: SettingsInventoryScope }> {
    const scope = await this.currentTenant.getCurrentSettingsScope(request);
    const [tenant, inventory] = await Promise.all([
      this.client.getTenant(scope.tenantId, request.signal),
      this.selectedInventory(scope.tenantId, scope.inventory.id, request)
    ]);

    return {
      tenant: { id: tenant.id, name: tenant.name, permissions: [...tenant.access.permissions] },
      inventory: { id: inventory.id, name: inventory.name, permissions: [...inventory.access.permissions] }
    };
  }
  private async selectedInventory(tenantId: string, inventoryId: string, request: ReadRequest) {
    let cursor: string | undefined;
    const visited = new Set<string>();
    do {
      request.signal?.throwIfAborted();
      const page = await this.client.listInventories(tenantId, 100, cursor, request.signal);
      const inventory = page.items.find((candidate) => candidate.id === inventoryId);
      if (inventory) return inventory;
      cursor = page.pagination.hasMore ? page.pagination.nextCursor ?? undefined : undefined;
      if (cursor && visited.has(cursor)) break;
      if (cursor) visited.add(cursor);
    } while (cursor);
    throw new SettingsScopeUnavailableError();
  }

}
