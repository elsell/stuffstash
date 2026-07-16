import React from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import type { TestInstance } from 'test-renderer';
import { CustomizationFailure } from '../../application/customization/CustomizationErrors';
import { runCustomizationLifecycle } from '../../application/customization/CustomizationEditorCommands';
import { MobileRenderHarness } from '../../test-support/render';
import { alertCount, focusedAccessibilityHandles, focusedInputLabels, latestAlert, pressAlertButton, resetNativeTestState } from '../../test-support/react-native';
import { attemptNavigation, dispatchedActions, navigationOptions, resetNavigation } from '../../test-support/navigation';
import { SettingsSegmentedControl } from '../components/SettingsSegmentedControl';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { CustomizationCollectionScreen } from './CustomizationCollectionScreen';
import { CustomizationEditorScreen } from './CustomizationEditorScreen';
import { DeniedSettingsState, HouseholdSettingsScreen, InventorySettingsScreen } from './ScopedSettingsScreens';

let harness: MobileRenderHarness | undefined;
beforeEach(() => { resetNativeTestState(); resetNavigation(); Reflect.deleteProperty(globalThis, 'expo'); });
afterEach(async () => { await harness?.unmount(); harness = undefined; Reflect.deleteProperty(globalThis, 'expo'); });

describe('rendered mobile customization production states', () => {
  it('aligns collection chrome and editor actions to the shared 16-point content column', async () => {
    const collection = await renderCollection({ query: collectionQuery({ tags: [tag('tag-1', 'Tools')] }) });
    expect(collection.byLabel('Search Tags')?.props.style).toMatchObject({ minHeight: 44 });
    expect(collection.allByType('View').some(hasStyle({ marginHorizontal: 16 }))).toBe(true);

    const editor = await renderEditor();
    await editor.changeText(editor.byLabel('Name'), 'Tools');
    expect(editor.byText('Save')?.parent?.props.style).toEqual(expect.arrayContaining([expect.objectContaining({ marginHorizontal: 16 })]));
  });

  it('refreshes permissions and removes retained rows when a collection load is denied', async () => {
    const contexts = sequence(context(['configure'], ['view', 'edit_asset']), context([], []));
    const tags = sequence<{ items: readonly ReturnType<typeof tag>[]; complete: true }>({ items: [tag('private-tag', 'Private tag')], complete: true }, new CustomizationFailure('permission-denied'));
    const screen = await renderCollection({ contextQuery: { execute: contexts.next }, query: { tags: tags.next } });
    expect(screen.allText()).toContain('Private tag');
    const refresh = screen.byType('ScrollView')?.props.refreshControl as React.ReactElement<{ onRefresh: () => void }>;
    refresh.props.onRefresh();
    await screen.settle();
    expect(screen.allText()).toContain('Settings unavailable');
    expect(screen.allText()).not.toContain('Private tag');
  });

  it('renders inherited/local grouping and role-specific mutation controls', async () => {
    const allowed = context(['view'], ['view', 'configure']);
    const screen = await renderCollection({
      contextQuery: { execute: async () => allowed }, kind: 'field', scope: 'inventory',
      query: collectionQuery({ fields: [field('tenant-field', 'Shared field', 'tenant'), field('local-field', 'Local field', 'inventory')] })
    });
    expect(screen.allText()).toEqual(expect.arrayContaining(['From Home', 'Only in Household', 'Shared field', 'Local field', 'Add']));
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
    expect(screen.all().filter((node) => node.props.accessibilityRole === 'radio')).toHaveLength(0);
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

  it('prevents a stale resource load from overwriting a newer editor route', async () => {
    const first = deferred<{ items: readonly ReturnType<typeof field>[]; complete: true }>();
    const resourceA = field('field-a', 'Resource A', 'inventory');
    const resourceB = field('field-b', 'Resource B', 'inventory');
    let calls = 0;
    const query = {
      assetTypes: async () => ({ items: [], complete: true }),
      fields: async () => ++calls === 1 ? first.promise : { items: [resourceB], complete: true }
    };
    harness = new MobileRenderHarness();
    await harness.render(editorElement({ kind: 'field', mode: 'edit', query, resourceId: resourceA.id }));
    await harness.render(editorElement({ kind: 'field', mode: 'edit', query, resourceId: resourceB.id }));
    await harness.settle();
    expect(harness.byLabel('Name')?.props.value).toBe('Resource B');
    first.resolve({ items: [resourceA], complete: true });
    await harness.settle();
    expect(harness.byLabel('Name')?.props.value).toBe('Resource B');
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
    expect(screen.allText()).toEqual(expect.arrayContaining(['Enum', 'Appliance · Existing', 'Expand to all assets']));
    expect(screen.byText('Enum')?.parent?.type).not.toBe('Pressable');
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
    await screen.changeText(screen.byLabel('Search Tags'), 'missing');
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
    await screen.run(() => pressAlertButton('Archive')); await screen.settle();
    expect(screen.byLabel('Name')?.props.value).toBe('Priority');
    expect(screen.allText()).toContain('Could not archive');
    expect(focusedAccessibilityHandles()).toContain(1);
    expect(screen.allText()).toContain('Archive');
  });

  it('uses a native segmented control when available and a semantic fallback without redundant selected checks', async () => {
    const changes: string[] = [];
    Reflect.set(globalThis, 'expo', { getViewConfig: () => ({ validAttributes: {}, directEventTypes: {} }) });
    harness = new MobileRenderHarness();
    await harness.render(<SettingsSegmentedControl onChange={(value) => changes.push(value)} segments={[{ label: 'Active', value: 'active' }, { label: 'Archived', value: 'archived' }]} value="active" />);
    expect(harness.byType('NativeSegmentedControl')?.props).toMatchObject({ selectedIndex: 0, values: ['Active', 'Archived'] });
    expect(harness.allText()).not.toContain('Selected');
    await harness.change(harness.byType('NativeSegmentedControl'), 'Archived');
    expect(changes).toEqual(['archived']);

    await harness.unmount(); harness = new MobileRenderHarness();
    Reflect.deleteProperty(globalThis, 'expo');
    await harness.render(<SettingsSegmentedControl onChange={(value) => changes.push(value)} segments={[{ label: 'Active', value: 'active' }, { label: 'Archived', value: 'archived' }]} value="active" />);
    const tabs = harness.all().filter((node) => node.props.accessibilityRole === 'tab');
    expect(tabs).toHaveLength(2);
    expect(tabs[0].props.accessibilityState).toMatchObject({ selected: true });
    expect(harness.allText()).not.toContain('Selected');
  });

  it('keeps prior rows with an inline loading row until an atomic lifecycle switch commits', async () => {
    const archived = deferred<{ items: readonly ReturnType<typeof field>[]; complete: true }>();
    const fields = sequence({ items: [field('active', 'Active field', 'inventory')], complete: true }, archived.promise);
    const screen = await renderCollection({ kind: 'field', query: { fields: fields.next }, scope: 'inventory' });
    const archivedTab = screen.byText('Archived')?.parent ?? undefined;
    await screen.press(archivedTab);
    expect(screen.allText()).toEqual(expect.arrayContaining(['Active field', 'Loading archived settings…']));
    expect(screen.all().find((node) => node.props.accessibilityRole === 'progressbar')).toBeDefined();
    expect(screen.allText()).not.toContain('No archived custom fields');

    archived.resolve({ items: [field('archived', 'Archived field', 'inventory')], complete: true });
    await screen.settle();
    expect(screen.allText()).toContain('Archived field');
    expect(screen.allText()).not.toContain('Active field');
  });

  it('rolls a failed lifecycle switch back to the prior selection and rows', async () => {
    const fields = sequence<{ items: readonly ReturnType<typeof field>[]; complete: true }>({ items: [field('active', 'Active field', 'inventory')], complete: true }, new CustomizationFailure('conflict'));
    const screen = await renderCollection({ kind: 'field', query: { fields: fields.next }, scope: 'inventory' });
    const archivedTab = screen.byText('Archived')?.parent ?? undefined;
    await screen.press(archivedTab);
    await screen.settle();
    expect(screen.allText()).toContain('Active field');
    expect(screen.allText()).not.toContain('Archived field');
    const activeTab = screen.byText('Active')?.parent ?? undefined;
    expect(activeTab?.props.accessibilityState).toMatchObject({ selected: true });
  });

  it('keeps a denied-save draft visible, read-only, and recoverable after access returns', async () => {
    const contexts = sequence(context(['configure'], ['view', 'edit_asset']), context([], []), context(['configure'], ['view', 'edit_asset']));
    const manager = managerFake({ create: async () => { throw new CustomizationFailure('permission-denied'); } });
    const screen = await renderEditor({ contextQuery: { execute: contexts.next }, manageTags: manager });
    await screen.changeText(screen.byLabel('Name'), 'Garage');
    await screen.press(screen.byText('Save')?.parent ?? undefined);
    await screen.settle();
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
      await screen.press(screen.byText('Save')?.parent ?? undefined);
      expect(calls).toHaveLength(1);
      expect(done).toBe(1);
      expect(alertCount()).toBe(0);
    }
  });

  it('keeps dirty navigation in place, disables gestures, and dispatches discard exactly once', async () => {
    const screen = await renderEditor();
    await screen.changeText(screen.byLabel('Name'), 'Populated draft');
    expect(navigationOptions().at(-1)).toMatchObject({ gestureEnabled: false, headerBackVisible: false });
    const action = { type: 'GESTURE_BACK' };
    attemptNavigation(action);
    expect(latestAlert()).toMatchObject({ title: 'Discard changes?', message: 'Your unsaved changes will be lost.' });
    await pressAlertButton('Keep Editing'); await screen.settle();
    expect(dispatchedActions()).toEqual([]);
    expect(screen.byLabel('Name')?.props.value).toBe('Populated draft');
    attemptNavigation(action);
    await pressAlertButton('Discard'); await pressAlertButton('Discard'); await screen.settle();
    expect(dispatchedActions()).toEqual([action]);
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
    pending.resolve(record); await mutation; await screen.settle();
  });

  it('focuses and announces direct denied settings states', async () => {
    harness = new MobileRenderHarness();
    await harness.render(<DeniedSettingsState message="You cannot view this setting." />);
    expect(harness.all().find((node) => node.props.accessibilityLiveRegion === 'assertive')).toBeDefined();
    expect(focusedAccessibilityHandles()).toEqual([1]);
  });

  it('renders the binding household/inventory role hierarchy', async () => {
    const readOnly = { execute: async () => settings(['view'], ['view']) };
    harness = new MobileRenderHarness();
    await harness.render(<InventorySettingsScreen onNavigate={() => undefined} settingsQuery={readOnly as never} />); await harness.settle();
    expect(harness.allText()).toEqual(expect.arrayContaining(['Tags', 'Custom fields', 'Asset types']));
    expect(harness.allText()).not.toContain('Sharing');

    await harness.render(<HouseholdSettingsScreen onNavigate={() => undefined} settingsQuery={readOnly as never} />); await harness.settle();
    expect(harness.allText()).toContain('Settings unavailable');
  });
});

async function renderCollection(overrides: Record<string, unknown> = {}) {
  harness = new MobileRenderHarness();
  const props = {
    accessPolicy: policy, contextQuery: { execute: async () => context(['configure'], ['view', 'configure', 'edit_asset']) }, kind: 'tag',
    onAdd: () => undefined, onOpen: () => undefined, query: collectionQuery(), scope: 'inventory', ...overrides
  } as unknown as React.ComponentProps<typeof CustomizationCollectionScreen>;
  await harness.render(<AppFeedbackProvider><CustomizationCollectionScreen {...props} /></AppFeedbackProvider>);
  await harness.settle(); return harness;
}

async function renderEditor(overrides: Record<string, unknown> = {}) {
  harness = new MobileRenderHarness();
  await harness.render(editorElement(overrides));
  await harness.settle(); return harness;
}

function editorElement(overrides: Record<string, unknown> = {}) {
  const inert = managerFake();
  const props = {
    accessPolicy: policy, contextQuery: { execute: async () => context(['configure'], ['view', 'configure', 'edit_asset']) }, inherited: false, kind: 'tag', lifecycle: 'active',
    manageAssetTypes: inert, manageFields: inert, manageTags: inert, mode: 'create', onDone: () => undefined, query: collectionQuery(), scope: 'inventory', ...overrides
  } as unknown as React.ComponentProps<typeof CustomizationEditorScreen>;
  return <AppFeedbackProvider><CustomizationEditorScreen {...props} /></AppFeedbackProvider>;
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
