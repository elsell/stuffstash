import { describe, expect, it } from 'vitest';
import { MobileAuthenticationRequiredError } from '../../application/auth/MobileAuthSession';
import {
  createMobileQueryClient,
  disposeMobileQueryClient,
  mobileQueryKeys,
  normalizeBrowseQueryIdentity,
  resetMobileInventorySelection,
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
    expect(mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a')).toEqual([
      'mobile', 'scope-a', 'tenant', 'tenant-a', 'inventory', 'inventory-a', 'home'
    ]);
    expect(mobileQueryKeys.inventoryScope('scope-a')).toEqual([
      'mobile',
      'scope-a',
      'inventory-scope'
    ]);
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

  it('deduplicates simultaneous reads of the same scoped resource', async () => {
    const client = createMobileQueryClient();
    let requestCount = 0;
    let resolveRequest: ((value: string) => void) | undefined;
    const queryFn = () => {
      requestCount += 1;
      return new Promise<string>((resolve) => { resolveRequest = resolve; });
    };

    const first = client.fetchQuery({ queryKey: mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a'), queryFn });
    const second = client.fetchQuery({ queryKey: mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a'), queryFn });
    await Promise.resolve();
    resolveRequest?.('ready');

    await expect(Promise.all([first, second])).resolves.toEqual(['ready', 'ready']);
    expect(requestCount).toBe(1);
  });

  it('clears only the active composition when inventory selection changes', async () => {
    const client = createMobileQueryClient();
    client.setQueryData(mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a'), 'old home');
    client.setQueryData(mobileQueryKeys.home('scope-b', 'tenant-a', 'inventory-a'), 'other composition');

    await resetMobileInventorySelection(client, 'scope-a');

    expect(client.getQueryData(mobileQueryKeys.home('scope-a', 'tenant-a', 'inventory-a'))).toBeUndefined();
    expect(client.getQueryData(mobileQueryKeys.home('scope-b', 'tenant-a', 'inventory-a'))).toBe('other composition');
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
      checkoutState: 'any',
      sort: 'default',
      scope: 'all'
    });
  });

  it('keys Browse pages by normalized filters and cursor', () => {
    expect(mobileQueryKeys.browse('scope-a', 'tenant-a', 'inventory-a', {
      query: ' drill ',
      tagIds: ['tag-b', 'tag-a', 'tag-b'],
      lifecycleState: 'active',
      checkoutState: 'any',
      sort: 'updated_desc',
      scope: 'items'
    }, 'cursor-two')).toEqual([
      'mobile', 'scope-a', 'tenant', 'tenant-a', 'inventory', 'inventory-a',
      'browse',
      {
        query: 'drill',
        tagIds: ['tag-a', 'tag-b'],
        lifecycleState: 'active',
        checkoutState: 'any',
        sort: 'updated_desc',
        scope: 'items'
      },
      'cursor-two'
    ]);
  });

  it('retries only one transient read failure and never retries caller errors', () => {
    expect(shouldRetryMobileQuery(0, Object.assign(new Error('server unavailable'), { status: 503 }))).toBe(true);
    expect(shouldRetryMobileQuery(1, Object.assign(new Error('server unavailable'), { status: 503 }))).toBe(false);
    expect(shouldRetryMobileQuery(0, Object.assign(new Error('sign in'), { status: 401 }))).toBe(false);
    expect(shouldRetryMobileQuery(0, Object.assign(new Error('not found'), { status: 404 }))).toBe(false);
    expect(shouldRetryMobileQuery(0, new MobileAuthenticationRequiredError())).toBe(false);
    expect(shouldRetryMobileQuery(0, new TypeError('Network request failed'))).toBe(true);
    expect(shouldRetryMobileQuery(0, new Error('Invalid inventory selection'))).toBe(false);
  });
});
