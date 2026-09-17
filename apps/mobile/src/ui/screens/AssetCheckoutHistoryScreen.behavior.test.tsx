import React from 'react';
import { dispatchedActions, resetNavigation, setCanGoBack } from '../../test-support/navigation';
import { describe, expect, it } from 'vitest';
import { AssetCheckoutHistorySheetRouteScreen } from './AssetCheckoutHistoryScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import { assetId } from '../../domain/assets/AssetSummary';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import { AssetCheckoutHistoryQuery } from '../../application/assets/AssetCheckoutHistoryQuery';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));

it.each(['loading', 'error', 'ready'] as const)('closes %s history from direct entry or the existing stack', async status => {
  for (const canGoBack of [false, true]) {
    resetNavigation(); setCanGoBack(canGoBack);
    const harness = new MobileRenderHarness();
    try {
      await harness.render(<MobileServerStateProvider client={createMobileQueryClient()} scopeId="scope"
        loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <AssetCheckoutHistorySheetRouteScreen assetId="asset"
          assetCoreQuery={{ execute: () => new Promise(() => undefined) }}
          assetCheckoutHistoryQuery={{ execute: async () => {
            if (status === 'loading') return new Promise(() => undefined);
            if (status === 'error') throw new Error('unavailable');
            return { assetId: 'asset', records: [], hasMore: false, emptyTitle: 'No checkouts', emptyMessage: 'No checkout history yet.' };
          } }} />
      </MobileServerStateProvider>);
      await settle(harness); await settle(harness);
      expect(harness.allText()).toContain(status === 'loading' ? 'Loading checkout history'
        : status === 'error' ? 'Could not load checkout history' : 'No checkouts');
      await harness.press(harness.byLabel('Close'));
      expect(dispatchedActions()).toEqual([canGoBack ? { type: 'back' } : { type: 'replace', href: '/' }]);
    } finally { await harness.unmount(); resetNavigation(); setCanGoBack(true); }
  }
});

describe('checkout History server state', () => {
  it('loads cursor pages independently of the title and retains rows when continuation fails', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const cursors: (string | undefined)[] = [];
    let fail = true;
    let denied = false;
    let retryFailure = false;
    const query = new AssetCheckoutHistoryQuery({ listAssetCheckoutHistory: async ({ cursor }) => {
      cursors.push(cursor);
      if (retryFailure) throw Object.assign(new Error('unavailable'), { status: 500 });
      if (denied) throw Object.assign(new Error('denied'), { status: 403 });
      if (cursor && fail) throw new Error('failed');
      return { records: [{ id: cursor ?? 'one', state: 'open', checkedOutAt: '2026-07-14T12:00:00Z', checkedOutByPrincipalId: cursor ?? 'first' }], hasMore: !cursor, nextCursor: cursor ? undefined : 'two' };
    } });
    try {
      await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <AssetCheckoutHistorySheetRouteScreen assetId="asset" assetCheckoutHistoryQuery={query} assetCoreQuery={{ execute: () => new Promise(() => undefined) }} />
      </MobileServerStateProvider>);
      await settle(harness); await settle(harness);
      expect(harness.allText().join(' ')).toContain('Principal first');
      await harness.press(harness.byLabel('Load older checkouts')); await settle(harness);
      expect(harness.allText().join(' ')).toContain('Principal first');
      expect(harness.allText().join(' ')).toContain('Older checkouts could not be loaded.');
      fail = false;
      await harness.press(harness.byLabel('Try older checkouts again')); await settle(harness);
      expect(harness.allText().join(' ')).toContain('Principal two');
      expect(cursors).toEqual([undefined, 'two', 'two']);
      denied = true;
      await harness.run(() => client.invalidateQueries({ predicate: (query) => query.queryKey.includes('checkouts') })); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('Principal first');
      denied = false; retryFailure = true;
      await harness.run(() => client.invalidateQueries({ predicate: (query) => query.queryKey.includes('checkouts') })); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('Principal first');
      retryFailure = false;
      expect(harness.allText().join(' ')).not.toContain('Principal two');
    } finally { await harness.unmount(); }
  });
});


it('offers native Close while checkout history is loading', async () => {
  resetNavigation();
  const harness = new MobileRenderHarness();
  try {
    await harness.render(<MobileServerStateProvider client={createMobileQueryClient()} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetCheckoutHistorySheetRouteScreen assetId="asset" assetCheckoutHistoryQuery={{ execute: () => new Promise(() => undefined) }} assetCoreQuery={{ execute: () => new Promise(() => undefined) }} />
    </MobileServerStateProvider>);
    expect(harness.byLabel('Close')).toBeDefined();
    await harness.press(harness.byLabel('Close'));
    expect(dispatchedActions()).toEqual([{ type: 'back' }]);
  } finally { await harness.unmount(); resetNavigation(); }
});


it.each([500, 401, 403, 404])('separates transient name recovery from history while preserving access failure %s', async (status) => {
  resetNavigation();
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  let failName = true;
  let retryGate: Promise<void> | undefined;
  let historyCalls = 0;
  const core = new AssetCoreQuery({ getAssetCore: async () => {
    await retryGate;
    if (failName) throw Object.assign(new Error('Name unavailable'), { status });
    return { tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), permissions: ['view'], revision: 'one',
      asset: { id: assetId('asset'), title: 'Ladder', kind: 'item', lifecycleState: 'active', description: '',
        locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false } };
  } });
  const history = new AssetCheckoutHistoryQuery({ listAssetCheckoutHistory: async () => {
    historyCalls++;
    return { records: [{ id: 'one', state: 'open', checkedOutAt: '2026-07-14T12:00:00Z', checkedOutByPrincipalId: 'first' }], hasMore: false };
  } });
  // Keep the real cache/error policy; shorten the production retry delay for this controlled boundary.
  client.setDefaultOptions({ queries: { ...client.getDefaultOptions().queries, retryDelay: 0 } });
  try {
    await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AssetCheckoutHistorySheetRouteScreen assetId="asset" assetCoreQuery={core} assetCheckoutHistoryQuery={history} />
    </MobileServerStateProvider>);
    await settle(harness); await settle(harness); await settle(harness);
    if (status === 500) {
      expect(harness.allText().join(' ')).toContain('Principal first');
      expect(harness.allText().join(' ')).toContain('Asset name could not be loaded.');
      expect(harness.allText().join(' ')).not.toContain('Could not load checkout history');
    } else {
      expect(harness.allText().join(' ')).not.toContain('Principal first');
      expect(harness.allText().join(' ')).toContain('Could not load checkout history');
    }
    let finishRetry!: () => void;
    retryGate = new Promise<void>(resolve => { finishRetry = resolve; });
    await harness.press(harness.byLabel(status === 500 ? 'Try loading asset name again' : 'Try again'));
    await settle(harness);
    if (status !== 500) expect(harness.allText().join(' ')).not.toContain('Principal first');
    await harness.run(() => finishRetry());
    await settle(harness); await settle(harness);
    if (status !== 500) expect(harness.allText().join(' ')).not.toContain('Principal first');
    retryGate = undefined;
    failName = false;
    await harness.press(harness.byLabel(status === 500 ? 'Try loading asset name again' : 'Try again'));
    await settle(harness); await settle(harness);
    expect(harness.allText().join(' ')).toContain('Ladder');
    expect(harness.allText().join(' ')).toContain('Principal first');
    expect(historyCalls).toBe(1);
    expect(harness.allText().join(' ')).not.toContain('Asset name could not be loaded.');
  } finally { await harness.unmount(); client.clear(); resetNavigation(); }
});
