import { SettingsQuery } from '../../application/settings/SettingsQuery';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import type { TestInstance } from 'test-renderer';
import { CustomizationFailure } from '../../application/customization/CustomizationErrors';
import { runCustomizationLifecycle } from '../../application/customization/CustomizationEditorCommands';
import { MobileRenderHarness } from '../../test-support/render';
import { alertCount, focusedAccessibilityHandles, focusedInputLabels, latestAlert, pressAlertButton, resetNativeTestState } from '../../test-support/react-native';
import { attemptNavigation, dispatchedActions, navigationOptions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { SettingsSegmentedControl } from '../components/SettingsSegmentedControl';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { CustomizationCollectionScreen } from './CustomizationCollectionScreen';
import { CustomizationEditorScreen } from './CustomizationEditorScreen';
import { DeniedSettingsState, HouseholdSettingsScreen, InventorySettingsScreen } from './ScopedSettingsScreens';

let harness: MobileRenderHarness | undefined;
let queryClient = createMobileQueryClient();
beforeEach(() => { queryClient = createMobileQueryClient(); resetNativeTestState(); resetNavigation(); Reflect.deleteProperty(globalThis, 'expo'); });
afterEach(async () => { await harness?.unmount(); harness = undefined; Reflect.deleteProperty(globalThis, 'expo'); });

describe('rendered mobile customization production states', () => {
  it('retains dormant enum options while saving only options applicable to the chosen field type', async () => {
    const calls: unknown[][] = [];
    const screen = await renderEditor({ kind: 'field', manageFields: managerFake({ create: async (...args: unknown[]) => { calls.push(args); return {}; } }) });
    await screen.changeText(screen.byLabel('Name'), 'Priority');
    await screen.press(screen.byLabel('Choose Type. Current value Text'));
    await screen.press(screen.byLabel('Enum'));
    await screen.changeText(screen.byLabel('New enum option'), 'high');
    await screen.press(screen.byLabel('Add option'));
    await screen.press(screen.byLabel('Choose Type. Current value Enum'));
    await screen.press(screen.byLabel('Text'));
    expect(screen.byLabel('New enum option')).toBeUndefined();
    await screen.press(screen.byLabel('Choose Type. Current value Text'));
    await screen.press(screen.byLabel('Enum'));
    expect(screen.byLabel('Remove high')).toBeDefined();
    await screen.press(screen.byLabel('Choose Type. Current value Enum'));
    await screen.press(screen.byLabel('Text'));
    await screen.press(screen.byLabel('Save'));
    expect(calls).toHaveLength(1);
    expect(calls[0][2]).toMatchObject({ type: 'text', enumOptions: [] });
  });
  it('protects an unsubmitted field option and requires adding it before Save', async () => {
    const record = { ...field('priority', 'Priority', 'inventory'), type: 'enum' as const, enumOptions: ['high'] };
    const calls: unknown[][] = [];
    const screen = await renderEditor({ kind: 'field', mode: 'edit', resourceId: record.id,
      query: collectionQuery({ fields: [record] }), manageFields: managerFake({ update: async (...args: unknown[]) => { calls.push(args); return record; } }) });
    await screen.changeText(screen.byLabel('New enum option'), 'low');
    attemptNavigation({ type: 'BACK' });
    expect(latestAlert()?.title).toBe('Discard changes?');
    await pressAlertButton('Keep Editing'); await settleQueries(screen);
    expect(screen.byLabel('New enum option')?.props.value).toBe('low');
    await screen.changeText(screen.byLabel('Name'), 'Priority level');
    expect(screen.byLabel('Save')?.props.disabled).toBe(true);
    expect(screen.allText()).toContain('Add or clear this option before saving.');
    await screen.press(screen.byLabel('Save'));
    expect(calls).toEqual([]);
    await screen.press(screen.byLabel('Add option'));
    expect(screen.byLabel('New enum option')?.props.value).toBe('');
    expect(screen.byLabel('Save')?.props.disabled).toBe(false);
    await screen.press(screen.byLabel('Save'));
    expect(calls).toHaveLength(1);
    expect(calls[0][2]).toMatchObject({ enumOptions: ['high', 'low'] });
  });
  it('exposes a named Save command and prevents another save while pending', async () => {
    const pending = deferred<Record<string, never>>(); let calls = 0;
    const screen = await renderEditor({ manageTags: managerFake({ create: async () => { calls++; return pending.promise; } }) });
    expect(screen.byLabel('Save')?.props.disabled).toBe(true);
    await screen.changeText(screen.byLabel('Name'), 'Tools');
    await screen.press(screen.byLabel('Save'));
    expect(calls).toBe(1);
    expect(screen.byLabel('Saving…')?.props.disabled).toBe(true);
    await screen.press(screen.byLabel('Saving…'));
    expect(calls).toBe(1);
    await screen.run(() => pending.resolve({})); await screen.settle();
    expect(screen.byText('Saved')).toBeDefined();
  });
  it.each([false, true])('does not navigate from an old save after leaving, returned=%s', async returned => {
    const pending = deferred<Record<string, never>>(); let done = 0;
    const screen = await renderEditor({ manageTags: managerFake({ create: async () => pending.promise }), onDone: () => { done++; } });
    await screen.changeText(screen.byLabel('Name'), 'Tools');
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    expect(screen.allText()).toContain('Saving…');
    await screen.run(() => setScreenFocused(false));
    if (returned) await screen.run(() => setScreenFocused(true));
    await screen.run(() => pending.resolve({}));
    await screen.settle();
    expect(done).toBe(0);
    expect(screen.allText()).not.toContain('Tag saved');
    expect(screen.allText()).not.toContain('Unsaved changes');
    await screen.run(() => setScreenFocused(true));
    expect(screen.allText()).toContain('Saved');
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).not.toContain('Save');
    await screen.press(screen.byLabel('Return to collection'));
    await screen.settle();
    expect(done).toBe(1);
  });

  it('does not navigate from an archive completed after focus returns', async () => {
    const pending = deferred<Record<string, never>>(); let done = 0;
    const record = tag('tools', 'Tools');
    const screen = await renderEditor({ mode: 'edit', resourceId: record.id, query: collectionQuery({ tags: [record] }), manageTags: managerFake({ archive: async () => pending.promise }), onDone: () => { done++; } });
    await screen.press(screen.byText('Archive')?.parent ?? undefined);
    const mutation = pressAlertButton('Archive');
    await screen.run(() => setScreenFocused(false));
    await screen.run(() => setScreenFocused(true));
    await screen.run(() => pending.resolve({}));
    await mutation; await screen.settle();
    expect(done).toBe(0);
    expect(screen.allText()).not.toContain('Tag archived');
    expect(screen.allText()).toContain('Archived');
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).not.toContain('Archive');
  });

  it('rejects an archive confirmation accepted after leaving and permits a fresh confirmation', async () => {
    let archives = 0;
    const record = tag('tools', 'Tools');
    const screen = await renderEditor({ mode: 'edit', resourceId: record.id, query: collectionQuery({ tags: [record] }), manageTags: managerFake({ archive: async () => { archives++; return record; } }) });
    await screen.press(screen.byText('Archive')?.parent ?? undefined);
    await screen.run(() => setScreenFocused(false));
    await screen.run(() => setScreenFocused(true));
    await screen.run(() => pressAlertButton('Archive'));
    expect(archives).toBe(0);
    await screen.press(screen.byText('Archive')?.parent ?? undefined);
    await screen.run(() => pressAlertButton('Archive'));
    expect(archives).toBe(1);
  });

  it.each(['success', 'denied'] as const)('isolates old-resource %s from a new editor draft', async outcome => {
    const pending = deferred<Record<string, never>>();
    const records = [tag('a', 'First'), tag('b', 'Second')];
    const common = { mode: 'edit', query: collectionQuery({ tags: records }), manageTags: managerFake({ update: async () => { await pending.promise; if (outcome === 'denied') throw new CustomizationFailure('permission-denied'); return {}; } }) };
    const screen = await renderEditor({ ...common, resourceId: 'a' });
    await screen.changeText(screen.byLabel('Name'), 'First edited');
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    await screen.render(editorElement({ ...common, resourceId: 'b' }));
    await settleQueries(screen);
    expect(screen.byLabel('Name')?.props.editable).toBe(true);
    await screen.changeText(screen.byLabel('Name'), 'Second draft');
    await screen.run(() => pending.resolve({}));
    await settleQueries(screen);
    expect(screen.byLabel('Name')?.props.value).toBe('Second draft');
    expect(screen.byLabel('Name')?.props.editable).toBe(true);
    expect(screen.allText()).toContain('Unsaved changes');
    expect(screen.allText()).not.toContain('Access changed');
  });

  it('reconciles a mounted collection from query invalidation without discarding the local search', async () => {
    let rows = [tag('one', 'Tools')]; let reads = 0;
    const screen = await renderCollection({ query: { tags: async () => { reads++; return { items: rows, complete: true }; } } });
    await screen.run(() => collectionSearch().onChangeText({ nativeEvent: { text: 'Tool' } }));
    rows = [tag('two', 'Toolboxes'), tag('three', 'Garden')];
    await screen.run(() => queryClient.invalidateQueries({ queryKey: mobileQueryKeys.customization('scope', 'tenant-1', 'inventory-1', 'inventory', 'tag', 'active') }));
    await settleQueries(screen);
    expect(reads).toBe(2);
    expect(screen.allText()).toContain('Toolboxes');
    expect(screen.allText()).not.toContain('Garden');
    expect(screen.allText()).not.toContain('No matches');
    expect(screen.allByType('TextInput')).toHaveLength(0);
    await screen.run(() => collectionSearch().onCancelButtonPress());
    expect(screen.allText()).toContain('Garden');
  });

  it('removes Add and disables its retained handler when collection edit permission is revoked', async () => {
    let editable = true; let added = 0;
    const screen = await renderCollection({
      contextQuery: { execute: async () => context([], editable ? ['view', 'edit_asset'] : ['view']) },
      onAdd: () => { added++; }, query: collectionQuery({ tags: [tag('one', 'Tools')] })
    });
    const add = screen.byLabel('Add Tag')!;
    await screen.press(add);
    expect(added).toBe(1);
    editable = false;
    await screen.run(() => queryClient.invalidateQueries({ queryKey: mobileQueryKeys.settingsScope('scope', 'tenant-1', 'inventory-1') }));
    await settleQueries(screen);
    expect(screen.byLabel('Add Tag')).toBeUndefined();
    await screen.run(() => add.props.onPress());
    expect(added).toBe(1);
    expect(screen.allText()).toContain('Tools');
  });

  it('keeps collection access denied when an older definition request completes late', async () => {
    const delayed = deferred<{ items: readonly ReturnType<typeof tag>[]; complete: true }>();
    let denied = false; let calls = 0;
    const screen = await renderCollection({ contextQuery: { execute: async () => { if (denied) throw new CustomizationFailure('permission-denied'); return context(['configure'], ['view', 'edit_asset']); } }, query: { tags: async () => ++calls === 1 ? { items: [tag('old', 'Old private row')], complete: true } : delayed.promise } });
    await screen.run(() => { void queryClient.invalidateQueries({ queryKey: mobileQueryKeys.customization('scope', 'tenant-1', 'inventory-1', 'inventory', 'tag', 'active') }); });
    denied = true;
    await screen.run(() => queryClient.invalidateQueries({ queryKey: mobileQueryKeys.settingsScope('scope', 'tenant-1', 'inventory-1') })); await settleQueries(screen);
    await screen.run(() => delayed.resolve({ items: [tag('late', 'Late private row')], complete: true })); await settleQueries(screen);
    expect(screen.allText()).toContain('Settings unavailable');
    expect(screen.allText()).not.toContain('Late private row');
  });

  it('refreshes field type choices while preserving a dirty field draft', async () => {
    let types = [assetType('old', 'Old type', 'inventory')];
    const record = { ...field('field', 'Field', 'inventory'), applicability: 'custom_asset_types' };
    const screen = await renderEditor({ kind: 'field', mode: 'edit', resourceId: record.id, query: { fields: async () => ({ items: [record], complete: true }), assetTypes: async () => ({ items: types, complete: true }) } });
    await screen.changeText(screen.byLabel('Name'), 'My dirty field');
    types = [assetType('new', 'New type', 'inventory')];
    await screen.run(() => queryClient.invalidateQueries({ queryKey: mobileQueryKeys.customization('scope', 'tenant-1', 'inventory-1', 'inventory', 'asset-type', 'active') })); await settleQueries(screen);
    expect(screen.byLabel('Name')?.props.value).toBe('My dirty field');
    expect(screen.allText()).toContain('New type');
    expect(screen.allText()).not.toContain('Old type');
  });

  it('aligns collection chrome and editor actions to the shared 16-point content column', async () => {
    const collection = await renderCollection({ query: collectionQuery({ tags: [tag('tag-1', 'Tools')] }) });
    expect(collectionSearch()).toMatchObject({ placement: 'integratedButton', placeholder: 'Search tags' });
    expect(collection.allByType('View').some(hasStyle({ marginHorizontal: 16 }))).toBe(true);

    const editor = await renderEditor();
    await editor.changeText(editor.byLabel('Name'), 'Tools');
    expect(editor.byLabel('Save')?.props.disabled).toBe(false);
    expect(editor.allByType('View').some(node => hasStyle({ marginHorizontal: 16 })(node) && node.queryAll(child => child.props.accessibilityLabel === 'Save').length > 0)).toBe(true);
  });

  it('refreshes permissions and removes retained rows when a collection load is denied', async () => {
    const contexts = sequence(context(['configure'], ['view', 'edit_asset']), context([], []));
    const tags = sequence<{ items: readonly ReturnType<typeof tag>[]; complete: true }>({ items: [tag('private-tag', 'Private tag')], complete: true }, new CustomizationFailure('permission-denied'));
    const screen = await renderCollection({ contextQuery: { execute: contexts.next }, query: { tags: tags.next } });
    expect(screen.allText()).toContain('Private tag');
    const refresh = screen.byType('ScrollView')?.props.refreshControl as React.ReactElement<{ onRefresh: () => void }>;
    refresh.props.onRefresh();
    await settleQueries(screen);
    expect(screen.allText()).toContain('Settings unavailable');
    expect(screen.allText()).not.toContain('Private tag');
  });

  it('renders inherited/local grouping and role-specific mutation controls', async () => {
    const allowed = context(['view'], ['view', 'configure']);
    const screen = await renderCollection({
      contextQuery: { execute: async () => allowed }, kind: 'field', scope: 'inventory',
      query: collectionQuery({ fields: [field('tenant-field', 'Shared field', 'tenant'), field('local-field', 'Local field', 'inventory')] })
    });
    expect(screen.allText()).toEqual(expect.arrayContaining(['From Home', 'Only in Household', 'Shared field', 'Local field']));
    expect(screen.byLabel('Add Custom field')).toBeDefined();
  });

  it('shows inherited inventory detail read-only with an explicit household management action', async () => {
    const inherited = field('tenant-field', 'Shared field', 'tenant');
    let managed = 0;
    const screen = await renderEditor({
      inherited: true, kind: 'field', mode: 'edit', onManageInherited: () => { managed += 1; }, resourceId: inherited.id,
      query: collectionQuery({ fields: [inherited] })
    });
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).toEqual(expect.arrayContaining(['Shared field', 'Inherited from Home. Manage it from household settings.', 'Manage in Home']));
    expect(screen.allText()).not.toContain('Save');
    await screen.press(screen.byText('Manage in Home')?.parent ?? undefined);
    expect(managed).toBe(1);
  });

  it('explains touched validation and uses compact disclosure pickers', async () => {
    const screen = await renderEditor({ kind: 'field', query: collectionQuery({ assetTypes: [assetType('type-1', 'Appliance', 'inventory')] }) });
    expect(screen.allText()).not.toContain('Name is required.');
    await screen.changeText(screen.byLabel('Name'), 'x');
    await screen.changeText(screen.byLabel('Name'), '');
    expect(screen.allText()).toContain('Name is required.');
    expect(screen.byLabel('Choose Type. Current value Text')).toBeDefined();
    expect(screen.byLabel('Choose Applies to. Current value All assets')).toBeDefined();
    expect(screen.byLabel('Choose Type. Current value Text')?.props.accessibilityState.expanded).toBe(false);
    expect(screen.byLabel('Choose Applies to. Current value All assets')?.props.accessibilityState.expanded).toBe(false);
  });

  it('reveals invalid generated stable-key validation and keeps save unavailable', async () => {
    const screen = await renderEditor({ kind: 'asset-type' });
    await screen.changeText(screen.byLabel('Name'), '123');
    expect(screen.byText('Save')?.parent?.props.disabled).toBe(true);
    expect(screen.allText()).toContain('Key must start with a letter and use lowercase letters, numbers, or hyphens.');
    expect(screen.allText()).toContain('Hide technical details');
    expect(screen.byLabel('Stable key')?.props.value).toBe('');
    expect(focusedInputLabels()).toContain('Stable key');
  });

  it('renders viewer tag detail as static labeled values without mutation controls', async () => {
    const record = { ...tag('tag-1', 'Tools'), color: '#2F80ED' };
    const screen = await renderEditor({
      contextQuery: { execute: async () => context(['view'], ['view']) }, kind: 'tag', mode: 'edit', query: collectionQuery({ tags: [record] }), resourceId: record.id
    });
    expect(screen.allText()).toEqual(expect.arrayContaining(['Name', 'Tools', 'Color', 'Blue']));
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).not.toContain('Save');
  });

  it('derives inherited ownership from the loaded record rather than trusting route hints', async () => {
    const inherited = field('tenant-field', 'Shared field', 'tenant');
    const screen = await renderEditor({ inherited: false, kind: 'field', mode: 'edit', query: collectionQuery({ fields: [inherited] }), resourceId: inherited.id });
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).toContain('Inherited from Home. Manage it from household settings.');
    expect(screen.allText()).not.toContain('Archive');
  });

  it('ignores a forged inherited hint when the loaded record is inventory-owned', async () => {
    const local = field('local-field', 'Local field', 'inventory');
    const screen = await renderEditor({ inherited: true, kind: 'field', mode: 'edit', query: collectionQuery({ fields: [local] }), resourceId: local.id });
    expect(screen.byLabel('Name')?.props.editable).toBe(true);
    expect(screen.allText()).not.toContain('Inherited from Home. Manage it from household settings.');
  });

  it.each([
    ['field', 'tenant'], ['field', 'inventory'],
    ['asset-type', 'tenant'], ['asset-type', 'inventory']
  ] as const)('shows the loaded owner of %s definitions from %s in Details', async (kind, recordScope) => {
    const record = kind === 'field'
      ? field('definition', 'Shared definition', recordScope)
      : assetType('definition', 'Shared definition', recordScope);
    const screen = await renderEditor({
      inherited: recordScope !== 'tenant', kind, mode: 'edit', resourceId: record.id,
      query: collectionQuery(kind === 'field' ? { fields: [record] } : { assetTypes: [record] })
    });
    await screen.press(screen.byText('Show technical details')?.parent ?? undefined);
    expect(screen.byLabel(`Scope, ${recordScope === 'tenant' ? 'Home' : 'Household'}`)).toBeDefined();
    expect(screen.byLabel(`Scope, ${recordScope === 'tenant' ? 'Household' : 'Home'}`)).toBeUndefined();
  });

  it.each([true, false])('shows expiration tracking as a static inherited value: %s', async enabled => {
    const record = { ...assetType('shared-type', 'Medicine', 'tenant'), expirationEnabled: enabled };
    const screen = await renderEditor({
      kind: 'asset-type', mode: 'edit', resourceId: record.id,
      query: collectionQuery({ assetTypes: [record] })
    });
    expect(screen.byLabel('Track expiration dates')).toBeUndefined();
    expect(screen.allText()).toContain('Track expiration dates');
    expect(screen.allText()).toContain(enabled ? 'Enabled' : 'Disabled');
    expect(screen.allText()).not.toContain('Save');
  });

  it('shares a pending definition list and selects the latest editor route', async () => {
    const first = deferred<{ items: readonly ReturnType<typeof field>[]; complete: true }>();
    const resourceA = field('field-a', 'Resource A', 'inventory');
    const resourceB = field('field-b', 'Resource B', 'inventory');
    let calls = 0;
    const query = {
      assetTypes: async () => ({ items: [], complete: true }),
      fields: async () => { calls++; return first.promise; }
    };
    harness = new MobileRenderHarness();
    await harness.render(editorElement({ kind: 'field', mode: 'edit', query, resourceId: resourceA.id }));
    await harness.render(editorElement({ kind: 'field', mode: 'edit', query, resourceId: resourceB.id }));
    await settleQueries(harness);
    expect(harness.byLabel('Name')).toBeUndefined();
    first.resolve({ items: [resourceA, resourceB], complete: true });
    await settleQueries(harness);
    expect(harness.byLabel('Name')?.props.value).toBe('Resource B');
    expect(calls).toBe(1);
  });

  it('lets read-only viewers inspect local detail without exposing controls', async () => {
    const local = field('local-field', 'Local field', 'inventory');
    const screen = await renderEditor({ contextQuery: { execute: async () => context(['view'], ['view']) }, kind: 'field', mode: 'edit', query: collectionQuery({ fields: [local] }), resourceId: local.id });
    expect(screen.allText()).toContain('Local field');
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).not.toContain('Save');
    expect(screen.allText()).not.toContain('Archive');
  });

  it('renders immutable field type and existing targets as static values', async () => {
    const type = assetType('type-1', 'Appliance', 'inventory');
    const record = { ...field('field-1', 'Priority', 'inventory'), type: 'enum' as const, enumOptions: ['high'], applicability: 'custom_asset_types' as const, customAssetTypeIds: [type.id] };
    const screen = await renderEditor({ kind: 'field', mode: 'edit', query: collectionQuery({ fields: [record], assetTypes: [type] }), resourceId: record.id });
    expect(screen.allText()).toEqual(expect.arrayContaining(['Enum', 'Appliance · Existing']));
    expect(screen.byLabel('Choose Applies to. Current value Selected asset types')).toBeDefined();
    expect(screen.byText('Enum')?.parent?.type).not.toBe('Pressable');
  });

  it('reverses an unsaved applicability expansion without losing target or name edits', async () => {
    const saved = assetType('saved', 'Appliance', 'inventory');
    const added = assetType('added', 'Furniture', 'inventory');
    const record = { ...field('priority', 'Priority', 'inventory'), applicability: 'custom_asset_types' as const, customAssetTypeIds: [saved.id] };
    const screen = await renderEditor({ kind: 'field', mode: 'edit', resourceId: record.id, query: collectionQuery({ fields: [record], assetTypes: [saved, added] }) });
    await screen.changeText(screen.byLabel('Name'), 'My priority');
    await screen.press(screen.byLabel('Furniture'));
    await screen.press(screen.byLabel('Choose Applies to. Current value Selected asset types'));
    await screen.press(screen.byLabel('All assets'));
    expect(screen.allText()).not.toContain('Appliance · Existing');
    await screen.press(screen.byLabel('Choose Applies to. Current value All assets'));
    await screen.press(screen.byLabel('Selected asset types'));
    expect(screen.byLabel('Name')?.props.value).toBe('My priority');
    expect(screen.allText()).toContain('Appliance · Existing');
    expect(screen.byLabel('Furniture')?.props.accessibilityState.checked).toBe(true);
    expect(screen.byLabel('Save')?.props.disabled).toBe(false);
  });

  it('does not offer narrowing for a saved all-assets field', async () => {
    const record = field('all', 'All', 'inventory');
    const screen = await renderEditor({ kind: 'field', mode: 'edit', resourceId: record.id, query: collectionQuery({ fields: [record] }) });
    expect(screen.allText()).toContain('All assets');
    expect(screen.byLabel('Choose Applies to. Current value All assets')).toBeUndefined();
  });

  it('reports unavailable existing targets without leaking their identifiers', async () => {
    const record = { ...field('field-1', 'Priority', 'inventory'), applicability: 'custom_asset_types' as const, customAssetTypeIds: ['hidden-type-id'] };
    const screen = await renderEditor({ kind: 'field', mode: 'edit', query: collectionQuery({ fields: [record] }), resourceId: record.id });
    expect(screen.allText()).toContain('1 existing asset type is unavailable');
    expect(JSON.stringify(screen.all().map((node) => ({ label: node.props.accessibilityLabel, type: node.type, value: node.props.value })))).not.toContain('hidden-type-id');
  });

  it('renders honest empty, incomplete, and filtered collection states', async () => {
    let screen = await renderCollection();
    expect(screen.allText()).toEqual(expect.arrayContaining(['No active tags', 'Add the first tag here.']));
    await harness?.unmount(); harness = undefined;
    screen = await renderCollection({ query: { tags: async () => ({ items: [tag('tools', 'Tools')], complete: false }) } });
    expect(screen.allText()).toContain('Some settings may be missing');
    await screen.run(() => collectionSearch().onChangeText({ nativeEvent: { text: 'missing' } }));
    expect(screen.allText()).toContain('No matches');
    expect(screen.allText()).toContain('No tags match “missing”.');
  });

  it('keeps colored and uncolored tag rows aligned with color described accessibly', async () => {
    const screen = await renderCollection({ query: collectionQuery({ tags: [{ ...tag('blue', 'Blue tag'), color: '#2F80ED' }, tag('none', 'No-color tag')] }) });
    expect(screen.byLabel('Blue tag, Blue')).toBeDefined();
    expect(screen.byLabel('No-color tag, No color')).toBeDefined();
    const slots = screen.allByType('View').filter(hasStyle({ width: 24, height: 24 }));
    expect(slots).toHaveLength(2);
  });

  it('turns a load-time permission failure into an explicit denied state', async () => {
    const contexts = sequence(context(['configure'], ['view', 'edit_asset']), context([], []));
    const screen = await renderEditor({ contextQuery: { execute: contexts.next }, mode: 'edit', query: { tags: async () => { throw new CustomizationFailure('permission-denied'); } }, resourceId: 'tag-1' });
    expect(screen.allText()).toEqual(expect.arrayContaining(['Settings unavailable', 'Your access changed. This setting can’t be shown.']));
  });

  it('saves edits through update rather than create', async () => {
    const record = tag('tag-1', 'Tools');
    let creates = 0; let updates = 0;
    const manager = managerFake({ create: async () => { creates += 1; return record; }, update: async () => { updates += 1; return record; } });
    const screen = await renderEditor({ manageTags: manager, mode: 'edit', query: collectionQuery({ tags: [record] }), resourceId: record.id });
    await screen.changeText(screen.byLabel('Name'), 'Workshop tools');
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    expect({ creates, updates }).toEqual({ creates: 0, updates: 1 });
  });

  it('never renders restore or permanent-delete controls for archived tags', async () => {
    const record = tag('tag-1', 'Tools');
    const screen = await renderEditor({ kind: 'tag', lifecycle: 'archived', mode: 'edit', query: collectionQuery({ tags: [record] }), resourceId: record.id });
    expect(screen.allText()).not.toContain('Restore');
    expect(screen.allText()).not.toContain('Delete permanently');
  });

  it('rejects illegal tag restore and delete lifecycle commands without archiving', async () => {
    let archives = 0;
    const managers = { assetTypes: managerFake(), fields: managerFake(), tags: managerFake({ archive: async () => { archives += 1; } }) };
    for (const action of ['restore', 'delete'] as const) {
      await expect(runCustomizationLifecycle({ action, context: context(['configure'], ['view', 'edit_asset']), kind: 'tag', managers, resourceId: 'tag-1', scope: 'inventory' } as never)).rejects.toThrow(/archive only/i);
    }
    expect(archives).toBe(0);
  });

  it('preserves an editable record and focuses the inline error after lifecycle failure', async () => {
    const record = field('field-1', 'Priority', 'inventory');
    const manager = managerFake({ archive: async () => { throw new CustomizationFailure('conflict'); } });
    const screen = await renderEditor({ kind: 'field', manageFields: manager, mode: 'edit', query: collectionQuery({ fields: [record] }), resourceId: record.id });
    await screen.press(screen.byText('Archive')?.parent ?? undefined);
    await screen.run(() => pressAlertButton('Archive')); await settleQueries(screen);
    expect(screen.byLabel('Name')?.props.value).toBe('Priority');
    expect(screen.allText()).toContain('Could not archive');
    expect(focusedAccessibilityHandles()).toContain(1);
    expect(screen.allText()).toContain('Archive');
  });

  it('uses the linked platform-native segmented control without redundant selected checks', async () => {
    const changes: string[] = [];
    harness = new MobileRenderHarness();
    await harness.render(<SettingsSegmentedControl onChange={(value) => changes.push(value)} segments={[{ label: 'Active', value: 'active' }, { label: 'Archived', value: 'archived' }]} value="active" />);
    expect(harness.byType('NativeSegmentedControl')?.props).toMatchObject({ selectedIndex: 0, values: ['Active', 'Archived'] });
    expect(harness.allText()).not.toContain('Selected');
    await harness.change(harness.byType('NativeSegmentedControl'), 'Archived');
    expect(changes).toEqual(['archived']);
  });

  it('keeps prior rows with an inline loading row until an atomic lifecycle switch commits', async () => {
    const archived = deferred<{ items: readonly ReturnType<typeof field>[]; complete: true }>();
    const fields = sequence({ items: [field('active', 'Active field', 'inventory')], complete: true }, archived.promise);
    const screen = await renderCollection({ kind: 'field', query: { fields: fields.next }, scope: 'inventory' });
    await screen.change(screen.byType('NativeSegmentedControl'), 'Archived');
    expect(screen.allText()).toEqual(expect.arrayContaining(['Active field', 'Loading archived settings…']));
    expect(screen.all().find((node) => node.props.accessibilityRole === 'progressbar')).toBeDefined();
    expect(screen.allText()).not.toContain('No archived custom fields');

    archived.resolve({ items: [field('archived', 'Archived field', 'inventory')], complete: true });
    await settleQueries(screen);
    expect(screen.allText()).toContain('Archived field');
    expect(screen.allText()).not.toContain('Active field');
  });

  it('rolls a failed lifecycle switch back to the prior selection and rows', async () => {
    const fields = sequence<{ items: readonly ReturnType<typeof field>[]; complete: true }>({ items: [field('active', 'Active field', 'inventory')], complete: true }, new CustomizationFailure('conflict'));
    const screen = await renderCollection({ kind: 'field', query: { fields: fields.next }, scope: 'inventory' });
    await screen.change(screen.byType('NativeSegmentedControl'), 'Archived');
    await settleQueries(screen);
    expect(screen.allText()).toContain('Active field');
    expect(screen.allText()).not.toContain('Archived field');
    expect(screen.byType('NativeSegmentedControl')?.props.selectedIndex).toBe(0);
  });

  it('keeps a denied-save draft visible, read-only, and recoverable after access returns', async () => {
    const contexts = sequence(context(['configure'], ['view', 'edit_asset']), context([], []), context(['configure'], ['view', 'edit_asset']));
    const manager = managerFake({ create: async () => { throw new CustomizationFailure('permission-denied'); } });
    const screen = await renderEditor({ contextQuery: { execute: contexts.next }, manageTags: manager });
    await screen.changeText(screen.byLabel('Name'), 'Garage');
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    await settleQueries(screen);
    expect(screen.allText()).toEqual(expect.arrayContaining(['Access changed', 'Garage', 'Refresh access']));
    expect(screen.byLabel('Name')).toBeUndefined();
    expect(screen.allText()).toContain('Garage');
    await screen.press(screen.byText('Refresh access')?.parent ?? undefined);
    expect(screen.byLabel('Name')?.props.editable).toBe(true);
  });

  it('saves through each domain manager and leaves without a discard prompt', async () => {
    for (const kind of ['tag', 'field', 'asset-type'] as const) {
      await harness?.unmount(); harness = undefined; resetNativeTestState(); resetNavigation();
      const calls: unknown[][] = [];
      const manager = managerFake({ create: async (...args: unknown[]) => { calls.push(args); return {}; } });
      let done = 0;
      const screen = await renderEditor({ kind, manageAssetTypes: manager, manageFields: manager, manageTags: manager, onDone: () => { done += 1; } });
      await screen.changeText(screen.byLabel('Name'), kind === 'tag' ? 'Tools' : 'Appliance');
      if (kind === 'asset-type') await screen.run(() => screen.byLabel('Track expiration dates')?.props.onValueChange(true));
      await screen.press(screen.byText('Save')?.parent ?? undefined);
      expect(calls).toHaveLength(1);
      if (kind === 'asset-type') expect(calls[0][2]).toMatchObject({ expirationEnabled: true });
      expect(done).toBe(1);
      expect(alertCount()).toBe(0);
    }
  });

  it('keeps dirty navigation in place, disables gestures, and dispatches discard exactly once', async () => {
    const action = { type: 'RETURN_COLLECTION' };
    const screen = await renderEditor({ onDone: () => attemptNavigation(action) });
    await screen.changeText(screen.byLabel('Name'), 'Populated draft');
    expect(navigationOptions().at(-1)).toMatchObject({ gestureEnabled: false, headerBackVisible: false });
    const header = new MobileRenderHarness();
    const options = navigationOptions().at(-1) as { headerLeft: () => React.ReactElement };
    await header.render(options.headerLeft());
    try {
      await header.press(header.byLabel('Back to settings collection'));
      expect(latestAlert()).toMatchObject({ title: 'Discard changes?', message: 'Your unsaved changes will be lost.' });
      await pressAlertButton('Keep Editing'); await settleQueries(screen);
      expect(dispatchedActions()).toEqual([]);
      expect(screen.byLabel('Name')?.props.value).toBe('Populated draft');
      await header.press(header.byLabel('Back to settings collection'));
      await pressAlertButton('Discard'); await pressAlertButton('Discard'); await settleQueries(screen);
      expect(dispatchedActions()).toEqual([action]);
    } finally { await header.unmount(); }
  });

  it.each([false, true])('ignores a discard confirmation after leaving, returned=%s', async returned => {
    const screen = await renderEditor();
    await screen.changeText(screen.byLabel('Name'), 'Keep this draft');
    attemptNavigation({ type: 'OLD_BACK' });
    await screen.run(() => setScreenFocused(false));
    if (returned) await screen.run(() => setScreenFocused(true));
    await pressAlertButton('Discard'); await settleQueries(screen);
    expect(dispatchedActions()).toEqual([]);
    expect(screen.byLabel('Name')?.props.value).toBe('Keep this draft');
    if (!returned) await screen.run(() => setScreenFocused(true));
    const action = { type: 'FRESH_BACK' };
    attemptNavigation(action);
    await pressAlertButton('Discard'); await settleQueries(screen);
    expect(dispatchedActions()).toEqual([action]);
  });

  it('ignores discard from a replaced definition and preserves its new draft', async () => {
    const first = field('first', 'First', 'inventory');
    const second = field('second', 'Second', 'inventory');
    const query = collectionQuery({ fields: [first, second] });
    const screen = await renderEditor({ kind: 'field', mode: 'edit', query, resourceId: first.id });
    await screen.changeText(screen.byLabel('Name'), 'First draft');
    attemptNavigation({ type: 'OLD_BACK' });
    await screen.render(editorElement({ kind: 'field', mode: 'edit', query, resourceId: second.id }));
    await settleQueries(screen);
    await screen.changeText(screen.byLabel('Name'), 'Second draft');
    await pressAlertButton('Discard'); await settleQueries(screen);
    expect(dispatchedActions()).toEqual([]);
    expect(screen.byLabel('Name')?.props.value).toBe('Second draft');
  });

  it.each(['tag', 'field', 'asset-type'] as const)('disables %s draft controls while saving without removing the fields', async kind => {
    const pending = deferred<Record<string, unknown>>();
    const manager = managerFake({ create: async () => { await pending.promise; throw new Error('Offline'); } });
    const screen = await renderEditor({ kind, manageFields: manager, manageTags: manager, manageAssetTypes: manager });
    await screen.changeText(screen.byLabel('Name'), 'Reviewed draft');
    if (kind !== 'tag') await screen.press(screen.byText('Show technical details')?.parent ?? undefined);
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    expect(screen.byLabel('Name')?.props.value).toBe('Reviewed draft');
    expect(screen.byLabel('Name')?.props.editable).toBe(false);
    if (kind !== 'tag') expect(screen.byLabel('Stable key')?.props.editable).toBe(false);
    if (kind === 'tag') expect(screen.byLabel('Choose a custom tag color')?.props.disabled).toBe(true);
    if (kind === 'field') expect(screen.byLabel('Choose Type. Current value Text')?.props.disabled).toBe(true);
    if (kind === 'asset-type') expect(screen.byLabel('Description')?.props.editable).toBe(false);
    pending.resolve({}); await settleQueries(screen);
    expect(screen.byLabel('Name')?.props.editable).toBe(true);
    expect(screen.byLabel('Name')?.props.value).toBe('Reviewed draft');
  });

  it('freezes an open color draft during save and resumes it after failure', async () => {
    const pending = deferred<Record<string, unknown>>();
    const manager = managerFake({ create: async () => { await pending.promise; throw new Error('Offline'); } });
    const screen = await renderEditor({ manageTags: manager });
    await screen.changeText(screen.byLabel('Name'), 'Tools');
    await screen.press(screen.byLabel('Choose a custom tag color'));
    await screen.changeText(screen.byLabel('Custom tag color hex value'), '#123456');
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    expect(screen.byText('Done')?.parent?.props.disabled).toBe(true);
    await screen.press(screen.byText('Done')?.parent ?? undefined);
    expect(screen.byLabel('No tag color')?.props.accessibilityState.selected).toBe(true);
    pending.resolve({}); await settleQueries(screen);
    expect(screen.byText('Done')?.parent?.props.disabled).toBe(false);
    await screen.press(screen.byText('Done')?.parent ?? undefined);
    expect(screen.byLabel('Choose a custom tag color')?.props.accessibilityState.selected).toBe(true);
  });

  it('keeps lifecycle mutations single-flight through confirmation and network completion', async () => {
    const pending = deferred<ReturnType<typeof field>>();
    let archives = 0;
    const manager = managerFake({ archive: async () => { archives += 1; return pending.promise; } });
    const record = field('field-1', 'Priority', 'inventory');
    const screen = await renderEditor({ kind: 'field', manageFields: manager, mode: 'edit', query: collectionQuery({ fields: [record] }), resourceId: record.id });
    await screen.press(screen.byText('Archive')?.parent ?? undefined);
    expect(alertCount()).toBe(1);
    const mutation = pressAlertButton('Archive');
    await Promise.resolve();
    await pressAlertButton('Archive');
    expect(archives).toBe(1);
    expect(screen.allText()).toContain('Working…');
    pending.resolve(record); await mutation; await settleQueries(screen);
  });

  it('focuses and announces direct denied settings states', async () => {
    harness = new MobileRenderHarness();
    await harness.render(<DeniedSettingsState message="You cannot view this setting." />);
    expect(harness.all().find((node) => node.props.accessibilityLiveRegion === 'assertive')).toBeDefined();
    expect(focusedAccessibilityHandles()).toEqual([1]);
  });

  it('renders the binding household/inventory role hierarchy', async () => {
    const value = settings(['view'], ['view']);
    const readOnly = new SettingsQuery({ getCurrentPrincipal: async () => ({ id: 'principal', email: 'person@example.test' }) }, { getDiagnostics: () => ({ apiBaseUrl: 'https://example.test', appVersion: 'test', authenticationMode: 'oidc-sso' }) }, { getSelectedScope: async () => ({ tenant: value.selectedTenant, inventory: value.selectedInventory }) });
    const client = createMobileQueryClient();
    const wrap = (child: React.ReactNode) => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant-1', inventoryId: 'inventory-1' })}>{child}</MobileServerStateProvider>;
    harness = new MobileRenderHarness();
    await harness.render(wrap(<InventorySettingsScreen onNavigate={() => undefined} settingsQuery={readOnly} />)); await harness.run(() => new Promise((resolve) => setTimeout(resolve, 20)));
    expect(harness.allText()).toEqual(expect.arrayContaining(['Tags', 'Custom fields', 'Asset types', 'Notifications']));
    expect(harness.allText()).not.toContain('Sharing');

    await harness.render(wrap(<HouseholdSettingsScreen onNavigate={() => undefined} settingsQuery={readOnly} />)); await settleQueries(harness);
    expect(harness.allText()).toContain('Settings unavailable');
  });
});

async function renderCollection(overrides: Record<string, unknown> = {}) {
  harness = new MobileRenderHarness();
  const props = {
    accessPolicy: policy, contextQuery: { execute: async () => context(['configure'], ['view', 'configure', 'edit_asset']) }, kind: 'tag',
    onAdd: () => undefined, onOpen: () => undefined, query: collectionQuery(), scope: 'inventory', ...overrides
  } as unknown as React.ComponentProps<typeof CustomizationCollectionScreen>;
  await harness.render(withQueries(<AppFeedbackProvider><CustomizationCollectionScreen {...props} /></AppFeedbackProvider>));
  await settleQueries(harness); return harness;
}

async function renderEditor(overrides: Record<string, unknown> = {}) {
  harness = new MobileRenderHarness();
  await harness.render(editorElement(overrides));
  await settleQueries(harness); return harness;
}

function editorElement(overrides: Record<string, unknown> = {}) {
  const inert = managerFake();
  const props = {
    accessPolicy: policy, contextQuery: { execute: async () => context(['configure'], ['view', 'configure', 'edit_asset']) }, inherited: false, kind: 'tag', lifecycle: 'active',
    manageAssetTypes: inert, manageFields: inert, manageTags: inert, mode: 'create', onDone: () => undefined, query: collectionQuery(), scope: 'inventory', ...overrides
  } as unknown as React.ComponentProps<typeof CustomizationEditorScreen>;
  return withQueries(<AppFeedbackProvider><CustomizationEditorScreen {...props} /></AppFeedbackProvider>);
}

function collectionSearch() {
  const options = navigationOptions().filter(value => Object.hasOwn(value as object, 'headerSearchBarOptions')).at(-1) as {
    headerSearchBarOptions: { onChangeText: (event: { nativeEvent: { text: string } }) => void; onCancelButtonPress: () => void }
  };
  return options?.headerSearchBarOptions;
}

function collectionQuery(values: { tags?: readonly Record<string, unknown>[]; fields?: readonly Record<string, unknown>[]; assetTypes?: readonly Record<string, unknown>[] } = {}) {
  return {
    tags: async () => ({ items: values.tags ?? [], complete: true }),
    fields: async () => ({ items: values.fields ?? [], complete: true }),
    assetTypes: async () => ({ items: values.assetTypes ?? [], complete: true })
  };
}

function managerFake(overrides: Record<string, (...args: never[]) => unknown> = {}) {
  const success = async () => ({});
  return new Proxy({}, { get: (_target, key) => overrides[String(key)] ?? success });
}

const policy = {
  canRead: (value: ReturnType<typeof context>, scope: string) => scope === 'tenant' ? value.tenantPermissions.includes('configure') : value.inventoryPermissions.includes('view'),
  canMutate: (value: ReturnType<typeof context>, kind: string, scope: string, inherited = false) => !inherited && (kind === 'tag' ? value.inventoryPermissions.includes('edit_asset') : scope === 'tenant' ? value.tenantPermissions.includes('configure') : value.inventoryPermissions.includes('configure')),
  readOrRecord(value: ReturnType<typeof context>, _kind: string, scope: string) { return this.canRead(value, scope); },
  mutationOrRecord(value: ReturnType<typeof context>, kind: string, scope: string, inherited = false) { return this.canMutate(value, kind, scope, inherited); }
};
function context(tenantPermissions: readonly string[], inventoryPermissions: readonly string[]) { return { tenantId: 'tenant-1', tenantName: 'Home', tenantPermissions, inventoryId: 'inventory-1', inventoryName: 'Household', inventoryPermissions }; }
function tag(id: string, displayName: string) { return { kind: 'tag', id, key: id, displayName } as const; }
function field(id: string, displayName: string, scope: 'tenant' | 'inventory') { return { kind: 'field', id, tenantId: 'tenant-1', inventoryId: scope === 'inventory' ? 'inventory-1' : undefined, scope, key: id, displayName, type: 'text', enumOptions: [], applicability: 'all_assets', customAssetTypeIds: [], lifecycle: 'active' } as const; }
function assetType(id: string, displayName: string, scope: 'tenant' | 'inventory') { return { kind: 'asset-type', id, tenantId: 'tenant-1', inventoryId: scope === 'inventory' ? 'inventory-1' : undefined, scope, key: id, displayName, description: '', lifecycle: 'active' } as const; }
function settings(tenantPermissions: readonly string[], inventoryPermissions: readonly string[]) { return { selectedTenant: { id: 'tenant-1', name: 'Home', permissions: tenantPermissions }, selectedInventory: { id: 'inventory-1', name: 'Household', permissions: inventoryPermissions } }; }
function hasStyle(expected: Record<string, unknown>) { return (node: TestInstance) => { const styles = Array.isArray(node.props.style) ? node.props.style : [node.props.style]; return styles.some((style) => style && Object.entries(expected).every(([key, value]) => style[key] === value)); }; }
function sequence<T>(...values: Array<T | Error | Promise<T>>) { let index = 0; return { next: async () => { const value = values[Math.min(index++, values.length - 1)]; if (value instanceof Error) throw value; return await value; } }; }
function deferred<T>() { let resolve!: (value: T) => void; const promise = new Promise<T>((done) => { resolve = done; }); return { promise, resolve }; }

function withQueries(child: React.ReactNode) {
  return <MobileServerStateProvider client={queryClient} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant-1', inventoryId: 'inventory-1' })}>{child}</MobileServerStateProvider>;
}
async function settleQueries(harness: MobileRenderHarness) {
  await harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
  await harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
}

it('loads enabled expiration tracking and saves an explicit disabled value', async () => {
  const record = { ...assetType('medicine', 'Medicine', 'inventory'), expirationEnabled: true };
  const saved: unknown[][] = [];
  const manager = managerFake({ update: async (...args: unknown[]) => { saved.push(args); return record; } });
  const screen = await renderEditor({ kind: 'asset-type', mode: 'edit', resourceId: record.id, manageAssetTypes: manager, query: collectionQuery({ assetTypes: [record] }) });
  expect(screen.byLabel('Track expiration dates')?.props.value).toBe(true);
  await screen.run(() => screen.byLabel('Track expiration dates')?.props.onValueChange(false));
  await screen.press(screen.byText('Save')?.parent ?? undefined);
  expect(saved).toHaveLength(1);
  expect(saved[0][1]).toMatchObject({ expirationEnabled: false });
});
