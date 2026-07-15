import { describe, expect, it } from 'vitest';
import type { Asset, SearchRequest, SearchResult } from '$lib/domain/inventory';
import {
  buildSearchSuggestions,
  executeWorkspaceSearch,
	  searchAssetHref,
	  searchCheckoutFilterOptions,
	  searchFilterHref,
  searchLifecycleFilterOptions,
  searchMatchFieldLabel,
  searchModeFilterOptions,
  searchPanelStatus
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

function searchResult(assetResult: Asset): SearchResult {
  return {
    type: 'asset',
    asset: assetResult,
    inventory: { id: assetResult.inventoryId, name: 'Household' },
    matches: [{ field: 'title', value: assetResult.title }]
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
    expect(searchAssetHref(asset('drill', 'Cordless drill', 'Power tool'))).toBe('/tenants/tenant-home/inventories/inventory-household/assets/drill');
    expect(searchAssetHref({ ...asset('garage', 'Garage', 'Place'), kind: 'location' })).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage'
    );
  });

  it('derives canonical hrefs for search filters', () => {
    expect(searchFilterHref('tenant-home', 'inventory-household', 'garage shelf', 'archived', 'fuzzy', 'any')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived'
    );
    expect(searchFilterHref('tenant-home', 'inventory-household', 'garage shelf', 'active', 'exact', 'any')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&mode=exact'
    );
    expect(searchFilterHref('tenant-home', 'inventory-household', '', 'all', 'fuzzy', 'checked_out')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/browse?lifecycle=all&availability=checked_out'
    );
    expect(searchFilterHref('tenant-home', 'inventory-household', '', 'all', 'fuzzy', 'any')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/browse?lifecycle=all'
    );
  });

  it('builds route-backed search filter options with stable labels', () => {
    expect(
      searchLifecycleFilterOptions({
        tenantId: 'tenant-home',
	        inventoryId: 'inventory-household',
	        query: 'garage shelf',
	        mode: 'fuzzy',
	        checkoutState: 'any'
      })
    ).toEqual([
      {
        value: 'active',
        label: 'Active',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf'
      },
      {
        value: 'archived',
        label: 'Archived',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived'
      },
      {
        value: 'all',
        label: 'All',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=all'
      }
    ]);
    expect(
      searchModeFilterOptions({
        tenantId: 'tenant-home',
	        inventoryId: 'inventory-household',
	        query: 'garage shelf',
	        lifecycleState: 'archived',
	        checkoutState: 'any'
      })
    ).toEqual([
      {
        value: 'fuzzy',
        label: 'Contains',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived'
      },
      {
        value: 'exact',
        label: 'Exact',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived&mode=exact'
	      }
	    ]);
    expect(
      searchCheckoutFilterOptions({
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        query: 'garage shelf',
        lifecycleState: 'archived',
        mode: 'exact'
      })
    ).toEqual([
      {
        value: 'any',
        label: 'Any',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived&mode=exact'
      },
      {
        value: 'checked_out',
        label: 'Checked out',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived&availability=checked_out&mode=exact'
      },
      {
        value: 'available',
        label: 'Available',
        href: '/tenants/tenant-home/inventories/inventory-household/browse?q=garage+shelf&lifecycle=archived&availability=available&mode=exact'
      }
    ]);
  });

  it('builds search panel status presentation for transient and empty states', () => {
    expect(searchPanelStatus({ error: 'Search service unavailable.', busy: false, submitted: true, query: 'box', resultCount: 0, lifecycleState: 'active' })).toEqual({
      kind: 'error',
      title: 'Search failed',
      message: 'Search service unavailable.',
      role: 'alert'
    });
    expect(searchPanelStatus({ error: '', busy: true, submitted: false, query: 'box', resultCount: 0, lifecycleState: 'active' })).toEqual({
      kind: 'busy',
      title: 'Searching',
      message: '',
      role: 'status'
    });
    expect(searchPanelStatus({ error: '', busy: false, submitted: false, query: '', resultCount: 0, lifecycleState: 'active' })).toEqual({
      kind: 'first-run',
      title: 'Search this inventory',
      message: 'Use asset, location, container, custom field, or attachment terms.'
    });
    expect(searchPanelStatus({ error: '', busy: false, submitted: true, query: '  box  ', resultCount: 0, lifecycleState: 'archived' })).toEqual({
      kind: 'empty',
      title: 'No results for "box"',
      message: 'No authorized archived assets matched this query.'
    });
    expect(searchPanelStatus({ error: '', busy: false, submitted: true, query: '', resultCount: 0, lifecycleState: 'all' })).toEqual({
      kind: 'empty',
      title: 'No results',
      message: 'No authorized assets matched this query.'
    });
    expect(searchPanelStatus({ error: '', busy: false, submitted: true, query: 'box', resultCount: 1, lifecycleState: 'all' }).kind).toBe('none');
  });

  it('labels tag-backed search match fields for users', () => {
    expect(searchMatchFieldLabel('tag_display_name')).toBe('Tag');
    expect(searchMatchFieldLabel('tag_key')).toBe('Tag');
    expect(searchMatchFieldLabel('title')).toBe('Title');
    expect(searchMatchFieldLabel('attachment_file_name')).toBe('Attachment');
    expect(searchMatchFieldLabel(undefined)).toBe('Match');
  });

  it('normalizes blank searches without calling the repository', async () => {
    const result = await executeWorkspaceSearch({
      repository: {
        searchAssets: async () => {
          throw new Error('should not search');
        }
      },
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: '   ',
      lifecycleState: 'active',
      mode: 'fuzzy',
      checkoutState: 'any'
    });

    expect(result).toEqual({ query: '', results: [], submitted: false, error: '' });
  });

  it('executes repository-backed search with trimmed query and explicit filters', async () => {
    const requests: SearchRequest[] = [];
    const result = await executeWorkspaceSearch({
      repository: {
        searchAssets: async (request) => {
          requests.push(request);
          return [searchResult(assets[0]!)];
        }
      },
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: ' tape ',
      lifecycleState: 'all',
      mode: 'exact',
      checkoutState: 'checked_out'
    });

    expect(requests).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        query: 'tape',
        lifecycleState: 'all',
        mode: 'exact',
        checkoutState: 'checked_out'
      }
    ]);
    expect(result).toMatchObject({ query: 'tape', results: [searchResult(assets[0]!)], submitted: true, error: '' });
  });

  it('keeps failed searches in a submitted state with a calm error message', async () => {
    const result = await executeWorkspaceSearch({
      repository: {
        searchAssets: async () => {
          throw new Error('Search service unavailable.');
        }
      },
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: 'drill',
      lifecycleState: 'active',
      mode: 'fuzzy',
      checkoutState: 'any'
    });

    expect(result).toEqual({
      query: 'drill',
      results: [],
      submitted: true,
      error: 'Search service unavailable.'
    });
  });

  it('replaces unsafe server search failures with a recovery-oriented message', async () => {
    const serverError = new Error('Internal server error.') as Error & { safeForUser: boolean; status: number };
    serverError.safeForUser = false;
    serverError.status = 500;

    const result = await executeWorkspaceSearch({
      repository: {
        searchAssets: async () => {
          throw serverError;
        }
      },
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: 'fertilizer',
      lifecycleState: 'active',
      mode: 'fuzzy',
      checkoutState: 'any'
    });

    expect(result).toEqual({
      query: 'fertilizer',
      results: [],
      submitted: true,
      error: 'Search could not complete. Try again, or check the server logs if it keeps happening.'
    });
  });
});
