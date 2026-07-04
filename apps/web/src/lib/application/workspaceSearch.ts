import type { Asset, SearchLifecycleFilter, SearchMode, SearchRequest, SearchResult } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import { workspaceRouteHref } from './workspaceRoute';
import { filterAssets } from './workspace';

type SearchRepository = Pick<InventoryRepository, 'searchAssets'>;

export interface ExecuteWorkspaceSearchInput {
  repository: SearchRepository;
  tenantId: string | null;
  inventoryId: string | null;
  query: string;
  lifecycleState: SearchLifecycleFilter;
  mode: SearchMode;
}

export interface WorkspaceSearchResultState {
  query: string;
  results: SearchResult[];
  submitted: boolean;
  error: string;
}

export function buildSearchSuggestions(assets: Asset[], query: string, limit = 6): Asset[] {
  return filterAssets(assets, query).slice(0, limit);
}

export function searchAssetHref(asset: Asset): string {
  if (asset.kind === 'location') {
    return workspaceRouteHref(
      { mode: 'location', tenantId: asset.tenantId, inventoryId: asset.inventoryId, locationId: asset.id },
      asset.tenantId,
      asset.inventoryId
    );
  }

  return workspaceRouteHref({ mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id }, asset.tenantId, asset.inventoryId);
}

export function searchFilterHref(
  tenantId: string,
  inventoryId: string,
  query: string,
  lifecycleState: SearchLifecycleFilter,
  mode: SearchMode
): string {
  return workspaceRouteHref({ mode: 'search', tenantId, inventoryId, searchQuery: query, searchLifecycleState: lifecycleState, searchMode: mode }, tenantId, inventoryId);
}

export async function executeWorkspaceSearch(input: ExecuteWorkspaceSearchInput): Promise<WorkspaceSearchResultState> {
  const query = input.query.trim();
  if (!query || !input.tenantId || !input.inventoryId) {
    return { query, results: [], submitted: false, error: '' };
  }

  const request: SearchRequest = {
    tenantId: input.tenantId,
    inventoryId: input.inventoryId,
    query,
    lifecycleState: input.lifecycleState,
    mode: input.mode
  };

  try {
    return {
      query,
      results: await input.repository.searchAssets(request),
      submitted: true,
      error: ''
    };
  } catch (caught) {
    return {
      query,
      results: [],
      submitted: true,
      error: caught instanceof Error ? caught.message : 'Search failed.'
    };
  }
}
