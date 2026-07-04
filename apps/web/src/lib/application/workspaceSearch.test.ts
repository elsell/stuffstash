import { describe, expect, it } from 'vitest';
import type { Asset, SearchRequest, SearchResult } from '$lib/domain/inventory';
import { buildSearchSuggestions, executeWorkspaceSearch } from './workspaceSearch';

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
      mode: 'fuzzy'
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
      mode: 'exact'
    });

    expect(requests).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-household',
        query: 'tape',
        lifecycleState: 'all',
        mode: 'exact'
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
      mode: 'fuzzy'
    });

    expect(result).toEqual({
      query: 'drill',
      results: [],
      submitted: true,
      error: 'Search service unavailable.'
    });
  });
});
