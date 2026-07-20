import type { SettingsQuery } from '../settings/SettingsQuery';
import type { CustomizationContext } from './CustomizationRepository';

export class CustomizationContextQuery {
  constructor(private readonly settings: SettingsQuery) {}

  async execute(): Promise<CustomizationContext> {
    const scope = await this.settings.execute();
    return {
      tenantId: scope.selectedTenant.id,
      tenantName: scope.selectedTenant.name,
      tenantPermissions: scope.selectedTenant.permissions,
      inventoryId: scope.selectedInventory.id,
      inventoryName: scope.selectedInventory.name,
      inventoryPermissions: scope.selectedInventory.permissions
    };
  }
}
