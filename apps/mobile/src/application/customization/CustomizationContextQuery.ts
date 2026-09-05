import type { ReadRequest } from '../shared/ReadRequest';
import type { SettingsScopeRepository } from '../settings/SettingsQuery';
import type { CustomizationContext } from './CustomizationRepository';

export class CustomizationContextQuery {
  constructor(private readonly settings: SettingsScopeRepository) {}

  async execute(request: ReadRequest = {}): Promise<CustomizationContext> {
    const scope = await this.settings.getSelectedScope(request);
    return {
      tenantId: scope.tenant.id,
      tenantName: scope.tenant.name,
      tenantPermissions: scope.tenant.permissions,
      inventoryId: scope.inventory.id,
      inventoryName: scope.inventory.name,
      inventoryPermissions: scope.inventory.permissions
    };
  }
}
