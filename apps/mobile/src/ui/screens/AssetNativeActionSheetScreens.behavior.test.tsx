import React from 'react';
import { attemptNavigation, dispatchedActions, resetNavigation } from '../../test-support/navigation';
import { expect, it } from 'vitest';
import { AssetEditSheetRouteScreen, AssetMoveHereSheetRouteScreen, AssetMoveSheetRouteScreen } from './AssetNativeActionSheetScreens';
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
      <AssetEditSheetRouteScreen inventoryAssetTypesQuery={{ execute: async () => [] }} assetId="asset" assetCoreQuery={query} inventoryAssetTagsQuery={{ execute: () => new Promise(() => undefined) }} updateAssetCommand={{ execute: async () => { throw new Error('No save requested'); } }} />
    </MobileServerStateProvider>);
    await settle(harness); await settle(harness);
    const name = harness.byLabel('Asset name');
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
    await harness.changeText(harness.byLabel('Put in'), 'My destination');
    parent = 'new-parent';
    await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(harness);
    expect(harness.allByType('TextInput').some((input) => input.props.value === 'My destination')).toBe(true);
  } finally { await harness.unmount(); }
});

it('submits an expiration clear through the native edit route', async () => {
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  const saved: unknown[] = [];
  const asset = { id: assetId('asset'), title: 'Medicine', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false,
    customAssetTypeId: 'medicine', expiration: { date: '2028-02', precision: 'month' as const } };
  const query = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={query} inventoryAssetTagsQuery={{ execute: async () => [] }}
        inventoryAssetTypesQuery={{ execute: async () => [{ kind: 'asset-type', id: 'medicine', key: 'medicine', displayName: 'Medicine', description: '', tenantId: 'tenant', inventoryId: 'inventory', scope: 'inventory', lifecycle: 'active', expirationEnabled: true }] }}
        updateAssetCommand={{ execute: async (input) => { saved.push(input); return { id: 'asset', title: 'Medicine', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(harness); await settle(harness);
    await harness.press(harness.byLabel('Clear expiration'));
    await harness.press(harness.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ assetId: 'asset', expiration: null })]);
  } finally { await harness.unmount(); }
});

it('creates the move destination with the kind selected in the native menu', async () => {
  const client = createMobileQueryClient(); const h = new MobileRenderHarness(); const created: unknown[] = [];
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core}
        createAssetCommand={{ execute: async input => { created.push(input); return { id: 'box', title: 'Camping box', message: 'Created' }; } }}
        moveAssetCommand={{ execute: async () => { throw new Error('No move requested'); } }} parentLookupQuery={{ execute: async () => [] }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.allByType('TextInput').find(input => input.props.placeholder === 'Search places, boxes, shelves'), 'Camping box');
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    await h.press(h.byLabel('Choose destination kind')); await h.press(h.byLabel('Container'));
    const create = h.allByType('Text').find(node => node.children.join('') === 'Create container "Camping box"')?.parent;
    await h.press(create ?? undefined);
    expect(created).toEqual([expect.objectContaining({ kind: 'container', title: 'Camping box' })]);
    expect(h.allText().join(' ')).toContain('Camping box');
  } finally { await h.unmount(); }
});

it.each(['move', 'move-here'] as const)('freezes %s submission and restores its draft after failure', async mode => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const submitted: unknown[] = [];
  let rejectSave: (error: Error) => void = () => {};
  const waiting = new Promise<never>((_, reject) => { rejectSave = reject; });
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'container' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  const candidate = { id: 'box', title: 'Camping box', kind: 'container' as const, subtitle: '', pathLabel: 'Camping box', selectionHint: 'Container', willPromoteToContainer: false };
  const props = { assetId: 'asset', assetCoreQuery: core, parentLookupQuery: { execute: async () => [candidate] }, moveAssetCommand: { execute: async (input: unknown) => { submitted.push(input); return waiting; } } };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      {mode === 'move' ? <AssetMoveSheetRouteScreen {...props} createAssetCommand={{ execute: async () => { throw new Error('No create requested'); } }} /> : <AssetMoveHereSheetRouteScreen {...props} />}
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    const input = h.byLabel(mode === 'move' ? 'Put in' : 'Find item, box, or place');
    await h.changeText(input, 'Camping'); await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
    const candidateRow = h.byText('Camping box')?.parent?.parent?.parent;
    await h.press(candidateRow ?? undefined);
    const save = h.byText(mode === 'move' ? 'Move' : 'Move here')?.parent;
    expect(save?.props.disabled).toBe(false);
    const submit = save!.props.onPress;
    await h.run(() => { submit(); submit(); });
    expect(submitted).toHaveLength(1);
    expect(h.byText('Cancel')?.parent?.props.disabled).toBe(true);
    expect(h.allByType('TextInput')[0]?.props.editable).toBe(false);
    await h.changeText(input, 'Wrong destination');
    expect(h.allByType('TextInput')[0]?.props.value).toBe('Camping');
    await h.run(() => rejectSave(new Error('Failed'))); await settle(h);
    expect(h.byText('Cancel')?.parent?.props.disabled).toBe(false);
    expect(h.allByType('TextInput')[0]?.props.value).toBe('Camping');
  } finally { await h.unmount(); }
});

it('shares the Move lock with destination creation and retains the query after failure', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let creates = 0; let moves = 0;
  let rejectCreate: (error: Error) => void = () => {};
  const waiting = new Promise<never>((_, reject) => { rejectCreate = reject; });
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, parentAssetId: assetId('old'), locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core} parentLookupQuery={{ execute: async () => [] }}
        createAssetCommand={{ execute: async () => { creates++; return waiting; } }}
        moveAssetCommand={{ execute: async () => { moves++; return { id: 'asset', title: 'Tent', message: 'Moved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.allByType('TextInput')[0], 'New box');
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    const create = h.byText('Create location "New box"')?.parent;
    const move = h.byText('Move')?.parent;
    await h.run(() => { create!.props.onPress(); create!.props.onPress(); move!.props.onPress(); });
    expect(creates).toBe(1); expect(moves).toBe(0);
    expect(h.byLabel('Choose destination kind')?.props.disabled).toBe(true);
    expect(h.allText()).toContain('Creating destination…');
    await h.changeText(h.allByType('TextInput')[0], 'Changed');
    await h.run(() => rejectCreate(new Error('Failed'))); await settle(h);
    expect(h.allByType('TextInput')[0]?.props.value).toBe('New box');
    expect(h.byText('Cancel')?.parent?.props.disabled).toBe(false);
  } finally { await h.unmount(); }
});

it.each(['failure', 'success', 'late completion'] as const)('protects Edit draft during submission and %s', async outcome => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let writes = 0; let unmounted = false;
  resetNavigation();
  let resolveSave: (value: { id: string; title: string; message: string }) => void = () => {};
  let rejectSave: (error: Error) => void = () => {};
  const waiting = new Promise<{ id: string; title: string; message: string }>((resolve, reject) => { resolveSave = resolve; rejectSave = reject; });
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
        updateAssetCommand={{ execute: async () => { writes++; return waiting; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    const input = h.allByType('TextInput')[0];
    await h.changeText(input, 'Submitted name');
    const save = h.byText('Save')!.parent!.props.onPress;
    await h.run(() => { save(); save(); });
    expect(writes).toBe(1);
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byText('Cancel')?.parent?.props.disabled).toBe(true);
    await h.changeText(input, 'Unsubmitted name');
    expect(h.allByType('TextInput')[0]?.props.value).toBe('Submitted name');
    if (outcome === 'failure') {
      await h.run(() => rejectSave(new Error('Failed'))); await settle(h);
      expect(h.allByType('TextInput')[0]?.props.value).toBe('Submitted name');
      expect(h.byText('Cancel')?.parent?.props.disabled).toBe(false);
    } else if (outcome === 'success') {
      await h.run(() => resolveSave({ id: 'asset', title: 'Submitted name', message: 'Saved' }));
      expect(dispatchedActions()).toEqual([{ type: 'back' }]);
    } else {
      await h.unmount(); unmounted = true;
      await h.run(() => resolveSave({ id: 'asset', title: 'Submitted name', message: 'Saved' }));
      expect(dispatchedActions()).toEqual([]);
    }
  } finally { if (!unmounted) await h.unmount(); resetNavigation(); }
});

it('keeps an existing tag selected when inline tag resolution updates the Edit draft', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  const asset = { id: assetId('asset'), title: 'Tent', description: 'Keep this description', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }}
        inventoryAssetTagsQuery={{ execute: async () => [{ id: 'camping', key: 'camping', label: 'Camping' }] }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('New tag name'), '  CAMPING  ');
    await h.press(h.byLabel('Add tag'));
    const save = h.byText('Save')?.parent;
    expect(save?.props.disabled).toBe(false);
    await h.press(save ?? undefined);
    expect(saved).toEqual([expect.objectContaining({ tagIds: ['camping'], newTags: [], description: 'Keep this description' })]);
  } finally { await h.unmount(); }
});

it('retries failed Edit metadata independently while retaining the dirty name', async () => {
  const client = createMobileQueryClient(); client.setDefaultOptions({ queries: { retry: false } });
  const harness = new MobileRenderHarness(); let typeReads = 0; let tagReads = 0;
  const query = new AssetCoreQuery({ getAssetCore: async () => ({
    tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: 'one',
    asset: { id: assetId('asset'), title: 'Tent', description: '', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={query}
        inventoryAssetTypesQuery={{ execute: async () => { if (++typeReads === 1) throw new Error('Types unavailable'); return []; } }}
        inventoryAssetTagsQuery={{ execute: async () => { if (++tagReads === 1) throw new Error('Tags unavailable'); return []; } }}
        updateAssetCommand={{ execute: async () => { throw new Error('Save not requested'); } }} />
    </MobileServerStateProvider>);
    await settle(harness); await settle(harness);
    const formScroll = harness.allByType('ScrollView').find(node =>
      node.queryAll(child => child.props.accessibilityLabel === 'Asset name').length > 0);
    expect(formScroll).toBeDefined();
    for (const label of ['Retry asset types', 'Retry tags']) {
      expect(formScroll?.queryAll(child => child.props.accessibilityLabel === label).length).toBeGreaterThan(0);
    }
    expect(formScroll?.queryAll(child => child.props.accessibilityLabel === 'Cancel')).toHaveLength(0);
    await harness.changeText(harness.byLabel('Asset name'), 'My retained name');
    await harness.press(harness.byLabel('Retry tags')); await settle(harness);
    expect(tagReads).toBe(2); expect(typeReads).toBe(1);
    expect(harness.byLabel('Retry tags')).toBeUndefined();
    expect(harness.byLabel('Retry asset types')).toBeDefined();
    await harness.press(harness.byLabel('Retry asset types')); await settle(harness);
    expect(typeReads).toBe(2);
    expect(harness.byLabel('Asset name')?.props.value).toBe('My retained name');
  } finally { await harness.unmount(); }
});

it('does not call failed move-here suggestions empty and recovers inside results', async () => {
  const client = createMobileQueryClient(); client.setDefaultOptions({ queries: { retry: false } });
  const h = new MobileRenderHarness(); let unavailable = true;
  const core = new AssetCoreQuery({ getAssetCore: async () => ({
    tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Camping box', description: '', kind: 'container', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveHereSheetRouteScreen assetId="asset" assetCoreQuery={core}
        parentLookupQuery={{ execute: async () => { if (unavailable) throw new Error('Unavailable'); return []; } }}
        moveAssetCommand={{ execute: async () => { throw new Error('Move not requested'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('Find item, box, or place'), 'Tent');
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    expect(h.byLabel('Retry suggestions')).toBeDefined();
    expect(h.byText('No movable matches')).toBeUndefined();
    expect(h.allByType('ScrollView').some(node => node.queryAll(child => child.props.accessibilityLabel === 'Retry suggestions').length > 0)).toBe(true);
    unavailable = false;
    await h.press(h.byLabel('Retry suggestions')); await settle(h);
    expect(h.byText('No movable matches')).toBeDefined();
    expect(h.byLabel('Find item, box, or place')?.props.value).toBe('Tent');
  } finally { await h.unmount(); }
});

it('waits for known Move suggestions before offering destination creation', async () => {
  const client = createMobileQueryClient(); client.setDefaultOptions({ queries: { retry: false } });
  const h = new MobileRenderHarness(); let unavailable = true;
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Tent', description: '', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core}
        createAssetCommand={{ execute: async () => { throw new Error('Create not requested'); } }}
        moveAssetCommand={{ execute: async () => { throw new Error('Move not requested'); } }}
        parentLookupQuery={{ execute: async () => { if (unavailable) throw new Error('Unavailable'); return []; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('Put in'), 'New room');
    expect(h.byText('Create location "New room"')).toBeUndefined();
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    expect(h.byLabel('Retry suggestions')).toBeDefined();
    expect(h.byText('Create location "New room"')).toBeUndefined();
    expect(h.allByType('ScrollView').some(node => node.queryAll(child => child.props.accessibilityLabel === 'Retry suggestions').length > 0)).toBe(true);
    unavailable = false; await h.press(h.byLabel('Retry suggestions')); await settle(h);
    expect(h.byText('Create location "New room"')).toBeDefined();
    expect(h.byLabel('Put in')?.props.value).toBe('New room');
  } finally { await h.unmount(); }
});
