import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';
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
    private readonly inventories: Pick<InventorySummaryRepository, 'selectInventory'>,
    private readonly observer: InventorySelectionObserver = noOpInventorySelectionObserver
  ) {}

  async execute(inventoryIdValue: string, request: ReadRequest = {}): Promise<SelectInventoryCommandResult> {
    assertReadActive(request.signal);
    const selectedInventoryId = inventoryId(inventoryIdValue);
    await this.inventories.selectInventory(selectedInventoryId, request);
    assertReadActive(request.signal);
    await this.observer.onInventorySelected();
    assertReadActive(request.signal);

    return { selectedInventoryId };
  }
}
