import { describe, expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { AssetCoreQuery } from './AssetCoreQuery';

describe('AssetCoreQuery', () => {
  it('maps one focused asset snapshot and forwards cancellation', async () => {
    const controller = new AbortController();
    let receivedSignal: AbortSignal | undefined;
    const query = new AssetCoreQuery({
      getAssetCore: async (_assetId, request) => {
        receivedSignal = request?.signal;
        return {
          tenantId: tenantId('tenant-home'),
          inventoryId: inventoryId('inventory-home'),
          permissions: ['view', 'edit_asset'],
          revision: '2026-09-04T12:00:00Z',
          asset: {
            id: assetId('asset-filter'),
            title: 'Furnace filter',
            kind: 'item',
            lifecycleState: 'active',
            locationLabel: 'Inventory root',
            locationTrail: ['Home', 'Furnace filter'],
            parentLocationTrail: [],
            description: '20x20',
            updatedAtLabel: 'Updated now',
            hasPhoto: false
          }
        };
      }
    });

    await expect(query.execute('asset-filter', { signal: controller.signal })).resolves.toMatchObject({
      snapshot: { revision: '2026-09-04T12:00:00Z' },
      view: {
        id: 'asset-filter',
        title: 'Furnace filter',
        canEdit: true,
        photos: [],
        containedAssets: []
      }
    });
    expect(receivedSignal).toBe(controller.signal);
  });
});
