export type CurrentInventoryScope = {
  readonly tenantId: string;
  readonly inventoryId: string;
};

export type CurrentInventoryScopeRepository = {
  getCurrentInventoryScope(request?: ReadRequest): Promise<CurrentInventoryScope>;
};

export class CurrentInventoryScopeQuery {
  constructor(private readonly inventories: CurrentInventoryScopeRepository) {}

  execute(request: ReadRequest = {}): Promise<CurrentInventoryScope> {
    return this.inventories.getCurrentInventoryScope(request);
  }
}
import type { ReadRequest } from '../shared/ReadRequest';
