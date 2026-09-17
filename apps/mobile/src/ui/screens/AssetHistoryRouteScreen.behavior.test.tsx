import { AppFeedbackProvider } from '../feedback/AppFeedback';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { AssetHistoryRouteScreen } from './AssetHistoryRouteScreen';
import { AssetActivityQuery } from '../../application/assets/AssetActivityQuery';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { setScreenFocused } from '../../test-support/navigation';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
describe('History query pages', () => {
  it('retries cached failure from its button without starting a pull indicator', async () => {
    const client = createMobileQueryClient(); const h = new MobileRenderHarness();
    let mode: 'ready' | 'fail' | 'retry' = 'ready'; let finish!: () => void;
    const query = new AssetActivityQuery({ listAssetActivity: async () => {
      if (mode === 'fail') throw new Error('Unavailable');
      if (mode === 'retry') await new Promise<void>(resolve => { finish = resolve; });
      return { entries: [], hasMore: false };
    } });
    try {
      await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <AppFeedbackProvider><AssetHistoryRouteScreen assetActivityQuery={query} tenantId="tenant" inventoryId="inventory" assetId="asset" assetTitle="Item" /></AppFeedbackProvider>
      </MobileServerStateProvider>); await settle(h);
      mode = 'fail'; await h.run(() => client.invalidateQueries()); await settle(h);
      mode = 'retry'; await h.press(h.byLabel('Try refreshing again')); await settle(h);
      expect(h.byType('RefreshControl')!.props.refreshing).toBe(false);
      expect(h.byLabel('Try refreshing again')?.props.disabled).toBe(true);
      await h.run(() => finish()); await settle(h);
      expect(h.byLabel('Try refreshing again')).toBeUndefined();
    } finally { await h.unmount(); client.clear(); }
  });
  it.each(['current', 'departed', 'returned'])('owns delayed pull failure in the %s visit', async visit => {
    const client = createMobileQueryClient(); const h = new MobileRenderHarness();
    let fail = false; let finish!: () => void;
    const query = new AssetActivityQuery({ listAssetActivity: async () => {
      if (fail) { await new Promise<void>(resolve => { finish = resolve; }); throw new Error('Unavailable'); }
      return { entries: [{ id: 'one', principalId: 'person', action: 'asset.updated', category: 'change', source: 'api', occurredAt: '2026-07-14T12:00:00Z', changes: [{ field: 'title', currentValue: 'Retained history' }], technical: {} }], hasMore: false };
    } });
    try {
      setScreenFocused(true);
      await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <AppFeedbackProvider><AssetHistoryRouteScreen assetActivityQuery={query} tenantId="tenant" inventoryId="inventory" assetId="asset" assetTitle="Item" /></AppFeedbackProvider>
      </MobileServerStateProvider>); await settle(h);
      fail = true;
      await h.run(() => h.byType('RefreshControl')!.props.onRefresh()); await settle(h);
      if (visit !== 'current') await h.run(() => setScreenFocused(false));
      if (visit === 'returned') await h.run(() => setScreenFocused(true));
      await h.run(() => finish()); await settle(h);
      expect(Boolean(h.byText('Could not refresh History'))).toBe(visit === 'current');
      expect(h.allText().join(' ')).toContain('Retained history');
      expect(h.byType('RefreshControl')!.props.refreshing).toBe(false);
    } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
  });
  it('shares warm pages on reopening and preserves them after a failed refresh', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const cursors: (string | undefined)[] = [];
    let fail = false;
    let denied = false;
    let retryFailure = false;
    const query = new AssetActivityQuery({ listAssetActivity: async ({ cursor }) => {
      cursors.push(cursor);
      if (retryFailure) throw Object.assign(new Error('unavailable'), { status: 500 });
      if (denied) throw Object.assign(new Error('denied'), { status: 403 });
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
      denied = true;
      await harness.run(() => client.invalidateQueries()); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('first');
      denied = false; retryFailure = true;
      await harness.run(() => client.invalidateQueries({ predicate: (query) => query.queryKey.includes('history') })); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('first');
      retryFailure = false;
      expect(harness.allText().join(' ')).not.toContain('second');
    } finally { await harness.unmount(); }
  });
});
