import React from 'react';
import { describe, expect, it } from 'vitest';
import { InventoryMapScreen } from './InventoryMapScreen';
import { dispatchedActions } from '../../test-support/navigation';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { InventoryMapQuery } from '../../application/assets/InventoryMapQuery';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
const selectedAsset: AssetSummary = { id: assetId('tent'), title: 'Tent', kind: 'item', lifecycleState: 'active', description: '', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
const mapSnapshot = { sessionScopeId: 'scope', tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), inventoryName: 'Home', permissions: ['view', 'edit_asset'], assets: [selectedAsset] };

describe('Map server state', () => {
  it('performs one initial traversal and reuses fresh data on remount', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let calls = 0;
    const query = new InventoryMapQuery({ listActiveInventoryMapAssets: async () => {
      calls++;
      return mapSnapshot;
    } });
    const props: React.ComponentProps<typeof InventoryMapScreen> = {
      canAdd: false, inventoryMapQuery: query, pathStore: { current: new Map() }, selectedSurface: 'map', onAdd: () => undefined, onChangeSurface: () => undefined
    };
    const render = (visible: boolean) => harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider>{visible ? <InventoryMapScreen {...props} /> : null}</AppFeedbackProvider>
    </MobileServerStateProvider>);
    try {
      await render(true); await settle(harness); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(calls).toBe(1);
      await render(false); await render(true); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(calls).toBe(1);
      await harness.press(harness.byLabel('Show details for Tent'));
      expect(dispatchedActions().at(-1)).toMatchObject({
        type: 'push',
        href: { pathname: '/assets/[assetId]', params: { assetId: 'tent' } }
      });
    } finally { await harness.unmount(); }
  });

});
