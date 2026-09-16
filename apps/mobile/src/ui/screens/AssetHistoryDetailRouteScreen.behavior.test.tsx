import { AppFeedbackProvider } from '../feedback/AppFeedback';
import React from 'react';
import { latestAlert } from '../../test-support/react-native';
import { dispatchedActions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { describe, expect, it } from 'vitest';
import { AssetHistoryDetailRouteScreen } from './AssetHistoryDetailRouteScreen';
import { AssetActivityQuery } from '../../application/assets/AssetActivityQuery';
import { RevertAssetChangeCommand } from '../../application/assets/RevertAssetChangeCommand';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';

const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
const entry = { id: 'activity', principalId: 'person', action: 'asset.updated', category: 'change' as const, source: 'api', occurredAt: '2026-07-14T12:00:00Z', changes: [{ field: 'title' as const, currentValue: 'Cached name' }], technical: {} };

describe('History detail cache', () => {
  it.each([undefined, '', '   ', '  owner@example.test  '])('presents a readable actor for email %s', async (email) => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    client.setQueryData(mobileQueryKeys.assetActivity('scope', 'tenant', 'inventory', 'asset', 'activity'), {
      ...entry, principalId: 'opaque-principal-123', principal: { id: 'opaque-principal-123', email }
    });
    const query = new AssetActivityQuery({ listAssetActivity: async () => { throw new Error('Fresh cache should be used'); } });
    try {
      await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <AppFeedbackProvider><AssetHistoryDetailRouteScreen assetActivityQuery={query} revertAssetChangeCommand={new RevertAssetChangeCommand({ reverseAssetOperation: async () => undefined })} activityId="activity" assetId="asset" assetTitle="Item" tenantId="tenant" inventoryId="inventory" /></AppFeedbackProvider>
      </MobileServerStateProvider>);
      const text = harness.allText().join(' ');
      expect(text).toContain(email?.trim() || 'Someone with access');
      expect(text).not.toContain('opaque-principal-123');
    } finally { await harness.unmount(); }
  });

  it('reuses a fresh scoped page and cancels an unrelated asset lookup on leaving', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const requests: string[] = [];
    let aborted = false;
    let denied = false;
    let retryFailure = false;
    const query = new AssetActivityQuery({ listAssetActivity: ({ assetId, signal }) => {
      requests.push(assetId);
      if (retryFailure) throw Object.assign(new Error('unavailable'), { status: 500 });
      if (denied) return Promise.reject(Object.assign(new Error('denied'), { status: 403 }));
      signal?.addEventListener('abort', () => { aborted = true; });
      return new Promise(() => undefined);
    } });
    client.setQueryData(mobileQueryKeys.assetHistory('scope', 'tenant', 'inventory', 'one', 'changes'), {
      pages: [{ entries: [entry], records: [], hasMore: false }], pageParams: [undefined]
    });
    const render = (assetId: string) => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider><AssetHistoryDetailRouteScreen assetActivityQuery={query} revertAssetChangeCommand={new RevertAssetChangeCommand({ reverseAssetOperation: async () => undefined })} activityId="activity" assetId={assetId} assetTitle="Item" tenantId="tenant" inventoryId="inventory" /></AppFeedbackProvider>
    </MobileServerStateProvider>;
    try {
      await harness.render(render('one')); await settle(harness);
      expect(harness.allText().join(' ')).toContain('Cached name');
      expect(requests).toEqual([]);
      denied = true;
      await harness.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.assetActivity('scope', 'tenant', 'inventory', 'one', 'activity') })); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('Cached name');
      denied = false; retryFailure = true;
      await harness.run(() => client.invalidateQueries({ predicate: (query) => query.queryKey.includes('activity') })); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('Cached name');
      retryFailure = false;
      denied = false;
      requests.length = 0;
      await harness.render(render('two')); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('Cached name');
      expect(requests).toEqual(['two']);
    } finally { await harness.unmount(); }
    expect(aborted).toBe(true);
  });
});

function deferredReversal() {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<void>((finish, fail) => { resolve = finish; reject = fail; });
  return { promise, resolve, reject };
}

async function reversalFixture() {
  resetNavigation();
  setScreenFocused(true);
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  const pending: ReturnType<typeof deferredReversal>[] = [];
  const operations: string[] = [];
  for (const activityId of ['one', 'two']) client.setQueryData(
    mobileQueryKeys.assetActivity('scope', 'tenant', 'inventory', 'asset', activityId),
    { ...entry, id: activityId, undo: { status: 'available', operationId: activityId } }
  );
  const command = new RevertAssetChangeCommand({ reverseAssetOperation: ({ operationId }) => {
    operations.push(operationId);
    const request = deferredReversal(); pending.push(request); return request.promise;
  } });
  const query = new AssetActivityQuery({ listAssetActivity: async () => { throw new Error('Fresh cache should be used'); } });
  const render = (activityId: string | undefined) => harness.render(
    <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <AppFeedbackProvider>{activityId ? <AssetHistoryDetailRouteScreen assetActivityQuery={query} revertAssetChangeCommand={command} activityId={activityId} assetId="asset" assetTitle="Tent" tenantId="tenant" inventoryId="inventory" /> : null}</AppFeedbackProvider>
    </MobileServerStateProvider>
  );
  const confirm = async () => {
    await harness.press(harness.byText('Revert change')?.parent ?? undefined);
    const callback = latestAlert()?.buttons.find((button) => button.text === 'Revert Change')?.onPress;
    expect(callback).toBeTypeOf('function');
    return callback!;
  };
  await render('one');
  return { harness, client, operations, pending, render, confirm, close: async () => { await harness.unmount(); setScreenFocused(true); resetNavigation(); } };
}

describe('History reversal presentation ownership', () => {
  it.each(['refresh failure', 'changed operation'] as const)('retires retained confirmation after %s', async change => {
    const f = await reversalFixture();
    const queryKey = mobileQueryKeys.assetActivity('scope', 'tenant', 'inventory', 'asset', 'one');
    try {
      const stale = await f.confirm();
      if (change === 'refresh failure') {
        await f.harness.run(() => f.client.invalidateQueries({ queryKey, exact: true }));
      } else {
        await f.harness.run(() => f.client.setQueryData(queryKey, { ...entry, id: 'one', undo: { status: 'available', operationId: 'new-operation' } }));
      }
      await settle(f.harness);
      await f.harness.run(stale);
      expect(f.operations).toEqual([]);
      await f.harness.run(() => f.client.setQueryData(queryKey, { ...entry, id: 'one', undo: { status: 'available', operationId: 'fresh-operation' } }));
      await settle(f.harness);
      await f.harness.run(stale);
      expect(f.operations).toEqual([]);
      await f.harness.run(await f.confirm());
      expect(f.operations).toEqual(['fresh-operation']);
      await f.harness.run(() => f.pending[0]!.resolve());
      expect(dispatchedActions()).toEqual([{ type: 'back' }]);
    } finally { await f.close(); }
  });

  it('submits once, retains failed activity for retry, and returns after successful retry', async () => {
    const f = await reversalFixture();
    try {
      const callback = await f.confirm();
      await f.harness.run(() => { callback(); callback(); });
      expect(f.operations).toEqual(['one']);
      expect(f.harness.byText('Reverting…')).toBeDefined();
      await f.harness.run(() => f.pending[0]!.reject(new Error('unavailable')));
      expect(f.harness.byText('Revert change')).toBeDefined();
      const retry = await f.confirm();
      await f.harness.run(retry);
      await f.harness.run(() => f.pending[1]!.resolve());
      await f.harness.run(retry);
      expect(f.operations).toEqual(['one', 'one']);
      expect(dispatchedActions()).toEqual([{ type: 'back' }]);
      expect(f.harness.allText().join(' ')).toContain('Change reverted');
    } finally { await f.close(); }
  });

  it('invalidates completed reversal without navigating or showing feedback after blur and return', async () => {
    const f = await reversalFixture();
    try {
      const confirmation = await f.confirm();
      await f.harness.run(confirmation);
      await f.harness.run(() => setScreenFocused(false));
      await f.harness.run(() => setScreenFocused(true));
      await f.harness.run(() => f.pending[0]!.resolve());
      expect(dispatchedActions()).toEqual([]);
      expect(f.harness.allText().join(' ')).not.toContain('Change reverted');
      expect(f.client.getQueryState(mobileQueryKeys.assetActivity('scope', 'tenant', 'inventory', 'asset', 'one'))?.isInvalidated).toBe(true);
      expect(f.harness.byText('Reverting…')).toBeUndefined();
      expect(f.harness.byText('Revert change')).toBeUndefined();
      expect(f.harness.byText('This change has been reverted.')).toBeDefined();
      await f.harness.run(confirmation);
      expect(f.operations).toEqual(['one']);
    } finally { await f.close(); }
  });

  it('ignores stale confirmations and leaves another activity independent of pending completion', async () => {
    const f = await reversalFixture();
    try {
      const firstConfirmation = await f.confirm();
      await f.harness.run(firstConfirmation);
      await f.render('two');
      expect(f.harness.byText('Revert change')).toBeDefined();
      await f.harness.run(firstConfirmation);
      expect(f.operations).toEqual(['one']);
      await f.harness.run(await f.confirm());
      expect(f.operations).toEqual(['one', 'two']);
      await f.harness.run(() => f.pending[0]!.resolve());
      expect(dispatchedActions()).toEqual([]);
      expect(f.harness.byText('Reverting…')).toBeDefined();
      await f.harness.run(() => f.pending[1]!.resolve());
      expect(dispatchedActions()).toEqual([{ type: 'back' }]);
    } finally { await f.close(); }
  });

  it('does not start a stale confirmation after leaving or publish a failed reversal after teardown', async () => {
    const f = await reversalFixture();
    try {
      const stale = await f.confirm();
      await f.harness.run(() => setScreenFocused(false));
      await f.harness.run(stale);
      expect(f.operations).toEqual([]);
      await f.harness.run(() => setScreenFocused(true));
      await f.harness.run(await f.confirm());
      await f.render(undefined);
      await f.harness.run(() => f.pending[0]!.reject(new Error('unavailable')));
      expect(f.harness.allText()).toEqual([]);
      expect(dispatchedActions()).toEqual([]);
    } finally { await f.close(); }
  });
});
