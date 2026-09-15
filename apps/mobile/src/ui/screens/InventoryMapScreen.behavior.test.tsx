import { animationStartCount, holdReduceMotionSnapshotForTest, resetNativeTestState, setReduceMotionEnabledForTest } from '../../test-support/react-native';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { InventoryMapScreen } from './InventoryMapScreen';
import { dispatchedActions, setScreenFocused } from '../../test-support/navigation';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { InventoryMapQuery } from '../../application/assets/InventoryMapQuery';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
const selectedAsset: AssetSummary = { id: assetId('tent'), title: 'Tent', kind: 'item', lifecycleState: 'active', description: '', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false };
const mapSnapshot = { sessionScopeId: 'scope', tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), inventoryName: 'Home', permissions: ['view', 'edit_asset'], assets: [selectedAsset] };

describe('Map server state', () => {
  it('performs one initial traversal and reuses fresh data on remount', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let calls = 0;
    const query = new InventoryMapQuery({ listActiveInventoryMapAssets: async () => {
      calls++;
      return mapSnapshot;
    } });
    const props: React.ComponentProps<typeof InventoryMapScreen> = {
      canAdd: false, inventoryMapQuery: query, pathStore: { current: new Map() }, selectedSurface: 'map', onAdd: () => undefined, onChangeSurface: () => undefined
    };
    const render = (visible: boolean) => harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider>{visible ? <InventoryMapScreen {...props} /> : null}</AppFeedbackProvider>
    </MobileServerStateProvider>);
    try {
      await render(true); await settle(harness); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(calls).toBe(1);
      await render(false); await render(true); await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(calls).toBe(1);
      await harness.press(harness.byLabel('Show details for Tent'));
      expect(dispatchedActions().at(-1)).toMatchObject({
        type: 'push',
        href: { pathname: '/assets/[assetId]', params: { assetId: 'tent' } }
      });
    } finally { await harness.unmount(); }
  });

});

it.each(['pending', 'late-read', 'failed-read'] as const)('keeps Map still with a %s Reduce Motion preference', async mode => {
  resetNativeTestState();
  const snapshot = holdReduceMotionSnapshotForTest();
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const query = new InventoryMapQuery({ listActiveInventoryMapAssets: async () => mapSnapshot });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider><InventoryMapScreen canAdd={false} inventoryMapQuery={query} pathStore={{ current: new Map() }} selectedSurface="map" onAdd={() => undefined} onChangeSurface={() => undefined} /></AppFeedbackProvider>
    </MobileServerStateProvider>);
    if (mode === 'late-read') {
      await h.run(() => setReduceMotionEnabledForTest(true));
      await h.run(() => snapshot.resolve(false));
    }
    if (mode === 'failed-read') await h.run(() => snapshot.reject(new Error('Preference unavailable')));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 80)));
    expect(h.allText()).toContain('Tent');
    expect(animationStartCount()).toBe(0);
    await h.run(() => setReduceMotionEnabledForTest(false));
    await h.run(() => new Promise(resolve => setTimeout(resolve, 80)));
    expect(animationStartCount()).toBeGreaterThan(0);
  } finally { await h.unmount(); snapshot.resolve(false); resetNativeTestState(); }
});

it('disables Map retry while pending and restores results without a pull indicator', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  let calls = 0; let finish!: (value: typeof mapSnapshot) => void;
  const query = new InventoryMapQuery({ listActiveInventoryMapAssets: async () => {
    calls++;
    if (calls === 1) throw new Error('Unavailable');
    return new Promise<typeof mapSnapshot>(resolve => { finish = resolve; });
  } });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider><InventoryMapScreen canAdd={false} inventoryMapQuery={query} pathStore={{ current: new Map() }} selectedSurface="map" onAdd={() => undefined} onChangeSurface={() => undefined} /></AppFeedbackProvider>
    </MobileServerStateProvider>);
    await settle(h); await settle(h);
    const retry = h.byLabel('Retry map');
    await h.run(() => setScreenFocused(false));
    await h.press(retry); expect(calls).toBe(1);
    await h.run(() => setScreenFocused(true));
    await h.press(h.byLabel('Retry map'));
    await settle(h);
    expect(h.byLabel('Retry map')?.props.disabled).toBe(true);
    await h.press(h.byLabel('Retry map'));
    expect(calls).toBe(2);
    await h.run(() => finish(mapSnapshot)); await settle(h);
    expect(h.allText()).toContain('Tent');
    expect(h.allByType('RefreshControl').every(node => node.props.refreshing === false)).toBe(true);
  } finally { await h.unmount(); }
});

it('explains unsuccessful path search and clears stale feedback when criteria change', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const query = new InventoryMapQuery({ listActiveInventoryMapAssets: async () => mapSnapshot });
  const render = (text: string) => h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
    <AppFeedbackProvider><InventoryMapScreen searchQuery={text} canAdd={false} inventoryMapQuery={query} pathStore={{ current: new Map() }} selectedSurface="map" onAdd={() => undefined} onChangeSurface={() => undefined} /></AppFeedbackProvider>
  </MobileServerStateProvider>);
  try {
    await render('missing'); await h.run(() => new Promise(resolve => setTimeout(resolve, 350)));
    expect(h.allText().join(' ')).toContain('No matching items');
    expect(h.allText()).toContain('Tent');
    await render('Tent'); expect(h.allText().join(' ')).not.toContain('No matching items');
    await h.run(() => new Promise(resolve => setTimeout(resolve, 350)));
    expect(h.allText().join(' ')).toContain('Found Tent');
    await render(''); expect(h.allText().join(' ')).not.toContain('Found Tent');
  } finally { await h.unmount(); }
});
