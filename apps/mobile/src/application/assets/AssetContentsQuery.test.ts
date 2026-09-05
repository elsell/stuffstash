import { describe, expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import type { AssetCoreSnapshot } from './AssetCoreQuery';
import { AssetContentsQuery } from './AssetContentsQuery';

describe('AssetContentsQuery', () => {
  it('maps placement and children independently and forwards cancellation', async () => {
    const core = coreSnapshot();
    const controller = new AbortController();
    let receivedSignal: AbortSignal | undefined;
    const query = new AssetContentsQuery({
      getAssetContents: async (_core, request) => {
        receivedSignal = request?.signal;
        return {
          asset: { ...core.asset, parentLocationTrail: [{ id: assetId('garage'), title: 'Garage' }] },
          allAssets: [{
            ...core.asset,
            id: assetId('child'),
            title: 'Child',
            parentAssetId: core.asset.id
          }]
        };
      }
    });

    await expect(query.execute(core, { signal: controller.signal })).resolves.toMatchObject({
      parentLocationTrailLabel: 'Garage',
      containedAssets: [{ id: 'child' }]
    });
    expect(receivedSignal).toBe(controller.signal);
  });
});

function coreSnapshot(): AssetCoreSnapshot {
  return {
    tenantId: tenantId('tenant-home'),
    inventoryId: inventoryId('inventory-home'),
    permissions: ['view', 'edit_asset'],
    revision: '2026-09-04T12:00:00Z',
    asset: {
      id: assetId('container'),
      title: 'Container',
      kind: 'container',
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Container'],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    }
  };
}
