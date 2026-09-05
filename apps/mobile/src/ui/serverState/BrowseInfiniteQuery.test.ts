import { describe, expect, it } from 'vitest';
import { InfiniteQueryObserver } from '@tanstack/react-query';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { browseInfiniteQueryOptions } from './BrowseInfiniteQuery';
import type { SearchAssetsViewModel } from '../../application/search/SearchAssetsQuery';

const page = (query: string, cursor?: string): SearchAssetsViewModel => ({
  query, mode: 'browse', lifecycleState: 'active', kind: 'all', checkoutState: 'any', sort: 'updated_desc', tagIds: [],
  assets: [], nextCursor: cursor, hasMore: Boolean(cursor)
});

describe('Browse infinite query', () => {
  it('shares fresh pages and owns continuation cursors', async () => {
    const client = createMobileQueryClient();
    const requests: (string | undefined)[] = [];
    const options = browseInfiniteQueryOptions('scope', 'tenant', 'inventory', {
      query: ' drill ', scope: 'all', lifecycleState: 'active', checkoutState: 'any', sort: 'updated_desc', tagIds: []
    }, { execute: async (input) => { requests.push(input.cursor); return page(input.query, input.cursor ? undefined : 'next'); } });
    await client.fetchInfiniteQuery(options);
    const observer = new InfiniteQueryObserver(client, options);
    const unsubscribe = observer.subscribe(() => undefined);
    await observer.fetchNextPage({ cancelRefetch: false });
    expect(requests).toEqual([undefined, 'next']);
    expect(observer.getCurrentResult().data?.pages).toHaveLength(2);
    await client.fetchInfiniteQuery(options);
    expect(requests).toHaveLength(2);
    unsubscribe(); client.clear();
  });

  it('cancels a superseded filter request at the application boundary', async () => {
    const client = createMobileQueryClient();
    let aborted = false;
    const criteria = { query: 'old', scope: 'all' as const, lifecycleState: 'active' as const, checkoutState: 'any' as const, sort: 'updated_desc' as const, tagIds: [] };
    const query = { execute: (input: { query: string; signal?: AbortSignal }) => input.query === 'new'
      ? Promise.resolve(page('new'))
      : new Promise<SearchAssetsViewModel>((_resolve, reject) => {
          input.signal?.addEventListener('abort', () => { aborted = true; reject(new Error('aborted')); });
        }) };
    const observer = new InfiniteQueryObserver(client, browseInfiniteQueryOptions('scope', 'tenant', 'inventory', criteria, query));
    const unsubscribe = observer.subscribe(() => undefined);
    observer.setOptions(browseInfiniteQueryOptions('scope', 'tenant', 'inventory', { ...criteria, query: 'new' }, query));
    await observer.refetch({ cancelRefetch: false });
    expect(aborted).toBe(true);
    expect(observer.getCurrentResult().data?.pages[0]?.query).toBe('new');
    unsubscribe(); client.clear();
  });
});
