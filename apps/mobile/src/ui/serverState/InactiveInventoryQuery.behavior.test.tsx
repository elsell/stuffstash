import React from 'react';
import { describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileInventoryServerQuery } from './useMobileInventoryServerQuery';

function Consumer({ enabled, query }: { enabled: boolean; query: (signal?: AbortSignal) => Promise<string> }) {
  useMobileInventoryServerQuery({ key: mobileQueryKeys.locations, query, enabled });
  return null;
}

describe('inactive inventory queries', () => {
  it('cancels hidden region work only after its last active consumer leaves', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let aborted = false;
    let calls = 0;
    const query = (signal?: AbortSignal) => {
      calls++;
      return new Promise<string>((_resolve, reject) => {
        signal?.addEventListener('abort', () => { aborted = true; reject(new Error('aborted')); });
      });
    };
    const render = (first: boolean, second: boolean) => harness.render(
      <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <Consumer enabled={first} query={query} /><Consumer enabled={second} query={query} />
      </MobileServerStateProvider>
    );
    try {
      await render(true, true);
      await harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
      expect(calls).toBe(1);
      await render(false, true);
      expect(aborted).toBe(false);
      await render(false, false);
      expect(aborted).toBe(true);
    } finally { await harness.unmount(); }
  });
});
