import type { StuffStashClient } from '@stuff-stash/api-client';
import type {
  SettingsInventoryScope,
  SettingsScopeRepository,
  SettingsTenantScope
} from '../../application/settings/SettingsQuery';

type SettingsScopeApiClient = Pick<StuffStashClient, 'listMyTenants'>;

type CurrentTenantScope = {
  getCurrentSettingsScope(): Promise<{
    readonly tenantId: string;
    readonly inventory: SettingsInventoryScope;
  }>;
};

export class ApiSettingsScopeRepository implements SettingsScopeRepository {
  constructor(
    private readonly client: SettingsScopeApiClient,
    private readonly currentTenant: CurrentTenantScope
  ) {}

  async getSelectedScope(): Promise<{ readonly tenant: SettingsTenantScope; readonly inventory: SettingsInventoryScope }> {
    const [scope, page] = await Promise.all([
      this.currentTenant.getCurrentSettingsScope(),
      this.client.listMyTenants(100)
    ]);
    const tenant = page.items.find((item) => item.id === scope.tenantId);
    if (!tenant) {
      throw new Error('The selected Stuff Stash tenant is no longer available.');
    }
    if (scope.inventory.id.length === 0) {
      throw new Error('The selected Stuff Stash inventory is no longer available.');
    }
    return {
      tenant: { id: tenant.id, name: tenant.name, permissions: [...tenant.access.permissions] },
      inventory: { ...scope.inventory, permissions: [...scope.inventory.permissions] }
    };
  }
}
