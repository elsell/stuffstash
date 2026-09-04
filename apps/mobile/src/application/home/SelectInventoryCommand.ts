import { inventoryId } from '../../domain/inventories/InventorySummary';
import type { InventorySummaryRepository } from './InventorySummaryRepository';

export type SelectInventoryCommandResult = {
  readonly selectedInventoryId: string;
};

export type InventorySelectionObserver = {
  onInventorySelected(): Promise<void>;
};

const noOpInventorySelectionObserver: InventorySelectionObserver = {
  async onInventorySelected() {}
};

export class SelectInventoryCommand {
  constructor(
    private readonly inventories: InventorySummaryRepository,
    private readonly observer: InventorySelectionObserver = noOpInventorySelectionObserver
  ) {}

  async execute(inventoryIdValue: string): Promise<SelectInventoryCommandResult> {
    const selectedInventoryId = inventoryId(inventoryIdValue);
    await this.inventories.selectInventory(selectedInventoryId);
    await this.observer.onInventorySelected();

    return { selectedInventoryId };
  }
}
