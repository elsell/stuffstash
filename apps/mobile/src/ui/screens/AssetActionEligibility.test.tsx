import React from 'react';
import { expect, it } from 'vitest';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation } from '../../test-support/navigation';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AssetEditSheetRouteScreen, AssetMoveSheetRouteScreen, AssetMoveHereSheetRouteScreen } from './AssetNativeActionSheetScreens';

const routes = ['edit', 'move', 'move-here'] as const;
const settle = (h: MobileRenderHarness) => h.run(() => new Promise(resolve => setTimeout(resolve, 25)));

it('does not submit Edit after permission changes during tag reconciliation', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation();
  let allowed = true; let reads = 0; let mutations = 0; let release!: () => void;
  const core = new AssetCoreQuery({ getAssetCore: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'),
    permissions: allowed ? ['edit_asset'] : ['view'], revision: String(allowed),
    asset: { id: assetId('asset'), title: 'Box', description: '', kind: 'container', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
  }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetEditSheetRouteScreen assetId="asset" assetCoreQuery={core} inventoryAssetTypesQuery={{ execute: async () => [] }}
        inventoryAssetTagsQuery={{ execute: async () => { if (++reads > 1) await new Promise<void>(resolve => { release = resolve; }); return []; } }}
        updateAssetCommand={{ execute: async () => { mutations++; return { id: 'asset', title: 'Box', message: 'Saved' }; } }} />
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    await h.changeText(h.byLabel('New tag name'), 'Retained tag'); await h.press(h.byLabel('Add tag'));
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetTags('scope', 'tenant', 'inventory'), refetchType: 'none' }));
    await h.press(h.byLabel('Save')); await settle(h);
    expect(release).toBeTypeOf('function');
    allowed = false;
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(h);
    await h.run(() => release()); await settle(h);
    expect(mutations).toBe(0);
    expect(h.byLabel('Remove new tag Retained tag')).toBeDefined();
    expect(h.byLabel('Save')?.props.disabled).toBe(true);
    expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
  } finally { await h.unmount(); client.clear(); resetNavigation(); }
});

for (const route of routes) {
  it.each([['revoked', false], ['archived', false], ['revoked', true], ['archived', true]] as const)(`${route} retains its draft but rejects stale actions when %s (denied entry=%s)`, async (restriction, deniedEntry) => {
    const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation();
    let restricted = deniedEntry; const mutations: string[] = [];
    const core = new AssetCoreQuery({ getAssetCore: async () => ({
      tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'),
      permissions: restricted && restriction === 'revoked' ? ['view'] : ['edit_asset', 'create_asset'], revision: String(restricted),
      asset: { id: assetId('asset'), title: 'Box', description: '', kind: 'container', lifecycleState: restricted && restriction === 'archived' ? 'archived' : 'active',
        locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false }
    }) });
    const lookup = { execute: async () => [{ id: 'candidate', title: 'Candidate', kind: 'container' as const, subtitle: '', pathLabel: 'Candidate', selectionHint: '', willPromoteToContainer: false }] };
    const move = { execute: async () => { mutations.push('move'); return { id: 'asset', title: 'Box', message: 'Moved' }; } };
    const shared = { assetId: 'asset', assetCoreQuery: core };
    const field = route === 'edit' ? 'Description' : route === 'move' ? 'Put in' : 'Find item, box, or place';
    const action = route === 'edit' ? 'Save' : route === 'move' ? 'Move' : 'Move here';
    try {
      await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        {route === 'edit' ? <AssetEditSheetRouteScreen {...shared} inventoryAssetTypesQuery={{ execute: async () => [] }} inventoryAssetTagsQuery={{ execute: async () => [] }}
          updateAssetCommand={{ execute: async () => { mutations.push('edit'); return { id: 'asset', title: 'Box', message: 'Saved' }; } }} />
          : route === 'move' ? <AssetMoveSheetRouteScreen {...shared} parentLookupQuery={lookup} moveAssetCommand={move}
            createAssetCommand={{ execute: async () => { mutations.push('create'); return { id: 'new', title: 'New', message: 'Created' }; } }} />
            : <AssetMoveHereSheetRouteScreen {...shared} parentLookupQuery={lookup} moveAssetCommand={move} />}
      </MobileServerStateProvider>);
      await settle(h); await settle(h);
      if (deniedEntry) {
        expect(h.byLabel(field)?.props.editable).toBe(false);
        expect(h.byLabel(action)?.props.disabled).toBe(true);
        expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
        expect(h.byText('This item cannot be changed here. Your draft is kept while this screen is open.')).toBeDefined();
        await h.press(h.byLabel(action)); expect(mutations).toEqual([]);
        restricted = false;
        await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(h);
      }
      await h.changeText(h.byLabel(field), 'Retained draft');
      if (route !== 'edit') {
        await h.run(() => new Promise(resolve => setTimeout(resolve, 300))); await settle(h);
        await h.press(h.byText('Candidate')?.parent?.parent?.parent ?? undefined);
      }
      expect(h.byLabel(action)?.props.disabled).toBe(false);
      const submit = h.byLabel(action)!.props.onPress;
      const edit = h.byLabel(field)!.props.onChangeText;
      restricted = true;
      await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(h);
      expect(h.byLabel(field)?.props.editable).toBe(false);
      expect(h.byLabel(action)?.props.disabled).toBe(true);
      expect(h.byLabel('Cancel')?.props.disabled).toBe(false);
      await h.run(() => { edit('Stale overwrite'); submit(); }); await settle(h);
      expect(mutations).toEqual([]);
      expect(h.byLabel(field)?.props.value).toBe('Retained draft');
      restricted = false;
      await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetCore('scope', 'tenant', 'inventory', 'asset') })); await settle(h);
      expect(h.byLabel(field)?.props.editable).toBe(true);
      await h.press(h.byLabel(action)); await settle(h);
      expect(mutations).toEqual([route === 'edit' ? 'edit' : 'move']);
    } finally { await h.unmount(); client.clear(); resetNavigation(); }
  });
}
