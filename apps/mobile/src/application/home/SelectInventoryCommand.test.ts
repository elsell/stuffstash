import { describe, expect, it } from 'vitest';
import { inventoryId } from '../../domain/inventories/InventorySummary';
import type { InventorySummaryRepository } from './InventorySummaryRepository';
import { SelectInventoryCommand } from './SelectInventoryCommand';

describe('SelectInventoryCommand', () => {
  it('publishes selection only after the repository accepts it', async () => {
    const events: string[] = [];
    const inventories = {
      selectInventory: async (selected: ReturnType<typeof inventoryId>) => {
        events.push(`selected:${selected}`);
      }
    };
    const command = new SelectInventoryCommand(inventories as unknown as InventorySummaryRepository, {
      onInventorySelected: async () => {
        events.push('cache-reset');
      }
    });

    await expect(command.execute('inventory-new')).resolves.toEqual({
      selectedInventoryId: 'inventory-new'
    });
    expect(events).toEqual(['selected:inventory-new', 'cache-reset']);
  });
});
