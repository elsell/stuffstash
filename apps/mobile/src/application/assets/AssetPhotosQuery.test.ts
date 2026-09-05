import { describe, expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import type { AssetCoreSnapshot } from './AssetCoreQuery';
import { AssetPhotosQuery } from './AssetPhotosQuery';

describe('AssetPhotosQuery', () => {
  it('maps photos independently and forwards cancellation', async () => {
    const controller = new AbortController();
    let receivedSignal: AbortSignal | undefined;
    const query = new AssetPhotosQuery({
      getAssetPhotos: async (_core, request) => {
        receivedSignal = request?.signal;
        return [{ id: 'photo-a', fileName: 'front.jpg', uri: 'https://example.test/front' }];
      }
    });

    await expect(query.execute(coreSnapshot(), { signal: controller.signal })).resolves.toEqual([{
      id: 'photo-a',
      fileName: 'front.jpg',
      label: 'front.jpg',
      uri: 'https://example.test/front'
    }]);
    expect(receivedSignal).toBe(controller.signal);
  });
});

function coreSnapshot(): AssetCoreSnapshot {
  return {
    tenantId: tenantId('tenant-home'),
    inventoryId: inventoryId('inventory-home'),
    permissions: ['view'],
    revision: '2026-09-04T12:00:00Z',
    asset: {
      id: assetId('item'),
      title: 'Item',
      kind: 'item',
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Item'],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    }
  };
}
