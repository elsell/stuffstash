import { describe, expect, it } from 'vitest';
import { AssetPlacementQuery } from './AssetPlacementQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';

it('loads only the selected placement through the port and forwards cancellation', async () => {
  const signal = new AbortController().signal;
  const asset = { id: assetId('asset'), title: 'Box', description: '', kind: 'container' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = { asset, tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: [], revision: '1' };
  let observed: unknown;
  const query = new AssetPlacementQuery({ getAssetPlacement: async (snapshot, request) => {
    observed = { snapshot, request };
    return { ...asset, locationTrail: ['Home', 'Room', 'Box'], parentLocationTrail: [{ id: assetId('room'), title: 'Room' }] };
  } });
  const result = await query.execute(core, { signal });
  expect(result.parentLocationTrail.map((crumb) => crumb.title)).toEqual(['Room']);
  expect(observed).toEqual({ snapshot: core, request: { signal } });
});
