import { afterEach, beforeEach, expect, it } from 'vitest';
import { Platform } from 'react-native';
import { NativeSearchDriver } from '../../test-support/NativeSearchDriver';
let search: NativeSearchDriver;
beforeEach(() => { search = new NativeSearchDriver(); });
afterEach(() => search.dispose());
import { AssetMoveSheetRouteScreen } from './AssetNativeActionSheetScreens';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { navigationOptions, resetNavigation } from '../../test-support/navigation';
import { latestAlert } from '../../test-support/react-native';
const settle = async (h: MobileRenderHarness, ms = 30) => { await h.run(() => new Promise(resolve => setTimeout(resolve, ms))); };
it('edits a creation name without collapsing Kind or changing search, validates that name, and retries it', async () => {
  resetNavigation();
  const currentTitle = () => (Object.assign({}, ...navigationOptions()) as { title?: string }).title;
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  const submitted: unknown[] = []; const lookups: string[] = []; let resolveName!: (value: []) => void;
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core}
        parentLookupQuery={{ execute: async query => { lookups.push(query); if (query === 'Garden shed') return new Promise<[]>(resolve => { resolveName = resolve; }); return []; } }}
        createAssetCommand={{ execute: async input => { submitted.push(input); if (submitted.length === 1) throw new Error('Create unavailable'); return { id: 'shed', title: input.title, message: 'Created' }; } }}
        moveAssetCommand={{ execute: async () => { throw new Error('Choosing does not move'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.run(() => search.change('Garden')); await settle(h, 350); await settle(h);
    await h.press(h.byLabel('New destination'));
    await settle(h, 350); await settle(h);
    expect(h.byLabel('Move')).toBeUndefined();
    expect(h.byLabel('Choose inventory root')).toBeUndefined();
    const retainedCreate = h.byLabel('Create destination')!.props.onPress;
    await h.press(h.byLabel('Cancel new destination'));
    await h.run(() => retainedCreate());
    expect(submitted).toEqual([]);
    await h.press(h.byLabel('New destination'));
    await h.changeText(h.byLabel('New destination name'), 'Garden shed');
    await h.run(() => retainedCreate());
    expect(submitted).toEqual([]);
    expect(currentTitle()).toBe('New destination');
    expect(h.byLabel('Choose destination kind')).toBeDefined();
    expect(search.text).toBe('Garden');
    expect(h.byLabel('Put in')).toBeUndefined();
    expect(h.byLabel('Create destination')?.props.disabled).toBe(true);
    await settle(h, 350); await settle(h);
    expect(lookups).toContain('Garden shed');
    await h.run(() => resolveName([])); await settle(h);
    await h.run(() => retainedCreate());
    expect(latestAlert()?.title).toBe('Could not create destination');
    expect(h.byLabel('New destination name')?.props.value).toBe('Garden shed');
    expect(h.byLabel('Choose destination kind')).toBeDefined();
    await h.press(h.byLabel('Create destination'));
    expect(submitted).toEqual([expect.objectContaining({ title: 'Garden shed' }), expect.objectContaining({ title: 'Garden shed' })]);
    expect(h.byLabel('New destination name')).toBeUndefined();
    expect(h.byLabel('Choose destination Garden shed')?.props.accessibilityState.checked).toBe(true);
    await h.run(() => retainedCreate());
    expect(submitted).toHaveLength(2);

  } finally { await h.unmount(); client.clear(); }
});

it.each(['ios', 'android'] as const)('selects one destination on %s and moves only after confirmation', async platform => {
  const originalPlatform = Platform.OS; Platform.OS = platform;
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const changeSearch = (text: string) => platform === 'ios' ? h.run(() => search.change(text)) : h.changeText(h.byLabel('Search places, boxes, shelves'), text);
  const asset = { id: assetId('asset'), title: 'Tent', description: '', kind: 'item' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  const submitted: unknown[] = [];
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveSheetRouteScreen assetId="asset" assetCoreQuery={core}
        parentLookupQuery={{ execute: async query => query ? [] : [{ id: 'garage', title: 'Garage', kind: 'location', subtitle: '', pathLabel: 'House / Garage', selectionHint: 'Place', willPromoteToContainer: false }] }}
        createAssetCommand={{ execute: async () => { throw new Error('No creation requested'); } }}
        moveAssetCommand={{ execute: async input => { submitted.push(input); throw new Error('Keep selection for retry'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    const choice = h.byLabel('Choose destination Garage');
    expect(choice?.props.accessibilityRole).toBe('radio');
    expect(choice?.props.accessibilityState.checked).toBe(false);
    expect(h.byText('Location · House / Garage')).toBeDefined();
    const choose = choice!.props.onPress;
    await h.run(choose);
    expect(h.byLabel('Choose destination Garage')?.props.accessibilityState.checked).toBe(true);
    expect(h.byLabel('Choose inventory root')?.props.accessibilityState.checked).toBe(false);
    expect(submitted).toEqual([]);
    await changeSearch('unmatched'); await settle(h, 350);
    expect(h.byLabel('Choose destination Garage')?.props.accessibilityState.checked).toBe(true);
    expect(h.byText('Selected')).toBeDefined();
    await h.press(h.byLabel('Choose inventory root'));
    await h.run(choose);
    expect(h.byLabel('Choose inventory root')?.props.accessibilityState.checked).toBe(true);
    expect(submitted).toEqual([]);
    await changeSearch(''); await settle(h, 350); await settle(h);
    await h.press(h.byLabel('Choose destination Garage'));
    if (platform === 'ios') await h.run(() => search.options!.onClose());
    else await changeSearch('');
    await settle(h);
    expect(h.byLabel('Choose destination Garage')?.props.accessibilityState.checked).toBe(true);
    await h.press(h.byLabel('Move'));
    expect(submitted).toEqual([{ assetId: 'asset', parentAssetId: 'garage' }]);
    expect(h.byLabel('Choose destination Garage')?.props.accessibilityState.checked).toBe(true);
  } finally { await h.unmount(); client.clear(); Platform.OS = originalPlatform; }
});
