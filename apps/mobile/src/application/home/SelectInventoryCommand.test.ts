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
  it('reconciles an accepted selection even when the caller cancels before completion', async () => {
    const controller = new AbortController();
    const events: string[] = [];
    const command = new SelectInventoryCommand({ async selectInventory() {
      events.push('accepted'); controller.abort();
    } }, { async onInventorySelected() { events.push('cache-reset'); } });
    await expect(command.execute('inventory-new', { signal: controller.signal })).rejects.toThrow();
    expect(events).toEqual(['accepted', 'cache-reset']);
  });

  it('does not publish rejected or initially canceled selections', async () => {
    const events: string[] = [];
    const command = new SelectInventoryCommand({ async selectInventory() {
      events.push('attempt'); throw new Error('unavailable');
    } }, { async onInventorySelected() { events.push('cache-reset'); } });
    const controller = new AbortController(); controller.abort();
    await expect(command.execute('inventory-new', { signal: controller.signal })).rejects.toThrow();
    expect(events).toEqual([]);
    await expect(command.execute('inventory-new')).rejects.toThrow('unavailable');
    expect(events).toEqual(['attempt']);
  });

});
