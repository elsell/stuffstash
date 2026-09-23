import { expect, it } from 'vitest';
import { AssetMoveSheetRouteScreen } from './AssetNativeActionSheetScreens';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { latestAlert } from '../../test-support/react-native';
const settle = async (h: MobileRenderHarness, ms = 30) => { await h.run(() => new Promise(resolve => setTimeout(resolve, ms))); };
it('edits a creation name without collapsing Kind or changing search, validates that name, and retries it', async () => {
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
    await h.changeText(h.byLabel('Put in'), 'Garden'); await settle(h, 350); await settle(h);
    await h.press(h.byLabel('New destination'));
    await settle(h, 350); await settle(h);
    const retainedCreate = h.byLabel('Create location "Garden"')!.props.onPress;
    await h.press(h.byLabel('Cancel new destination'));
    await h.run(() => retainedCreate());
    expect(submitted).toEqual([]);
    await h.press(h.byLabel('New destination'));
    await h.changeText(h.byLabel('New destination name'), 'Garden shed');
    await h.run(() => retainedCreate());
    expect(submitted).toEqual([]);
    expect(h.byLabel('Choose destination kind')).toBeDefined();
    expect(h.byLabel('Put in')?.props.value).toBe('Garden');
    expect(h.byLabel('Create location "Garden shed"')?.props.disabled).toBe(true);
    await settle(h, 350); await settle(h);
    expect(lookups).toContain('Garden shed');
    await h.run(() => resolveName([])); await settle(h);
    await h.run(() => retainedCreate());
    expect(latestAlert()?.title).toBe('Could not create destination');
    expect(h.byLabel('New destination name')?.props.value).toBe('Garden shed');
    expect(h.byLabel('Choose destination kind')).toBeDefined();
    await h.press(h.byLabel('Create location "Garden shed"'));
    expect(submitted).toEqual([expect.objectContaining({ title: 'Garden shed' }), expect.objectContaining({ title: 'Garden shed' })]);
    expect(h.byLabel('New destination name')).toBeUndefined();
    expect(h.byText('Selected: Garden shed')).toBeDefined();
    await h.run(() => retainedCreate());
    expect(submitted).toHaveLength(2);

  } finally { await h.unmount(); client.clear(); }
});
