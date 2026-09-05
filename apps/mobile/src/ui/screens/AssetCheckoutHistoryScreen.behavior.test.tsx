import React from 'react';
import { describe, expect, it } from 'vitest';
import { AssetCheckoutHistorySheetRouteScreen } from './AssetCheckoutHistoryScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { AssetCheckoutHistoryQuery } from '../../application/assets/AssetCheckoutHistoryQuery';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));

describe('checkout History server state', () => {
  it('loads cursor pages independently of the title and retains rows when continuation fails', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const cursors: (string | undefined)[] = [];
    let fail = true;
    const query = new AssetCheckoutHistoryQuery({ listAssetCheckoutHistory: async ({ cursor }) => {
      cursors.push(cursor);
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
      await harness.press(harness.byLabel('Load older checkouts')); await settle(harness);
      expect(harness.allText().join(' ')).toContain('Principal two');
      expect(cursors).toEqual([undefined, 'two', 'two']);
    } finally { await harness.unmount(); }
  });
});
