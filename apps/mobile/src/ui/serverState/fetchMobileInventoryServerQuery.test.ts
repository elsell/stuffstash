import { describe, expect, it } from 'vitest';
import { QueryObserver } from '@tanstack/react-query';
import { QueryClientInventoryMutationObserver } from '../../adapters/serverState/QueryClientInventoryMutationObserver';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { fetchMobileInventoryServerQuery } from './fetchMobileInventoryServerQuery';

describe('fetchMobileInventoryServerQuery', () => {
  it('joins active mutation reconciliation and reuses it after completion', async () => {
    const client = createMobileQueryClient();
    const serverState = {
      scopeId: 'scope-one',
      loadInventoryScope: async () => ({ tenantId: 'tenant-home', inventoryId: 'inventory-home' })
    };
    const key = (scope: string, tenant: string, inventory: string) =>
      mobileQueryKeys.assetCore(scope, tenant, inventory, 'asset-one');
    const queryKey = key('scope-one', 'tenant-home', 'inventory-home');
    client.setQueryData(queryKey, 'available');
    let requests = 0;
    let aborted = false;
    let finish!: (value: string) => void;
    const query = (signal?: AbortSignal) => {
      requests += 1;
      signal?.addEventListener('abort', () => { aborted = true; });
      return new Promise<string>((resolve) => { finish = resolve; });
    };
    const active = new QueryObserver(client, { queryKey, queryFn: ({ signal }) => query(signal) });
    const unsubscribe = active.subscribe(() => undefined);
    new QueryClientInventoryMutationObserver(client, 'scope-one').onInventoryMutation({
      kind: 'asset_checkout_changed', tenantId: 'tenant-home', inventoryId: 'inventory-home', assetId: 'asset-one'
    });
    const reconcile = () => fetchMobileInventoryServerQuery({ client, serverState, key, query });
    const pending = reconcile();
    await Promise.resolve();
    finish('checked out');
    await expect(pending).resolves.toBe('checked out');
    await expect(reconcile()).resolves.toBe('checked out');
    expect(requests).toBe(1);
    expect(aborted).toBe(false);
    unsubscribe();
    client.clear();
  });
  it('shares selected scope and fresh resource results across imperative consumers', async () => {
    const client = createMobileQueryClient();
    let scopeRequests = 0;
    let resourceRequests = 0;
    const serverState = {
      scopeId: 'scope-one',
      loadInventoryScope: async () => {
        scopeRequests += 1;
        return { tenantId: 'tenant-home', inventoryId: 'inventory-home' };
      }
    };
    const read = () => fetchMobileInventoryServerQuery({
      client,
      serverState,
      key: mobileQueryKeys.inventoryAssets,
      query: async () => {
        resourceRequests += 1;
        return 'assets';
      }
    });

    await expect(read()).resolves.toBe('assets');
    await expect(read()).resolves.toBe('assets');

    expect(scopeRequests).toBe(1);
    expect(resourceRequests).toBe(1);
  });

  it('forwards Query Client cancellation to the application read', async () => {
    const client = createMobileQueryClient();
    let aborted = false;
    let markStarted: (() => void) | undefined;
    const started = new Promise<void>((resolve) => { markStarted = resolve; });
    const request = fetchMobileInventoryServerQuery({
      client,
      serverState: {
        scopeId: 'scope-one',
        loadInventoryScope: async () => ({ tenantId: 'tenant-home', inventoryId: 'inventory-home' })
      },
      key: mobileQueryKeys.inventoryAssets,
      query: (signal) => new Promise<string>((_resolve, reject) => {
        markStarted?.();
        signal?.addEventListener('abort', () => {
          aborted = true;
          reject(new Error('aborted'));
        });
      })
    });
    await started;

    await client.cancelQueries({ queryKey: mobileQueryKeys.root('scope-one') });

    await expect(request).rejects.toThrow();
    expect(aborted).toBe(true);
  });
});

it('does not start a resource read after its cached scope was reset during reconciliation', async () => {
  const client = createMobileQueryClient(); let reads = 0;
  client.setQueryData(mobileQueryKeys.inventoryScope('scope'), { tenantId: 'old', inventoryId: 'old' });
  const pending = fetchMobileInventoryServerQuery({ client, serverState: { scopeId: 'scope', loadInventoryScope: async () => ({ tenantId: 'next', inventoryId: 'next' }) }, key: mobileQueryKeys.home, query: async () => { reads++; return 'new data'; } });
  client.removeQueries({ queryKey: mobileQueryKeys.inventoryScope('scope') });
  await expect(pending).rejects.toThrow(); expect(reads).toBe(0);
  expect(client.getQueryData(mobileQueryKeys.home('scope', 'old', 'old'))).toBeUndefined();
});
