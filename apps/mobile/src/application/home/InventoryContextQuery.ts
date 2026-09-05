import type { ReadRequest } from '../shared/ReadRequest';

export type InventoryContextRepository = {
  getCurrentSettingsScope(request?: ReadRequest): Promise<{
    readonly tenantId: string;
    readonly inventory: { readonly id: string; readonly name: string; readonly permissions: readonly string[] };
  }>;
};

export class InventoryContextQuery {
  constructor(private readonly inventories: InventoryContextRepository) {}

  async execute(request: ReadRequest = {}): Promise<{ readonly inventoryName: string; readonly canAdd: boolean }> {
    const scope = await this.inventories.getCurrentSettingsScope(request);
    return { inventoryName: scope.inventory.name, canAdd: scope.inventory.permissions.includes('create_asset') };
  }
}
