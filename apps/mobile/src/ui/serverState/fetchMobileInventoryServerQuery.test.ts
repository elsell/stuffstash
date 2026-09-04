import { describe, expect, it } from 'vitest';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { fetchMobileInventoryServerQuery } from './fetchMobileInventoryServerQuery';

describe('fetchMobileInventoryServerQuery', () => {
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
