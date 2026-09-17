import React from 'react';
import { expect, it } from 'vitest';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { dispatchedActions, resetNavigation, setCanGoBack } from '../../test-support/navigation';
import { latestAlert, pressAlertButton } from '../../test-support/react-native';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AssetEditSheetRouteScreen, AssetMoveSheetRouteScreen, AssetMoveHereSheetRouteScreen } from './AssetNativeActionSheetScreens';
import { consumeAssetActionCompletion } from './AssetActionCompletion';

for (const route of ['edit', 'move', 'move-here'] as const) {
  for (const action of ['cancel', 'save'] as const) {
    it.each([false, true])(`returns from loaded ${route} after ${action} (back=%s)`, async canGoBack => {
      const h = new MobileRenderHarness(); const client = createMobileQueryClient();
      resetNavigation(); setCanGoBack(canGoBack); consumeAssetActionCompletion('asset');
      const submitted: unknown[] = [];
      const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'container' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
      const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
      const candidate = { id: 'box', title: 'Camping box', kind: 'container' as const, subtitle: '', pathLabel: 'Camping box', selectionHint: 'Container', willPromoteToContainer: false };
      const command = { execute: async (input: unknown) => { submitted.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } };
      const shared = { assetId: 'asset', assetCoreQuery: core, parentLookupQuery: { execute: async () => [candidate] }, moveAssetCommand: command };
      try {
        await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
          {route === 'edit' ? <AssetEditSheetRouteScreen {...shared} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }} updateAssetCommand={command} />
            : route === 'move' ? <AssetMoveSheetRouteScreen {...shared} createAssetCommand={{ execute: async () => { throw new Error('Unexpected creation'); } }} />
              : <AssetMoveHereSheetRouteScreen {...shared} />}
        </MobileServerStateProvider>);
        await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
        expect(h.byLabel('Cancel')).toBeDefined();
        if (action === 'cancel') {
          await h.press(h.byLabel('Cancel'));
          expect(submitted).toEqual([]);
        } else {
          if (route === 'edit') await h.changeText(h.byLabel('Asset name'), 'Updated tent');
          else {
            await h.changeText(h.byLabel(route === 'move' ? 'Put in' : 'Find item, box, or place'), 'Camping');
            await h.run(() => new Promise(resolve => setTimeout(resolve, 350)));
            await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
            await h.press(h.byText('Camping box')?.parent?.parent?.parent ?? undefined);
          }
          await h.press(h.byLabel(route === 'edit' ? 'Save' : route === 'move' ? 'Move' : 'Move here'));
          expect(submitted).toHaveLength(1);
          expect(consumeAssetActionCompletion('asset')?.action).toBe(route === 'edit' ? 'edit' : 'move');
        }
        expect(dispatchedActions()).toEqual([canGoBack ? { type: 'back' } : { type: 'replace', href: '/' }]);
      } finally { await h.unmount(); client.clear(); resetNavigation(); setCanGoBack(true); }
    });
  }
}

it('keeps direct-entry edits until discard is confirmed, then returns Home', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation(); setCanGoBack(false);
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }} updateAssetCommand={{ execute: async () => { throw new Error('Unexpected save'); } }} />
    </MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 30)));
    await h.changeText(h.byLabel('Asset name'), 'Keep my draft'); await h.press(h.byLabel('Cancel'));
    expect(latestAlert()?.title).toBe('Discard changes?');
    await h.run(() => pressAlertButton('Keep editing'));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byLabel('Asset name')?.props.value).toBe('Keep my draft');
    await h.press(h.byLabel('Cancel')); await h.run(() => pressAlertButton('Discard'));
    expect(dispatchedActions()).toEqual([{ type: 'replace', href: '/' }]);
  } finally { await h.unmount(); client.clear(); resetNavigation(); setCanGoBack(true); }
});
