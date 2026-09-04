import { describe, expect, it } from 'vitest';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';
import {
  createMobileQueryClient,
  disposeMobileQueryClient,
  mobileQueryKeys,
  normalizeBrowseQueryIdentity,
  shouldRetryMobileQuery
} from './MobileQueryClient';

describe('mobile server-state coordination', () => {
  it('uses one bounded, non-mutating retry and finite session retention', () => {
    const client = createMobileQueryClient();
    const defaults = client.getDefaultOptions();

    expect(defaults.queries).toMatchObject({
      staleTime: 30_000,
      gcTime: 300_000,
      refetchOnReconnect: true,
      refetchOnWindowFocus: true
    });
    expect(defaults.queries?.retry).toBe(shouldRetryMobileQuery);
    expect(defaults.mutations).toMatchObject({ retry: false });
  });

  it('cancels active work and clears cached server state when its composition ends', async () => {
    const client = createMobileQueryClient();
    let wasAborted = false;
    const activeQuery = client.fetchQuery({
      queryKey: ['mobile', 'scope-one', 'slow-query'],
      queryFn: ({ signal }) => new Promise<string>((_resolve, reject) => {
        signal.addEventListener('abort', () => {
          wasAborted = true;
          reject(new Error('aborted'));
        });
      })
    });
    await Promise.resolve();

    await disposeMobileQueryClient(client);

    await expect(activeQuery).rejects.toThrow();
    expect(wasAborted).toBe(true);
    expect(client.getQueryCache().getAll()).toEqual([]);
  });

  it('isolates keys by composition and complete resource scope', () => {
    expect(mobileQueryKeys.asset('scope-a', 'tenant-a', 'inventory-a', 'asset-a')).toEqual([
      'mobile',
      'scope-a',
      'tenant',
      'tenant-a',
      'inventory',
      'inventory-a',
      'asset',
      'asset-a'
    ]);
    expect(mobileQueryKeys.asset('scope-b', 'tenant-a', 'inventory-a', 'asset-a'))
      .not.toEqual(mobileQueryKeys.asset('scope-a', 'tenant-a', 'inventory-a', 'asset-a'));
  });

  it('normalizes equivalent Browse identities for request deduplication', () => {
    expect(normalizeBrowseQueryIdentity({
      query: '  drill  ',
      tagIds: ['tag-b', 'tag-a', 'tag-b'],
      lifecycleState: undefined,
      checkoutState: undefined,
      sort: undefined,
      scope: undefined
    })).toEqual({
      query: 'drill',
      tagIds: ['tag-a', 'tag-b'],
      lifecycleState: 'active',
      checkoutState: 'all',
      sort: 'default',
      scope: 'all'
    });
  });

  it('retries only one transient read failure and never retries caller errors', () => {
    expect(shouldRetryMobileQuery(0, Object.assign(new Error('server unavailable'), { status: 503 }))).toBe(true);
    expect(shouldRetryMobileQuery(1, Object.assign(new Error('server unavailable'), { status: 503 }))).toBe(false);
    expect(shouldRetryMobileQuery(0, Object.assign(new Error('sign in'), { status: 401 }))).toBe(false);
    expect(shouldRetryMobileQuery(0, Object.assign(new Error('not found'), { status: 404 }))).toBe(false);
    expect(shouldRetryMobileQuery(0, new MobileAuthenticationRequiredError())).toBe(false);
    expect(shouldRetryMobileQuery(0, new Error('network disconnected'))).toBe(true);
  });
});
