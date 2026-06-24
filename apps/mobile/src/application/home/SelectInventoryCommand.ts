import { inventoryId } from '../../domain/inventories/InventorySummary';
import type { InventorySummaryRepository } from './InventorySummaryRepository';

export type SelectInventoryCommandResult = {
  readonly selectedInventoryId: string;
};

export class SelectInventoryCommand {
  constructor(private readonly inventories: InventorySummaryRepository) {}

  async execute(inventoryIdValue: string): Promise<SelectInventoryCommandResult> {
    const selectedInventoryId = inventoryId(inventoryIdValue);
    await this.inventories.selectInventory(selectedInventoryId);

    return { selectedInventoryId };
  }
}
