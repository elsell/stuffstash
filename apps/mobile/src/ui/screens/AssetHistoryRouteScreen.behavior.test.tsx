import { AppFeedbackProvider } from '../feedback/AppFeedback';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { AssetHistoryRouteScreen } from './AssetHistoryRouteScreen';
import { AssetActivityQuery } from '../../application/assets/AssetActivityQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
describe('History query pages', () => {
  it('shares warm pages on reopening and preserves them after a failed refresh', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const cursors: (string | undefined)[] = [];
    let fail = false;
    const query = new AssetActivityQuery({ listAssetActivity: async ({ cursor }) => {
      cursors.push(cursor);
      if (fail) throw new Error('unavailable');
      return { entries: [{ id: cursor ?? 'one', principalId: 'person', action: 'asset.updated', category: 'change', source: 'api', occurredAt: '2026-07-14T12:00:00Z', changes: [{ field: 'title', currentValue: cursor ?? 'first' }], technical: {} }], hasMore: !cursor, nextCursor: cursor ? undefined : 'second' };
    } });
    const render = (shown: boolean) => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider>{shown ? <AssetHistoryRouteScreen assetActivityQuery={query} tenantId="tenant" inventoryId="inventory" assetId="asset" assetTitle="Item" /> : null}</AppFeedbackProvider>
    </MobileServerStateProvider>;
    try {
      await harness.render(render(true)); await settle(harness);
      await harness.press(harness.byLabel('Load older activity')); await settle(harness);
      expect(harness.allText().join(' ')).toContain('second');
      await harness.render(render(false)); await harness.render(render(true)); await settle(harness);
      expect(cursors).toEqual([undefined, 'second']);
      fail = true;
      await harness.run(() => harness.byType('RefreshControl')!.props.onRefresh()); await settle(harness);
      expect(harness.allText().join(' ')).toContain('first');
      expect(harness.allText().join(' ')).toContain('second');
    } finally { await harness.unmount(); }
  });
});
