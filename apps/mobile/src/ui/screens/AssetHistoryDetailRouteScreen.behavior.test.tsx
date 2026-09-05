import { AppFeedbackProvider } from '../feedback/AppFeedback';
import React from 'react';
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
  it('reuses a fresh scoped page and cancels an unrelated asset lookup on leaving', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const requests: string[] = [];
    let aborted = false;
    const query = new AssetActivityQuery({ listAssetActivity: ({ assetId, signal }) => {
      requests.push(assetId);
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
      await harness.render(render('two')); await settle(harness);
      expect(harness.allText().join(' ')).not.toContain('Cached name');
      expect(requests).toEqual(['two']);
    } finally { await harness.unmount(); }
    expect(aborted).toBe(true);
  });
});
