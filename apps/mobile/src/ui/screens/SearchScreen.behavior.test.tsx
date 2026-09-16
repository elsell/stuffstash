import { setNativeHeaderHeight } from '../../test-support/react-navigation-elements';
import { navigationOptions, dispatchedActions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { AppFeedbackProvider } from '../feedback/AppFeedback';
import { InventoryMapQuery } from '../../application/assets/InventoryMapQuery';
import { tenantId, inventoryId } from '../../domain/inventories/InventorySummary';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { SearchScreen } from './SearchScreen';
import { MobileRenderHarness } from '../../test-support/render';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { SearchAssetsQuery } from '../../application/search/SearchAssetsQuery';
import { toAssetCardViewModel } from '../../application/assets/AssetViewModels';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import type { AssetBrowsePage } from '../../application/home/InventorySummaryRepository';

function latestNativeSearch() {
  // Other Stack.Screen calls update title/actions independently of search.
  // Read the latest explicit search update, including an explicit removal.
  const options = navigationOptions().findLast(value => Object.prototype.hasOwnProperty.call(value, 'headerSearchBarOptions')) as {
    headerSearchBarOptions?: { onChangeText: (event: { nativeEvent: { text: string } }) => void; onCancelButtonPress: () => void; placeholder: string }
  } | undefined;
  expect(options?.headerSearchBarOptions).toBeDefined();
  return options!.headerSearchBarOptions!;
}

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
      expect(dispatchedActions().filter(action => action.type === 'push').at(-1)).toMatchObject({
        href: { pathname: '/browse-filters', params: { inventoryId: 'new' } }
      });
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

it('keeps native search and refinements across an immediate List/Map switch',async()=>{
 const h=new MobileRenderHarness();const client=createMobileQueryClient();resetNavigation();
 const props=propsFor({initialTagIds:['tag'],inventoryMapQuery:new InventoryMapQuery({listActiveInventoryMapAssets:async()=>({sessionScopeId:'scope',tenantId:tenantId('tenant'),inventoryId:inventoryId('inventory'),inventoryName:'Home',permissions:['view'],assets:[]})})});
 const nativeSearch=latestNativeSearch;
 const switchTo=async(label:string)=>h.run(()=>h.allByType('NativeSegmentedControl').find(node=>node.props.values?.includes('Map'))?.props.onValueChange(label));
 try{
  await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async()=>({tenantId:'tenant',inventoryId:'inventory'})}><AppFeedbackProvider><SearchScreen {...props}/></AppFeedbackProvider></MobileServerStateProvider>);
  await settle(h);await settle(h);
  await h.run(()=>nativeSearch().onChangeText({nativeEvent:{text:'Tent'}}));
  await switchTo('Map');await settle(h);
  expect(nativeSearch().placeholder).toBe('Find and expand path');
  expect(h.byTestId('browse-map-frame')?.props.style).toContainEqual({ paddingTop: 144 });
  await h.run(() => setNativeHeaderHeight(210));
  expect(h.byTestId('browse-map-frame')?.props.style).toContainEqual({ paddingTop: 210 });
  await h.run(() => setNativeHeaderHeight(144));
  await h.run(()=>new Promise(resolve=>setTimeout(resolve,320)));
  expect(dispatchedActions().filter(action=>action.type==='setParams').at(-1)).toMatchObject({params:{surface:'map',query:'Tent',tagId:['tag']}});
  await switchTo('List');await settle(h);
  expect(nativeSearch().placeholder).toBe('Search names, places, or tags');
  expect(h.byLabel('Filters, 1 applied')).toBeDefined();
  await h.run(()=>nativeSearch().onCancelButtonPress());await settle(h);
 }finally{await h.unmount();}
});

it('settles pending search and opens a scoped native filter sheet', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); resetNavigation();
  const props = propsFor({ initialTagIds: ['tag'] });
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><SearchScreen {...props} /></MobileServerStateProvider>);
    await settle(h); await settle(h);
    const search = latestNativeSearch();
    await h.run(() => search.onChangeText({ nativeEvent: { text: 'Fresh query' } }));
    await h.press(h.byLabel('Filters, 1 applied'));
    expect(dispatchedActions().filter(action => action.type === 'push').at(-1)).toMatchObject({
      href: { pathname: '/browse-filters', params: { query: 'Fresh query', tagId: ['tag'], tenantId: 'tenant', inventoryId: 'inventory', sessionScope: 'scope' } }
    });
    const count = dispatchedActions().length;
    await h.run(() => new Promise(resolve => setTimeout(resolve, 320)));
    expect(dispatchedActions()).toHaveLength(count);
  } finally { await h.unmount(); }
});

it('pauses pending Browse query on blur and resumes it on return', async () => {
 const h=new MobileRenderHarness();const client=createMobileQueryClient();resetNavigation();setScreenFocused(true);
 const requests:string[]=[];const props=propsFor({searchAssetsQuery:new SearchAssetsQuery({browseAssets:async input=>{requests.push(input.query);return {assets:[],hasMore:false};}})});
 try {
  await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async()=>({tenantId:'tenant',inventoryId:'inventory'})}><AppFeedbackProvider><SearchScreen {...props}/></AppFeedbackProvider></MobileServerStateProvider>);await settle(h);await settle(h);
  const search=latestNativeSearch();
  await h.run(()=>search.onChangeText({nativeEvent:{text:'retained'}}));
  await h.run(()=>setScreenFocused(false));
  await h.run(()=>new Promise(resolve=>setTimeout(resolve,350)));
  expect(requests).not.toContain('retained');
  expect(dispatchedActions().filter(action=>action.type==='setParams')).toEqual([]);
  await h.run(()=>setScreenFocused(true));
  await h.run(()=>new Promise(resolve=>setTimeout(resolve,350)));await settle(h);
  expect(requests).toContain('retained');
 }finally{await h.unmount();client.clear();resetNavigation();setScreenFocused(true);}
});

it('uses external Browse criteria instead of a paused draft after returning', async () => {
 const h=new MobileRenderHarness();const client=createMobileQueryClient();resetNavigation();setScreenFocused(true);
 const requests:string[]=[];const props=propsFor({searchAssetsQuery:new SearchAssetsQuery({browseAssets:async input=>{requests.push(input.query);return {assets:[],hasMore:false};}})});
 const view=(initialQuery:string)=><MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async()=>({tenantId:'tenant',inventoryId:'inventory'})}><AppFeedbackProvider><SearchScreen {...props} initialQuery={initialQuery}/></AppFeedbackProvider></MobileServerStateProvider>;
 try {
  await h.render(view(''));await settle(h);await settle(h);
  const search=latestNativeSearch();
  await h.run(()=>search.onChangeText({nativeEvent:{text:'abandoned'}}));
  await h.run(()=>setScreenFocused(false));
  await h.render(view('replacement'));await settle(h);
  await h.run(()=>setScreenFocused(true));
  await h.run(()=>new Promise(resolve=>setTimeout(resolve,350)));await settle(h);
  expect(requests).toContain('replacement');expect(requests).not.toContain('abandoned');
  expect(dispatchedActions().filter(action=>action.type==='setParams')).toEqual([]);
 }finally{await h.unmount();client.clear();resetNavigation();setScreenFocused(true);}
});
