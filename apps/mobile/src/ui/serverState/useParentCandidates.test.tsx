import React from 'react';
import { Text } from 'react-native';
import { describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { useParentCandidates } from './useParentCandidates';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';

const settle = (harness: MobileRenderHarness, ms = 10) => harness.run(() => new Promise((resolve) => setTimeout(resolve, ms)));
it('debounces typing, cancels superseded reads and reuses warm suggestions', async () => {
  const client = createMobileQueryClient();
  const harness = new MobileRenderHarness();
  const requests: string[] = [];
  let aborted = false;
  const lookup = { execute: (query: string, request?: { signal?: AbortSignal }): Promise<readonly ParentLookupResult[]> => {
    requests.push(query);
    if (query === 'old') {
      request?.signal?.addEventListener('abort', () => { aborted = true; });
      return new Promise(() => undefined);
    }
    return Promise.resolve([]);
  } };
  function Picker({ query, enabled }: { query: string; enabled: boolean }) {
    const result = useParentCandidates(query, lookup, enabled);
    return <Text>{result.data ? 'Ready' : 'Pending'}</Text>;
  }
  const render = (query: string, enabled = true) => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><Picker query={query} enabled={enabled} /></MobileServerStateProvider>;
  try {
    await harness.render(render('old')); await settle(harness);
    await harness.render(render('n')); await harness.render(render('ne')); await harness.render(render('new'));
    expect(aborted).toBe(true);
    expect(requests).toEqual(['old']);
    await settle(harness, 270); await settle(harness);
    expect(requests).toEqual(['old', 'new']);
    await harness.render(render('new', false)); await harness.render(render('new')); await settle(harness);
    expect(requests).toEqual(['old', 'new']);
  } finally { await harness.unmount(); }
});
