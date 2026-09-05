import React from 'react';
import { expect, it } from 'vitest';
import { AssetEditSheetRouteScreen, AssetMoveSheetRouteScreen } from './AssetNativeActionSheetScreens';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
it('opens Edit before tags load and preserves a dirty draft after background core refresh', async () => {
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  let title = 'Tent';
  const query = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: title, asset: { ...asset, title } }) });
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={query} inventoryAssetTagsQuery={{ execute: () => new Promise(() => undefined) }} updateAssetCommand={{ execute: async () => { throw new Error('No save requested'); } }} />
    </MobileServerStateProvider>);
    await settle(harness); await settle(harness);
    const name = harness.allByType('TextInput').find((input) => input.props.value === 'Tent');
    expect(name).toBeDefined();
    await harness.changeText(name, 'My draft');
    title = 'Changed remotely';
    await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(harness);
    expect(harness.allByType('TextInput').some((input) => input.props.value === 'My draft')).toBe(true);
  } finally { await harness.unmount(); }
});

it('keeps a Move draft mounted when background refresh discovers a different parent', async () => {
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  let parent = 'old-parent';
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: parent, asset: { ...asset, parentAssetId: assetId(parent) } }) });
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core} assetPlacementQuery={{ execute: async () => parent === 'old-parent' ? (await core.execute('asset')).view : new Promise(() => undefined) }}
        createAssetCommand={{ execute: async () => { throw new Error('No create requested'); } }} moveAssetCommand={{ execute: async () => { throw new Error('No move requested'); } }} parentLookupQuery={{ execute: async () => [] }} />
    </MobileServerStateProvider>);
    await settle(harness); await settle(harness);
    await harness.changeText(harness.allByType('TextInput').find((input) => input.props.placeholder === 'Search places, boxes, shelves'), 'My destination');
    parent = 'new-parent';
    await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(harness);
    expect(harness.allByType('TextInput').some((input) => input.props.value === 'My destination')).toBe(true);
  } finally { await harness.unmount(); }
});
