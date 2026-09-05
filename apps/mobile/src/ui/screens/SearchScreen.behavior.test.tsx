import React from 'react';
import { describe, expect, it } from 'vitest';
import { SearchScreen } from './SearchScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { toAssetCardViewModel } from '../../application/assets/AssetViewModels';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import { PhotoSelectionQuery } from '../../application/add/PhotoSelectionQuery';
import type { AssetBrowsePage } from '../../application/home/InventorySummaryRepository';

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((finish, fail) => { resolve = finish; reject = fail; });
  return { promise, resolve, reject };
}
const settle = (harness: MobileRenderHarness) => harness.run(() => new Promise((resolve) => setTimeout(resolve, 10)));
const asset = (title: string): AssetSummary => ({
  id: assetId(title), title, description: '', kind: 'item', lifecycleState: 'active', locationLabel: '', locationTrail: [], parentLocationTrail: [], updatedAtLabel: '', hasPhoto: false
});

function propsFor(overrides: Partial<React.ComponentProps<typeof SearchScreen>>): React.ComponentProps<typeof SearchScreen> {
  return {
    searchAssetsQuery: new SearchAssetsQuery({ browseAssets: async () => ({ assets: [], hasMore: false }) }),
    inventoryContextQuery: { execute: async () => ({ inventoryName: 'Home', canAdd: true }) },
    inventoryAssetTagsQuery: { execute: async () => [] },
    locationsQuery: { execute: async () => ({ inventoryName: 'Home', tenantName: 'Tenant', canAdd: true, locations: [] }) },
    inventoryMapQuery: { execute: async () => { throw new Error('Map must not be loaded by List'); } },
    assetCoreQuery: { execute: async () => { throw new Error('No detail selected'); } },
    assetContentsQuery: { execute: async () => { throw new Error('No detail selected'); } },
    assetPhotosQuery: { execute: async () => [] },
    assetCheckoutCommand: { execute: async () => { throw new Error('No checkout selected'); } },
    assetLifecycleCommand: { execute: async () => undefined },
    addAssetPhotosCommand: { execute: async () => ({ attachedCount: 0, failedCount: 0, failedPhotos: [], canRetry: false, message: '' }) },
    deleteAssetPhotoCommand: { execute: async () => ({ message: '' }) },
    photoSelectionQuery: new PhotoSelectionQuery({ selectFromLibrary: async () => [], captureFromCamera: async () => [] }),
    ...overrides
  };
}

describe('mounted Browse server state', () => {
  it('switches inventory without retaining old rows or allowing delayed old work to win', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let inventory = 'old';
    const oldPage = deferred<AssetBrowsePage>();
    const newPage = deferred<AssetBrowsePage>();
    let aborted = false;
    let catalogRequests = 0;
    const requests: string[] = [];
    const props = propsFor({
      searchAssetsQuery: new SearchAssetsQuery({ browseAssets: (input) => {
        requests.push(inventory);
        if (inventory === 'old') {
          input.signal?.addEventListener('abort', () => { aborted = true; });
          return oldPage.promise;
        }
        return newPage.promise;
      } }),
      inventoryContextQuery: { execute: async () => ({ inventoryName: inventory, canAdd: true }) },
      inventoryAssetTagsQuery: { execute: async () => [{ id: inventory, key: inventory, label: inventory }] },
      locationsQuery: { execute: async () => { catalogRequests++; return { inventoryName: inventory, tenantName: 'Tenant', canAdd: true, locations: [] }; } },
    });
    const criteria = { query: '', scope: 'all' as const, lifecycleState: 'active' as const, checkoutState: 'any' as const, sort: 'updated_desc' as const, tagIds: [] };
    const oldKey = mobileQueryKeys.browsePages('scope', 'tenant', 'old', criteria);
    client.setQueryData(oldKey, { pages: [{ ...criteria, criteria, kind: 'all', mode: 'browse', assets: [toAssetCardViewModel(asset('Old inventory item'))], hasMore: false }], pageParams: [undefined] });
    await client.invalidateQueries({ queryKey: oldKey, refetchType: 'none' });
    try {
      await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: inventory })}>
        <SearchScreen {...props} />
      </MobileServerStateProvider>);
      await settle(harness); await settle(harness);
      expect(harness.allText()).toContain('Old inventory item');
      await harness.run(() => {
        inventory = 'new';
        client.setQueryData(mobileQueryKeys.inventoryScope('scope'), { tenantId: 'tenant', inventoryId: inventory });
      });
      await settle(harness);
      expect(aborted).toBe(true);
      expect(harness.allText()).not.toContain('Old inventory item');
      await harness.run(() => newPage.resolve({ assets: [asset('New inventory item')], hasMore: false }));
      await settle(harness);
      await harness.run(() => oldPage.resolve({ assets: [asset('Old inventory item')], hasMore: false }));
      await settle(harness);
      expect(harness.allText()).toContain('New inventory item');
      expect(harness.allText()).not.toContain('Old inventory item');
      expect(requests).toEqual(['old', 'new']);
      expect(catalogRequests).toBe(0);
      await harness.press(harness.byLabel('Filters'));
      expect(harness.byLabel('Filter by tag new')).toBeDefined();
      expect(harness.byLabel('Filter by tag old')).toBeUndefined();
    } finally { await harness.unmount(); }
  });
  it('keeps Places rows available when summaries fail and retries only summaries', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    let pages = 0;
    let summaries = 0;
    const props = propsFor({
      initialScope: 'places',
      searchAssetsQuery: new SearchAssetsQuery({ browseAssets: async () => {
        pages++; return { assets: [{ ...asset('Garage'), kind: 'location' }], hasMore: false };
      } }),
      locationsQuery: { execute: async () => {
        summaries++;
        if (summaries === 1) throw new Error('Summary service unavailable');
        return { inventoryName: 'Home', tenantName: 'Tenant', canAdd: true, locations: [{ id: 'Garage', title: 'Garage', description: '', containedAssetCountLabel: '2 assets', recentAssetLabel: '', photoLabel: 'Needs photo' }] };
      } }
    });
    try {
      await harness.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <SearchScreen {...props} />
      </MobileServerStateProvider>);
      await settle(harness); await settle(harness);
      expect(harness.allText()).toContain('Garage');
      expect(harness.allText()).toContain('Place summaries could not load. Your places are still available.');
      const retry = harness.allByType('Pressable').find((node) => node.queryAll((child) => child.type === 'Text' && child.children.includes('Retry')).length > 0);
      await harness.press(retry);
      await settle(harness);
      expect(harness.allText()).toContain('2 assets');
      expect(harness.allText()).not.toContain('Place summaries could not load. Your places are still available.');
      expect({ pages, summaries }).toEqual({ pages: 1, summaries: 2 });
    } finally { await harness.unmount(); }
  });

  it('preserves the last successful filter result through replacement failure without appending its next page', async () => {
    const client = createMobileQueryClient();
    const harness = new MobileRenderHarness();
    const replacement = deferred<AssetBrowsePage>();
    const requests: string[] = [];
    const props = propsFor({ searchAssetsQuery: new SearchAssetsQuery({ browseAssets: async (input) => {
      requests.push(`${input.query}:${input.cursor ?? 'first'}`);
      return input.query ? replacement.promise : { assets: [asset('Tent')], hasMore: true, nextCursor: 'next' };
    } }) });
    const render = (initialQuery: string) => harness.render(
      <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
        <SearchScreen {...props} initialQuery={initialQuery} />
      </MobileServerStateProvider>
    );
    try {
      await render('');
      await settle(harness); await settle(harness);
      await render('broken');
      expect(harness.allText()).toContain('Tent');
      await harness.run(() => harness.byType('FlatList')?.props.onEndReached());
      await harness.run(() => replacement.reject(new Error('Search unavailable')));
      await settle(harness);
      expect(harness.allText()).toContain('Tent');
      expect(harness.allText()).toContain('This inventory could not be loaded.');
      await harness.run(() => harness.byType('FlatList')?.props.onEndReached());
      expect(requests).toEqual([':first', 'broken:first']);
    } finally { await harness.unmount(); }
  });

});

it('offers continuation rather than a false empty state for sparse filtered pages', async () => {
  const client = createMobileQueryClient(); const h = new MobileRenderHarness(); let reads = 0;
  const props = propsFor({ searchAssetsQuery: new SearchAssetsQuery({ browseAssets: async input => { reads++; return input.cursor ? { assets: [asset('Matching item')], hasMore: false } : { assets: [], hasMore: true, nextCursor: 'next' }; } }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><SearchScreen {...props} /></MobileServerStateProvider>); await settle(h); await settle(h);
    expect(h.allText()).toContain('No matching items in the pages loaded so far.'); expect(reads).toBe(1);
    await h.press(h.byLabel('Continue loading results')); await settle(h); expect(h.allText()).toContain('Matching item'); expect(reads).toBe(2);
  } finally { await h.unmount(); }
});

it('does not reuse previous filter rows after an authorization denial', async () => {
  const client = createMobileQueryClient(); const h = new MobileRenderHarness(); let status = 0;
  const props = propsFor({ searchAssetsQuery: new SearchAssetsQuery({ browseAssets: async () => { if (status) throw Object.assign(new Error('unavailable'), { status }); return { assets: [asset('Private row')], hasMore: false }; } }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><SearchScreen {...props} /></MobileServerStateProvider>); await settle(h); await settle(h); expect(h.allText()).toContain('Private row');
    status = 403; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.inventory('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).not.toContain('Private row');
    status = 500; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.inventory('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).not.toContain('Private row');
    status = 0; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.inventory('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).toContain('Private row');
  } finally { await h.unmount(); }
});

it('retries unavailable context before showing cached Browse results again', async () => {
  const client = createMobileQueryClient(); const h = new MobileRenderHarness(); let status = 0;
  const props = propsFor({ inventoryContextQuery: { execute: async () => { if (status) throw Object.assign(new Error('unavailable'), { status }); return { inventoryName: 'Home', canAdd: true }; } }, searchAssetsQuery: new SearchAssetsQuery({ browseAssets: async () => ({ assets: [asset('Verified row')], hasMore: false }) }) });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><SearchScreen {...props} /></MobileServerStateProvider>); await settle(h); await settle(h);
    expect(h.allText()).toContain('Verified row'); status = 403;
    await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.inventoryContext('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).not.toContain('Verified row');
    status = 500; await h.run(() => client.invalidateQueries({ queryKey: mobileQueryKeys.inventoryContext('scope', 'tenant', 'inventory') })); await settle(h); expect(h.allText()).not.toContain('Verified row');
    status = 0;
    const retry = h.allByType('Pressable').find(node => node.queryAll(child => child.type === 'Text' && child.children.includes('Retry')).length > 0);
    await h.press(retry!); await settle(h); expect(h.allText()).toContain('Verified row');
  } finally { await h.unmount(); }
});
