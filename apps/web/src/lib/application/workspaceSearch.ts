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

export interface SearchFilterOption<TValue extends string = string> {
  value: TValue;
  label: string;
  href: string;
}

const searchLifecycleFilters: SearchLifecycleFilter[] = ['active', 'archived', 'all'];
const searchModes: SearchMode[] = ['fuzzy', 'exact'];

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

export function searchLifecycleFilterOptions(input: {
  tenantId: string;
  inventoryId: string;
  query: string;
  mode: SearchMode;
}): SearchFilterOption<SearchLifecycleFilter>[] {
  return searchLifecycleFilters.map((lifecycleState) => ({
    value: lifecycleState,
    label: searchLifecycleFilterLabel(lifecycleState),
    href: searchFilterHref(input.tenantId, input.inventoryId, input.query, lifecycleState, input.mode)
  }));
}

export function searchModeFilterOptions(input: {
  tenantId: string;
  inventoryId: string;
  query: string;
  lifecycleState: SearchLifecycleFilter;
}): SearchFilterOption<SearchMode>[] {
  return searchModes.map((mode) => ({
    value: mode,
    label: searchModeLabel(mode),
    href: searchFilterHref(input.tenantId, input.inventoryId, input.query, input.lifecycleState, mode)
  }));
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

function searchLifecycleFilterLabel(lifecycleState: SearchLifecycleFilter): string {
  if (lifecycleState === 'archived') {
    return 'Archived';
  }
  if (lifecycleState === 'all') {
    return 'All';
  }
  return 'Active';
}

function searchModeLabel(mode: SearchMode): string {
  return mode === 'exact' ? 'Exact' : 'Contains';
}
