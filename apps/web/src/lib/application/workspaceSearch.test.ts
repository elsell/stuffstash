import { describe, expect, it } from 'vitest';
import type { Asset, AssetTag } from '$lib/domain/inventory';
import type { BrowseAssetsRequest, InventoryBrowseRepository } from '$lib/ports/inventoryBrowseRepository';
import {
  browseEmptyPresentation,
  browseFilterOptions,
  browseFiltersAreDirty,
  browseFilterCount,
  browseSearchHref,
  browseSearchRoute,
  buildAppliedBrowseFilters,
  buildSearchSuggestions,
  executeBrowseSearch,
  filterBrowseAssets,
  loadBrowsePage,
  mergeBrowseSearchState,
  normalizeBrowseTagIds,
  searchAssetHref
} from './workspaceSearch';

const assets: Asset[] = [
  asset('tape', 'Tape measure', 'Garage measuring tool'),
  asset('labels', 'Shelf labels', 'Printed tags', 'Supply'),
  asset('drill', 'Cordless drill', '18V kit'),
  asset('archived', 'Old tape', 'Archived asset')
];

function asset(id: string, title: string, description: string, customAssetTypeLabel?: string): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'item',
    title,
    description,
    parentAssetId: null,
    lifecycleState: id === 'archived' ? 'archived' : 'active',
    customAssetTypeLabel
  };
}

describe('workspace search helpers', () => {
  it('builds bounded suggestions from visible asset title, description, and custom type labels', () => {
    expect(buildSearchSuggestions(assets, 'ta').map((suggestion) => suggestion.id)).toEqual(['tape', 'archived', 'labels']);
    expect(buildSearchSuggestions(assets, 'sup').map((suggestion) => suggestion.id)).toEqual(['labels']);
    expect(buildSearchSuggestions(assets, ' ', 2)).toEqual([]);
  });

  it('ranks suggestions by title strength before looser metadata matches', () => {
    expect(
      buildSearchSuggestions(
        [
          asset('description-match', 'Drill charger', 'garage tape'),
          asset('contains-title', 'Blue tape roll', ''),
          asset('exact-title', 'Tape', ''),
          asset('type-match', 'Packing labels', '', 'Tape supplies'),
          asset('prefix-title', 'Tape measure', '')
        ],
        'tape'
      ).map((suggestion) => suggestion.id)
    ).toEqual(['exact-title', 'prefix-title', 'contains-title', 'description-match', 'type-match']);
  });

  it('derives canonical hrefs for asset and location search hits', () => {
    expect(searchAssetHref(asset('drill', 'Cordless drill', 'Power tool'))).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/drill'
    );
    expect(searchAssetHref({ ...asset('garage', 'Garage', 'Place'), kind: 'location' })).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage'
    );
  });

  it('normalizes Browse search navigation and derives its canonical href', () => {
    const route = browseSearchRoute('tenant-home', 'inventory-household', {
      query: '  drill  ',
      lifecycleState: 'all',
      mode: 'exact',
      checkoutState: 'available',
      scope: 'items',
      sort: 'id_asc',
      selectedTagIds: ['tag-tools']
    });

    expect(route).toMatchObject({
      mode: 'browse',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      searchQuery: 'drill',
      searchLifecycleState: 'all',
      searchMode: 'exact',
      searchCheckoutState: 'available',
      browseSurface: 'list',
      browseScope: 'items',
      browseSort: 'id_asc',
      browseTagIds: ['tag-tools']
    });
    expect(browseSearchHref('tenant-home', 'inventory-household', {
      query: '  drill  ', lifecycleState: 'all', mode: 'exact', checkoutState: 'available',
      scope: 'items', sort: 'id_asc', selectedTagIds: ['tag-tools']
    })).toBe(
      '/tenants/tenant-home/inventories/inventory-household/browse?scope=items&q=drill&tag=tag-tools&lifecycle=all&availability=available&sort=id_asc&mode=exact'
    );
  });

  it('executes Browse requests through the repository boundary with canonical request state', async () => {
    const requests: BrowseAssetsRequest[] = [];
    const repository: InventoryBrowseRepository = {
      async browseAssets(request) {
        requests.push(request);
        return { assets: [], searchResults: [], nextCursor: null, hasMore: false };
      },
      async hasAnyAssets() { return false; },
      async loadActiveContainmentMap() { return []; }
    };

    await executeBrowseSearch(repository, {
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: '  drill  ', tagIds: ['tag-tools'],
      lifecycleState: 'active', checkoutState: 'checked_out', scope: 'all', sort: 'updated_desc', mode: 'fuzzy', cursor: 'next'
    });

    expect(requests).toEqual([{
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'drill', tagIds: ['tag-tools'],
      lifecycleState: 'active', checkoutState: 'checked_out', scope: 'all', sort: 'updated_desc', mode: 'fuzzy',
      limit: 20, cursor: 'next'
    }]);
  });

  it('owns Browse filter options and applied-filter presentation outside components', () => {
    const tags: AssetTag[] = [
      { id: 'tag-z', key: 'zippers', displayName: 'Zippers' },
      { id: 'tag-ten', key: 'bin-10', displayName: 'Bin 10' },
      { id: 'tag-eight', key: 'bin-8', displayName: 'Bin 8' },
      { id: 'tag-a', key: 'adapters', displayName: 'Adapters' }
    ];

    expect(browseFilterOptions.scopes.map((option) => option.label)).toEqual(['All', 'Places', 'Containers', 'Items']);
    expect(buildAppliedBrowseFilters('archived', 'checked_out', ['tag-z', 'tag-ten', 'tag-eight', 'tag-a'], tags)).toEqual([
      { key: 'lifecycle', label: 'Status: Archived' },
      { key: 'availability', label: 'Availability: Checked out' },
      { key: 'tag:tag-a', label: 'Tag: Adapters' },
      { key: 'tag:tag-eight', label: 'Tag: Bin 8' },
      { key: 'tag:tag-ten', label: 'Tag: Bin 10' },
      { key: 'tag:tag-z', label: 'Tag: Zippers' }
    ]);

    expect(normalizeBrowseTagIds(['', 'tag-a', 'tag-a', 'missing'])).toEqual(['tag-a', 'missing']);
    expect(browseFilterCount('active', 'any', ['', 'tag-a', 'tag-a', 'missing'])).toBe(2);
    expect(buildAppliedBrowseFilters('active', 'any', ['', 'tag-a', 'tag-a', 'missing'], tags)).toEqual([
      { key: 'tag:tag-a', label: 'Tag: Adapters' },
      { key: 'tag:missing', label: 'Unavailable tag: missing' }
    ]);
  });

  it('derives Browse filtering, dirty state, route merging, and empty presentation outside components', () => {
    const tagged = { ...assets[0]!, tags: [{ id: 'tag-tools', key: 'tools', displayName: 'Tools' }] };
    const checkedOut = { ...assets[1]!, currentCheckout: { id: 'checkout-1', state: 'open' as const, checkedOutAt: '2026-07-15T00:00:00Z', checkedOutByPrincipalId: 'Alex' } };
    expect(filterBrowseAssets([tagged, checkedOut, assets[3]!], {
      scope: 'items', lifecycleState: 'active', checkoutState: 'available', selectedTagIds: ['tag-tools']
    })).toEqual([tagged]);
    expect(browseFiltersAreDirty('active', 'all', 'any', 'any', ['tag-tools'], ['tag-tools'])).toBe(true);
    expect(mergeBrowseSearchState({
      query: 'drill', lifecycleState: 'active', mode: 'fuzzy', checkoutState: 'any', surface: 'list',
      scope: 'all', sort: 'updated_desc', selectedTagIds: ['tag-tools']
    }, { scope: 'items', selectedTagIds: ['', 'tag-tools', 'tag-tools'] })).toMatchObject({
      query: 'drill', scope: 'items', selectedTagIds: ['tag-tools']
    });
    expect(browseEmptyPresentation(true, '', 'all', 'active', 'any', [], true)).toEqual({
      kind: 'inventory', title: 'No stuff here yet', description: 'Add an item or location to start this inventory.',
      showCreateActions: true, showClearSearch: false
    });
    expect(browseEmptyPresentation(false, ' drill ', 'items', 'active', 'any', [], true)).toMatchObject({
      kind: 'query', title: 'No results for “drill”', showClearSearch: true
    });
  });

  it('loads and merges Browse pages while deriving submitted and inventory-empty state', async () => {
    const requests: BrowseAssetsRequest[] = [];
    let emptinessChecks = 0;
    const repository: InventoryBrowseRepository = {
      async browseAssets(request) {
        requests.push(request);
        return request.cursor
          ? { assets: [assets[1]!], searchResults: [], nextCursor: null, hasMore: false }
          : { assets: [], searchResults: [], nextCursor: 'next', hasMore: true };
      },
      async hasAnyAssets() { emptinessChecks += 1; return false; },
      async loadActiveContainmentMap() { return []; }
    };

    const initial = await loadBrowsePage(repository, {
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: '', selectedTagIds: [], lifecycleState: 'active',
      checkoutState: 'any', scope: 'all', sort: 'updated_desc', mode: 'fuzzy', append: false,
      currentAssets: [assets[0]!], currentSearchResults: [], currentInventoryEmpty: false
    });
    expect(initial).toMatchObject({ assets: [], searchResults: [], nextCursor: 'next', hasMore: true, inventoryEmpty: true, submitted: false });
    expect(emptinessChecks).toBe(1);

    const appended = await loadBrowsePage(repository, {
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: ' tape ', selectedTagIds: ['tag-tools'], lifecycleState: 'active',
      checkoutState: 'any', scope: 'items', sort: 'updated_desc', mode: 'fuzzy', append: true, cursor: 'next',
      currentAssets: [assets[0]!], currentSearchResults: [], currentInventoryEmpty: true
    });
    expect(appended).toMatchObject({ assets: [assets[0], assets[1]], inventoryEmpty: false, submitted: true });
    expect(emptinessChecks).toBe(1);
  });
});
