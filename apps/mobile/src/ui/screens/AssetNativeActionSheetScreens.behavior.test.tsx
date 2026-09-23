import { AssetTagSelectionTaskProvider, useAssetTagSelectionTask } from '../navigation/AssetTagSelectionTask';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import { consumeAssetActionCompletion } from './AssetActionCompletion';
import { latestAlert, pressAlertButton } from '../../test-support/react-native';
import React from 'react';
import { Platform } from 'react-native';
import { attemptNavigation, dispatchedActions, resetNavigation, navigationOptions, setScreenFocused, setCanGoBack } from '../../test-support/navigation';
import { afterEach, beforeEach, expect, it } from 'vitest';
import { NativeSearchDriver } from '../../test-support/NativeSearchDriver';
let moveSearch: NativeSearchDriver;
beforeEach(() => { moveSearch = new NativeSearchDriver(); });
afterEach(() => moveSearch.dispose());
import { AssetEditSheetRouteScreen, AssetMoveHereSheetRouteScreen, AssetMoveSheetRouteScreen } from './AssetNativeActionSheetScreens';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));

it.each([
  ['loading', false], ['error', false], ['loading', true], ['error', true]
] as const)('exposes Close from Edit %s without a loaded draft (back=%s)', async (state, canGoBack) => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation(); setCanGoBack(canGoBack);
  const core = new AssetCoreQuery({ getAssetCore: async () => {
    if (state === 'error') throw new Error('Unavailable');
    return new Promise(() => undefined);
  } });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
        updateAssetCommand={{ execute: async () => { throw new Error('Must not save'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    expect(state === 'error' ? h.byLabel('Retry asset') : h.byText('Loading asset')).toBeDefined();
    expect(h.byLabel('Close')).toBeDefined();
    await h.press(h.byLabel('Close'));
    expect(dispatchedActions()).toContainEqual(canGoBack ? { type: 'back' } : { type: 'replace', href: '/' });
  } finally { await h.unmount(); client.clear(); resetNavigation(); setCanGoBack(true); }
});

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
    await harness.run(() => moveSearch.change('My destination'));
    parent = 'new-parent';
    await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(harness);
    expect(moveSearch.text).toBe('My destination');
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
  let lookupResults: ParentLookupResult[] = [];
  const client = createMobileQueryClient(); const h = new MobileRenderHarness(); const created: unknown[] = [];
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core}
        createAssetCommand={{ execute: async input => { created.push(input); if (created.length === 1) throw new Error('Create temporarily unavailable'); return { id: 'box', title: 'Camping box', message: 'Created' }; } }}
        moveAssetCommand={{ execute: async () => { throw new Error('No move requested'); } }} parentLookupQuery={{ execute: async () => lookupResults }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    const originalSearch = moveSearch.options;
    await h.run(() => moveSearch.change('camping box'));
    expect(moveSearch.options).toBe(originalSearch);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    expect(h.byLabel('Choose destination kind')).toBeUndefined();
    await h.press(h.byLabel('New destination'));
    expect(h.byLabel('Choose destination kind')).toBeDefined();
    await h.press(h.byLabel('Cancel new destination'));
    expect(h.byLabel('Choose destination kind')).toBeUndefined();
    expect(moveSearch.text).toBe('camping box');
    await h.press(h.byLabel('New destination'));
    await h.press(h.byLabel('Choose destination kind')); await h.press(h.byLabel('Container'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
    const create = h.allByType('Text').find(node => node.children.join('') === 'Create container "camping box"')?.parent;
    await h.press(create ?? undefined);
    expect(latestAlert()?.title).toBe('Could not create destination');
    expect(moveSearch.options).toBeUndefined();
    expect(moveSearch.text).toBe('camping box');
    await h.press(create ?? undefined);
    expect(created).toEqual([expect.objectContaining({ kind: 'container', title: 'camping box' }), expect.objectContaining({ kind: 'container', title: 'camping box' })]);
    expect(moveSearch.options).toBeDefined();
    expect(moveSearch.text).toBe('Camping box');
    expect(h.allText().join(' ')).toContain('Camping box');
    expect(h.byLabel('Create container "Camping box"')).toBeUndefined();
    const selectedRows = () => h.allByType('Pressable').filter(row => row.props.accessibilityState?.checked === true);
    expect(selectedRows()).toHaveLength(1);
    await h.run(() => moveSearch.change('Kitchen'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
    expect(selectedRows()).toHaveLength(0);
    expect(h.byLabel('Choose destination kind')).toBeUndefined();
    await h.press(h.byLabel('New destination'));
    expect(h.byLabel('Create container "Kitchen"')).toBeDefined();
    await h.press(h.byLabel('Cancel new destination'));
    await h.run(() => moveSearch.change('  CAMPING box  '));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
    expect(selectedRows()).toHaveLength(1);
    expect(h.byLabel('Create container "CAMPING box"')).toBeUndefined();
    lookupResults = [{ id: 'box', title: 'Camping box', kind: 'container', subtitle: '', pathLabel: 'Garage / Camping box', selectionHint: 'Container', willPromoteToContainer: false }];
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.parentCandidates('scope', 'tenant', 'inventory', 'CAMPING box') }));
    await settle(h);
    expect(selectedRows()).toHaveLength(1);
    expect(h.allText().join(' ')).toContain('Garage / Camping box');
    await h.press(h.byLabel('Move')); await settle(h);
    expect(latestAlert()?.title).toBe('Could not move asset');
    expect(selectedRows()).toHaveLength(1);

  } finally { await h.unmount(); }
});

it.each([['move', false, 'failure'], ['move-here', false, 'failure'], ['move', true, 'failure'], ['move-here', true, 'failure'], ['move', true, 'success'], ['move-here', true, 'success'], ['move', false, 'success'], ['move-here', false, 'success']] as const)('owns %s submission (returned=%s, outcome=%s)', async (mode, returned, outcome) => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const submitted: unknown[] = [];
  let rejectSave: (error: Error) => void = () => {};
  let resolveSave!: (value: { id: string; title: string; message: string }) => void;
  const waiting = new Promise<{ id: string; title: string; message: string }>((resolve, reject) => { resolveSave = resolve; rejectSave = reject; });
  resetNavigation(); consumeAssetActionCompletion('asset');
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'container' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  const candidate = { id: 'box', title: 'Camping box', kind: 'container' as const, subtitle: '', pathLabel: 'Camping box', selectionHint: 'Container', willPromoteToContainer: false };
  const props = { assetId: 'asset', assetCoreQuery: core, parentLookupQuery: { execute: async () => [candidate] }, moveAssetCommand: { execute: async (input: unknown) => { submitted.push(input); return waiting; } } };
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      {mode === 'move' ? <AssetMoveSheetRouteScreen {...props} createAssetCommand={{ execute: async () => { throw new Error('No create requested'); } }} /> : <AssetMoveHereSheetRouteScreen {...props} />}
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    if (mode === 'move') {
      expect(h.byText('Tent')).toBeDefined();
      expect(h.byText('Current location: Inventory root')).toBeDefined();
      expect(h.byText('Selected: Camping box')).toBeUndefined();
      expect(h.byLabel('Move')?.props.disabled).toBe(true);
    }
    const retainedSearch = moveSearch.options;
    await h.run(() => moveSearch.change('Camping')); await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
    const candidateRow = h.byLabel(mode === 'move' ? 'Choose destination Camping box' : 'Choose item Camping box');
    await h.press(candidateRow ?? undefined);
    if (mode === 'move') expect(h.byText('Selected: Camping box')).toBeDefined();
    const save = h.byLabel(mode === 'move' ? 'Move' : 'Move here');
    expect(save?.props.disabled).toBe(false);
    const submit = save!.props.onPress;
    await h.run(() => { submit(); submit(); });
    expect(submitted).toHaveLength(1);
    if (mode === 'move') expect(h.byText('Moving…')).toBeDefined();
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byLabel('Cancel')?.props.disabled).toBe(true);
    expect(moveSearch.options).toBeUndefined();
    await h.run(() => retainedSearch!.onChangeText({ nativeEvent: { text: 'Wrong destination' } }));
    const alertBefore = latestAlert();
    if (returned) { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
    await h.run(() => outcome === 'failure' ? rejectSave(new Error('Failed')) : resolveSave({ id: 'asset', title: 'Tent', message: 'Moved' })); await settle(h);
    if (returned) expect(latestAlert()).toBe(alertBefore);
    expect(dispatchedActions()).toEqual(outcome === 'success' && !returned ? [{ type: 'back' }] : []);
    const completion = consumeAssetActionCompletion('asset');
    if (outcome === 'success' && !returned) expect(completion?.action).toBe('move');
    else expect(completion).toBeUndefined();
    expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
    expect(moveSearch.text).toBe('Camping');
    if (outcome === 'failure' && !returned) {
      const action = { type: 'GO_BACK', source: mode };
      await h.run(() => attemptNavigation(action));
      expect(dispatchedActions()).toEqual([action]);
    }
  } finally { await h.unmount(); setScreenFocused(true); resetNavigation(); }
});

it.each([[false, 'failure'], [true, 'failure'], [true, 'success']] as const)('shares Move creation lock (returned=%s, outcome=%s)', async (returned, outcome) => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let creates = 0; let moves = 0;
  let rejectCreate: (error: Error) => void = () => {};
  let resolveCreate!: (value: { id: string; title: string; message: string }) => void;
  const waiting = new Promise<{ id: string; title: string; message: string }>((resolve, reject) => { resolveCreate = resolve; rejectCreate = reject; });
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, parentAssetId: assetId('old'), locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core} parentLookupQuery={{ execute: async () => [] }}
        createAssetCommand={{ execute: async () => { creates++; return waiting; } }}
        moveAssetCommand={{ execute: async () => { moves++; return { id: 'asset', title: 'Tent', message: 'Moved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.run(() => moveSearch.change('New box'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    await h.press(h.byLabel('New destination'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
    const create = h.byLabel('Create location "New box"');
    expect(create).toBeDefined();
    const move = h.byLabel('Move');
    await h.run(() => { create!.props.onPress(); create!.props.onPress(); move!.props.onPress(); });
    expect(creates).toBe(1); expect(moves).toBe(0);
    expect(h.byLabel('Create location "New box"')?.props.disabled).toBe(true);
    expect(h.byLabel('Choose destination kind')?.props.disabled).toBe(true);
    expect(h.allText()).toContain('Creating destination…');
    await h.changeText(h.byLabel('New destination name'), 'Changed');
    const alertBefore = latestAlert();
    if (returned) { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
    await h.run(() => outcome === 'failure' ? rejectCreate(new Error('Failed')) : resolveCreate({ id: 'new', title: 'Created destination', message: 'Created' })); await settle(h);
    if (returned) expect(latestAlert()).toBe(alertBefore);
    expect(h.byLabel('New destination name')?.props.value).toBe('New box');
    expect(h.byText('Created destination')).toBeUndefined();
    expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
  } finally { await h.unmount(); setScreenFocused(true); }
});

it.each(['failure', 'success', 'late completion', 'return success', 'return failure'] as const)('protects Edit draft during submission and %s', async outcome => {
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
    const save = h.byLabel('Save')!.props.onPress;
    await h.run(() => { save(); save(); });
    expect(writes).toBe(1);
    expect(h.byText('Saving changes…')).toBeDefined();
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    expect(dispatchedActions()).toEqual([]);
    expect(h.byLabel('Cancel')?.props.disabled).toBe(true);
    await h.changeText(input, 'Unsubmitted name');
    expect(h.allByType('TextInput')[0]?.props.value).toBe('Submitted name');
    if (outcome === 'return success' || outcome === 'return failure') {
      consumeAssetActionCompletion('asset');
      const alertBefore = latestAlert();
      await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true));
      await h.run(() => outcome === 'return success' ? resolveSave({ id: 'asset', title: 'Submitted name', message: 'Saved' }) : rejectSave(new Error('Departed failure')));
      await settle(h);
      expect(dispatchedActions()).toEqual([]);
      expect(latestAlert()).toBe(alertBefore);
      expect(consumeAssetActionCompletion('asset')).toBeUndefined();
      expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
      expect(h.allByType('TextInput')[0]?.props.value).toBe('Submitted name');
    } else if (outcome === 'failure') {
      await h.run(() => rejectSave(new Error('Failed'))); await settle(h);
      expect(h.allByType('TextInput')[0]?.props.value).toBe('Submitted name');
      expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
    } else if (outcome === 'success') {
      await h.run(() => resolveSave({ id: 'asset', title: 'Submitted name', message: 'Saved' }));
      expect(dispatchedActions()).toEqual([{ type: 'back' }]);
    } else {
      await h.unmount(); unmounted = true;
      await h.run(() => resolveSave({ id: 'asset', title: 'Submitted name', message: 'Saved' }));
      expect(dispatchedActions()).toEqual([]);
    }
  } finally { if (!unmounted) await h.unmount(); setScreenFocused(true); resetNavigation(); }
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
    await h.press(h.byLabel('New tag'));
    await h.changeText(h.byLabel('New tag name'), '  CAMPING  ');
    await h.press(h.byLabel('Add tag'));
    const save = h.byLabel('Save');
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
    expect(formScroll?.queryAll(child => child.children.join('') === 'Edit asset')).toHaveLength(0);
    expect(navigationOptions()).toEqual(expect.arrayContaining([expect.objectContaining({
      headerLeft: expect.any(Function), headerRight: expect.any(Function), headerBackVisible: false
    })]));
    for (const label of ['Retry asset types', 'Retry tags']) {
      expect(formScroll?.queryAll(child => child.props.accessibilityLabel === label).length).toBeGreaterThan(0);
    }
    expect(formScroll?.queryAll(child => child.props.accessibilityLabel === 'Cancel')).toHaveLength(0);
    expect(harness.byText('Loading expiration settings…')).toBeUndefined();
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
    await h.run(() => moveSearch.change('Tent'));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    expect(h.byLabel('Retry suggestions')).toBeDefined();
    const form = h.allByType('ScrollView').find(node => node.queryAll(child => child.props.accessibilityLabel === 'Retry suggestions').length > 0);
    expect(moveSearch.options).toBeDefined();
    expect(form?.queryAll(child => child.props.accessibilityLabel === 'Cancel')).toHaveLength(0);
    expect(h.byText('No movable matches')).toBeUndefined();
    expect(h.allByType('ScrollView').some(node => node.queryAll(child => child.props.accessibilityLabel === 'Retry suggestions').length > 0)).toBe(true);
    unavailable = false;
    await h.press(h.byLabel('Retry suggestions')); await settle(h);
    expect(h.byText('No movable matches')).toBeDefined();
    expect(moveSearch.text).toBe('Tent');
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
    await h.run(() => moveSearch.change('New room'));
    expect(h.byText('Create location "New room"')).toBeUndefined();
    await h.run(() => new Promise(resolve => setTimeout(resolve, 400))); await settle(h);
    expect(h.byLabel('Retry suggestions')).toBeDefined();
    const form = h.allByType('ScrollView').find(node => node.queryAll(child => child.props.accessibilityLabel === 'Retry suggestions').length > 0);
    expect(moveSearch.options).toBeDefined();
    expect(h.byLabel('Put in')).toBeUndefined();
    expect(form?.queryAll(child => child.props.accessibilityLabel === 'Cancel')).toHaveLength(0);
    expect(h.byText('Create location "New room"')).toBeUndefined();
    expect(h.allByType('ScrollView').some(node => node.queryAll(child => child.props.accessibilityLabel === 'Retry suggestions').length > 0)).toBe(true);
    unavailable = false; await h.press(h.byLabel('Retry suggestions')); await settle(h);
    expect(h.byLabel('Choose destination kind')).toBeUndefined();
    await h.press(h.byLabel('New destination'));
    expect(h.byText('Create location "New room"')).toBeDefined();
    expect(moveSearch.text).toBe('New room');
  } finally { await h.unmount(); }
});

it('names staged Edit tag removal and preserves other tags and edited fields', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Tent', description: 'Keep me', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('Description'), 'Keep my edit');
    await h.press(h.byLabel('New tag'));
    for (const name of ['Camping', 'Outdoors']) {
      await h.changeText(h.byLabel('New tag name'), name); await h.press(h.byLabel('Add tag'));
    }
    const remove = h.byLabel('Remove new tag Camping');
    expect(remove).toBeDefined();
    expect(remove?.props.accessibilityState.selected).toBeUndefined();
    await h.press(remove);
    expect(h.byLabel('Remove new tag Camping')).toBeUndefined();
    expect(h.byLabel('Remove new tag Outdoors')).toBeDefined();
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ description: 'Keep my edit', newTags: [{ displayName: 'Outdoors' }] })]);
  } finally { await h.unmount(); client.clear(); }
});

it.each(['ios', 'android'])('preserves rejected Edit tag drafts and resets accepted input on %s', async platform => {
  const originalPlatform = Platform.OS; Platform.OS = platform as typeof Platform.OS;
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Tent', description: 'Keep me', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    expect(h.byText('Use a shorter tag name.')).toBeUndefined();
    await h.press(h.byLabel('New tag'));
    const input = h.byLabel('New tag name');
    const longName = 'Camping equipment '.repeat(8);
    await h.changeText(h.byLabel('New tag name'), longName);
    expect(h.byText('Use a shorter tag name.')).toBeDefined();
    expect(h.byLabel('New tag name')?.props.value).toBe(longName);
    await h.press(h.byLabel('Add tag'));
    expect(h.byLabel('New tag name')?.props.value).toBe(longName);
    await h.changeText(h.byLabel('New tag name'), 'Camping');
    expect(h.byText('Use a shorter tag name.')).toBeUndefined();
    expect(h.byLabel('New tag name')).toBe(input);
    await h.press(h.byLabel('Add tag'));
    const cleared = h.byLabel('New tag name');
    if (platform === 'ios') expect(cleared).not.toBe(input);
    else expect(cleared).toBe(input);
    expect(cleared?.props.value).toBe('');
    expect(h.byLabel('Remove new tag Camping')).toBeDefined();
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ description: 'Keep me', newTags: [{ displayName: 'Camping' }] })]);
  } finally { await h.unmount(); Platform.OS = originalPlatform; }
});


it('selects existing Edit tags in a separate visit and saves the parent draft', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  const asset = { id: assetId('asset'), title: 'Tent', description: 'Keep description', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  const tags = Array.from({ length: 14 }, (_, index) => ({ id: `tag-${index + 1}`, key: `tag-${index + 1}`, label: `Tag ${index + 1}` })).reverse();
  try {
    await h.render(<AssetTagSelectionTaskProvider><SelectionContent /><MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => tags }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } }} />
    </MobileServerStateProvider></AssetTagSelectionTaskProvider>);
    await settle(h); await settle(h);
    expect(h.byText('Tag 13')).toBeUndefined();
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Select tag Tag 14'));
    await h.press(h.byLabel('Cancel selecting tags'));
    await h.press(h.byLabel('Choose tags'));
    expect(h.byLabel('Select tag Tag 14')?.props.accessibilityState.checked).toBe(false);
    await h.press(h.byLabel('Select tag Tag 14'));
    await h.press(h.byLabel('Done selecting tags'));
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ tagIds: ['tag-14'], description: 'Keep description' })]);
  } finally { await h.unmount(); }
});


it('protects an unstaged Edit tag from cancellation and silent omission on Save', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.press(h.byLabel('New tag'));
    await h.changeText(h.byLabel('New tag name'), 'Camping');
    await h.press(h.byLabel('Cancel'));
    expect(latestAlert()?.title).toBe('Discard changes?');
    await h.run(() => pressAlertButton('Keep editing'));
    expect(h.byLabel('New tag name')?.props.value).toBe('Camping');
    await h.changeText(h.byLabel('Description'), 'Keep this edit');
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([]);
    expect(h.byText('Add this tag or clear its name and color before saving.')).toBeDefined();
    await h.press(h.byLabel('Add tag'));
    expect(h.byLabel('New tag name')?.props.value).toBe('');
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ description: 'Keep this edit', newTags: [{ displayName: 'Camping' }] })]);
  } finally { await h.unmount(); }
});


it.each(['draft', 'visit', 'unmount', 'current'] as const)('owns Edit discard confirmation across %s', async change => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation();
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }} updateAssetCommand={{ execute: async () => { throw new Error('No save requested'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('Asset name'), 'First draft'); await h.press(h.byLabel('Cancel'));
    const confirm = latestAlert()?.buttons.find(button => button.text === 'Discard')?.onPress;
    expect(confirm).toBeTypeOf('function');
    if (change === 'draft') await h.changeText(h.byLabel('Asset name'), 'Later draft');
    if (change === 'visit') { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
    if (change === 'unmount') await h.unmount();
    const before = dispatchedActions().length;
    await h.run(() => confirm?.()); await h.run(() => confirm?.());
    expect(dispatchedActions().length - before).toBe(change === 'current' ? 1 : 0);
    if (change === 'draft') expect(h.byLabel('Asset name')?.props.value).toBe('Later draft');
  } finally { await h.unmount(); setScreenFocused(true); resetNavigation(); }
});

it.each(['GO_BACK', 'POP'] as const)('protects a dirty Edit draft from native %s removal', async type => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation();
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }} updateAssetCommand={{ execute: async () => { throw new Error('No save requested'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('Asset name'), 'Keep this draft');
    const action = { type, source: 'edit' };
    await h.run(() => attemptNavigation(action));
    expect(dispatchedActions()).toEqual([]);
    expect(latestAlert()?.title).toBe('Discard changes?');
    await h.run(() => pressAlertButton('Keep editing'));
    expect(h.byLabel('Asset name')?.props.value).toBe('Keep this draft');
    await h.run(() => attemptNavigation(action));
    const discard = latestAlert()?.buttons.find(button => button.text === 'Discard')?.onPress;
    await h.run(() => discard?.()); await h.run(() => discard?.());
    expect(dispatchedActions()).toEqual([action]);
  } finally { await h.unmount(); resetNavigation(); }
});


it('retained Edit footer reads the latest draft and rejects hidden or removed actions', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  resetNavigation();
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Tent', description: '', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false } }) });
  let save!: () => void; let cancel!: () => void;
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: input.title ?? '', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('Asset name'), 'Earlier name');
    save = h.byLabel('Save')!.props.onPress; cancel = h.byLabel('Cancel')!.props.onPress;
    await h.changeText(h.byLabel('Asset name'), '');
    await h.run(save);
    expect(saved).toEqual([]);
    await h.changeText(h.byLabel('Asset name'), 'Current name');
    await h.run(() => setScreenFocused(false));
    await h.run(save); await h.run(cancel);
    expect(saved).toEqual([]); expect(dispatchedActions()).toEqual([]);
    await h.run(() => setScreenFocused(true));
    await h.run(save);
    expect(saved).toEqual([expect.objectContaining({ title: 'Current name' })]);
  } finally { await h.unmount(); client.clear(); resetNavigation(); setScreenFocused(true); }
  saved.length = 0;
  await h.run(save); await h.run(cancel);
  expect(saved).toEqual([]); expect(dispatchedActions()).toEqual([]);
});


it.each(['move', 'here'] as const)('retained %s footer uses current selection and ignores blurred cancellation', async mode => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const submitted: unknown[] = [];
  resetNavigation();
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Tent', description: '', kind: 'container', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false } }) });
  const candidates = ['First box', 'Second box'].map((title, index) => ({ id: `box-${index}`, title, kind: 'container' as const, subtitle: '', pathLabel: title, selectionHint: 'Container', willPromoteToContainer: false }));
  const props = { assetId: 'asset', assetCoreQuery: core, parentLookupQuery: { execute: async () => candidates }, moveAssetCommand: { execute: async (input: unknown) => { submitted.push(input); return { id: 'asset', title: 'Tent', message: 'Moved' }; } } };
  let save!: () => void; let cancel!: () => void;
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      {mode === 'move' ? <AssetMoveSheetRouteScreen {...props} createAssetCommand={{ execute: async () => { throw new Error('No creation requested'); } }} /> : <AssetMoveHereSheetRouteScreen {...props} />}
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.press(h.byLabel(mode === 'move' ? 'Choose destination First box' : 'Choose item First box'));
    save = h.byLabel(mode === 'move' ? 'Move' : 'Move here')!.props.onPress;
    cancel = h.byLabel('Cancel')!.props.onPress;
    await h.press(h.byLabel(mode === 'move' ? 'Choose destination Second box' : 'Choose item Second box'));
    await h.run(() => setScreenFocused(false));
    await h.run(save); await h.run(cancel);
    expect(submitted).toEqual([]); expect(dispatchedActions()).toEqual([]);
    await h.run(() => setScreenFocused(true));
    await h.run(save);
    expect(submitted).toEqual([mode === 'move' ? { assetId: 'asset', parentAssetId: 'box-1' } : { assetId: 'box-1', parentAssetId: 'asset' }]);
  } finally { await h.unmount(); client.clear(); resetNavigation(); setScreenFocused(true); }
  submitted.length = 0;
  await h.run(save); await h.run(cancel);
  expect(submitted).toEqual([]); expect(dispatchedActions()).toEqual([]);
});


it('does not retarget retained Edit commands after switching the scoped asset', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  resetNavigation();
  const core = new AssetCoreQuery({ getAssetCore: async requestedId => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId(requestedId), title: requestedId, description: '', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false } }) });
  const render = (id: string) => h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
    <AssetEditSheetRouteScreen assetId={id} assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
      updateAssetCommand={{ execute: async input => { saved.push(input); return { id: input.assetId, title: input.title ?? '', message: 'Saved' }; } }} />
  </MobileServerStateProvider>);
  try {
    await render('first'); await settle(h); await settle(h);
    await h.changeText(h.byLabel('Asset name'), 'First draft');
    const oldSave = h.byLabel('Save')!.props.onPress; const oldCancel = h.byLabel('Cancel')!.props.onPress;
    await render('second'); await settle(h); await settle(h);
    await h.changeText(h.byLabel('Asset name'), 'Second draft');
    await h.run(oldSave); await h.run(oldCancel);
    expect(saved).toEqual([]); expect(dispatchedActions()).toEqual([]);
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ assetId: 'second', title: 'Second draft' })]);
  } finally { await h.unmount(); client.clear(); resetNavigation(); }
});


it('keeps tag creation secondary and cancels only the unstaged tag', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); const saved: unknown[] = [];
  const core = new AssetCoreQuery({ getAssetCore: async () => ({
    tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1',
    asset: { id: assetId('asset'), title: 'Tent', description: 'Packed', kind: 'item', lifecycleState: 'active',
      locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await h.render(<AssetTagSelectionTaskProvider><SelectionContent /><MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }}
        inventoryAssetTagsQuery={{ execute: async () => [{ id: 'camping', key: 'camping', label: 'Camping' }] }}
        updateAssetCommand={{ execute: async input => { saved.push(input); return { id: 'asset', title: 'Tent', message: 'Saved' }; } }} />
    </MobileServerStateProvider></AssetTagSelectionTaskProvider>);
    await settle(h); await settle(h);
    expect(h.byLabel('New tag name')).toBeUndefined();
    expect(h.byLabel('New tag color')).toBeUndefined();
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Select tag Camping'));
    await h.press(h.byLabel('Done selecting tags'));
    await h.changeText(h.byLabel('Description'), 'Ready');
    await h.press(h.byLabel('New tag'));
    const cancelCreation = h.byLabel('Cancel new tag')!.props.onPress;
    await h.changeText(h.byLabel('Description'), 'Ready for the weekend');
    await h.changeText(h.byLabel('New tag name'), 'Outdoors');
    await h.press(h.byLabel('Add tag'));
    await h.changeText(h.byLabel('New tag name'), 'Discard this');
    await h.changeText(h.byLabel('New tag color'), '#123456');
    await h.run(() => cancelCreation());
    expect(h.byLabel('New tag name')).toBeUndefined();
    expect(h.byLabel('Remove new tag Outdoors')).toBeDefined();
    await h.press(h.byLabel('New tag'));
    expect(h.byLabel('New tag name')?.props.value).toBe('');
    expect(h.byLabel('New tag color')?.props.value).toBe('');
    await h.press(h.byLabel('Save'));
    expect(saved).toEqual([expect.objectContaining({ description: 'Ready for the weekend', tagIds: ['camping'], newTags: [{ displayName: 'Outdoors' }] })]);
  } finally { await h.unmount(); client.clear(); }
});

function SelectionContent() { const task = useAssetTagSelectionTask(); return <>{task?.content}</>; }
