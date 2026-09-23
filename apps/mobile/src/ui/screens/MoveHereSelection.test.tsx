import { afterEach, beforeEach, expect, it } from 'vitest';
import { Platform } from 'react-native';
import { NativeSearchDriver } from '../../test-support/NativeSearchDriver';
let search: NativeSearchDriver;
beforeEach(() => { search = new NativeSearchDriver(); });
afterEach(() => search.dispose());
import { AssetMoveHereSheetRouteScreen } from './AssetNativeActionSheetScreens';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { latestAlert } from '../../test-support/react-native';
const settle = async (h: MobileRenderHarness, ms = 30) => { await h.run(() => new Promise(resolve => setTimeout(resolve, ms))); };
it.each(['ios', 'android'] as const)('keeps Move Here selection through search and failed submission on %s', async platform => {
  const originalPlatform = Platform.OS; Platform.OS = platform;
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const changeSearch = (text: string) => platform === 'ios' ? h.run(() => search.change(text)) : h.changeText(h.byLabel('Search your inventory'), text);
  const asset = { id: assetId('garage'), title: 'Garage', description: '', kind: 'location' as const, lifecycleState: 'active' as const, locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['edit_asset'], revision: '1', asset }) });
  const submitted: unknown[] = [];
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetMoveHereSheetRouteScreen assetId="garage" assetCoreQuery={core}
        parentLookupQuery={{ execute: async query => query ? [] : [{ id: 'tent', title: 'Tent', kind: 'item', subtitle: '', pathLabel: 'Attic / Tent', selectionHint: 'Item', willPromoteToContainer: false }] }}
        moveAssetCommand={{ execute: async input => { submitted.push(input); throw new Error('Keep selection for retry'); } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    expect(h.byText('Destination: Garage')).toBeDefined();
    const choice = h.byLabel('Choose item Tent');
    expect(choice?.props.accessibilityRole).toBe('radio');
    expect(choice?.props.accessibilityState.checked).toBe(false);
    expect(h.byText('Item · Attic / Tent')).toBeDefined();
    await h.press(choice);
    expect(h.byLabel('Choose item Tent')?.props.accessibilityState.checked).toBe(true);
    expect(submitted).toEqual([]);
    await changeSearch('unmatched'); await settle(h, 350); await settle(h);
    expect(h.byLabel('Choose item Tent')).toBeUndefined();
    expect(h.byText('Selected: Tent')).toBeDefined();
    expect(h.byLabel('Move here')?.props.disabled).toBe(false);
    await changeSearch(''); await settle(h, 350); await settle(h);
    expect(h.byLabel('Choose item Tent')?.props.accessibilityState.checked).toBe(true);
    await h.press(h.byLabel('Move here'));
    expect(submitted).toEqual([{ assetId: 'tent', parentAssetId: 'garage' }]);
    expect(h.byLabel('Choose item Tent')?.props.accessibilityState.checked).toBe(true);
    expect(latestAlert()?.title).toBe('Could not move asset here');
  } finally { await h.unmount(); client.clear(); Platform.OS = originalPlatform; }
});
