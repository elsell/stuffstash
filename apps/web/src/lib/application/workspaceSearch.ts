import type { Asset, SearchCheckoutFilter, SearchLifecycleFilter, SearchMode, SearchRequest, SearchResult } from '$lib/domain/inventory';
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
  checkoutState: SearchCheckoutFilter;
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

export type SearchPanelStatusKind = 'none' | 'error' | 'busy' | 'first-run' | 'empty';

export interface SearchPanelStatusPresentation {
  kind: SearchPanelStatusKind;
  title: string;
  message: string;
  role?: 'alert' | 'status';
}

const searchLifecycleFilters: SearchLifecycleFilter[] = ['active', 'archived', 'all'];
const searchModes: SearchMode[] = ['fuzzy', 'exact'];
const searchCheckoutFilters: SearchCheckoutFilter[] = ['any', 'checked_out', 'available'];

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
  mode: SearchMode,
  checkoutState: SearchCheckoutFilter
): string {
  return workspaceRouteHref(
    { mode: 'search', tenantId, inventoryId, searchQuery: query, searchLifecycleState: lifecycleState, searchMode: mode, searchCheckoutState: checkoutState },
    tenantId,
    inventoryId
  );
}

export function searchLifecycleFilterOptions(input: {
  tenantId: string;
  inventoryId: string;
  query: string;
  mode: SearchMode;
  checkoutState: SearchCheckoutFilter;
}): SearchFilterOption<SearchLifecycleFilter>[] {
  return searchLifecycleFilters.map((lifecycleState) => ({
    value: lifecycleState,
    label: searchLifecycleFilterLabel(lifecycleState),
    href: searchFilterHref(input.tenantId, input.inventoryId, input.query, lifecycleState, input.mode, input.checkoutState)
  }));
}

export function searchModeFilterOptions(input: {
  tenantId: string;
  inventoryId: string;
  query: string;
  lifecycleState: SearchLifecycleFilter;
  checkoutState: SearchCheckoutFilter;
}): SearchFilterOption<SearchMode>[] {
  return searchModes.map((mode) => ({
    value: mode,
    label: searchModeLabel(mode),
    href: searchFilterHref(input.tenantId, input.inventoryId, input.query, input.lifecycleState, mode, input.checkoutState)
  }));
}

export function searchCheckoutFilterOptions(input: {
  tenantId: string;
  inventoryId: string;
  query: string;
  lifecycleState: SearchLifecycleFilter;
  mode: SearchMode;
}): SearchFilterOption<SearchCheckoutFilter>[] {
  return searchCheckoutFilters.map((checkoutState) => ({
    value: checkoutState,
    label: searchCheckoutFilterLabel(checkoutState),
    href: searchFilterHref(input.tenantId, input.inventoryId, input.query, input.lifecycleState, input.mode, checkoutState)
  }));
}

export function searchPanelStatus(input: {
  error: string;
  busy: boolean;
  submitted: boolean;
  query: string;
  resultCount: number;
  lifecycleState: SearchLifecycleFilter;
}): SearchPanelStatusPresentation {
  if (input.error) {
    return { kind: 'error', title: 'Search failed', message: input.error, role: 'alert' };
  }
  if (input.busy) {
    return { kind: 'busy', title: 'Searching', message: '', role: 'status' };
  }
  if (!input.submitted) {
    return {
      kind: 'first-run',
      title: 'Search this inventory',
      message: 'Use asset, location, container, custom field, or attachment terms.'
    };
  }
  if (input.resultCount === 0) {
    const query = input.query.trim();
    const querySuffix = query ? ` for "${query}"` : '';
    return {
      kind: 'empty',
      title: `No results${querySuffix}`,
      message:
        input.lifecycleState === 'all'
          ? 'No authorized assets matched this query.'
          : `No authorized ${input.lifecycleState} assets matched this query.`
    };
  }
  return { kind: 'none', title: '', message: '' };
}

export function searchMatchFieldLabel(field: string | undefined): string {
  switch (field) {
    case 'tag_display_name':
    case 'tag_key':
      return 'Tag';
    case 'title':
      return 'Title';
    case 'description':
      return 'Description';
    case 'custom_field':
      return 'Custom field';
    case 'custom_asset_type_key':
    case 'custom_asset_type_name':
      return 'Asset type';
    case 'attachment_file_name':
      return 'Attachment';
    case 'attachment_content_type':
      return 'Attachment type';
    case undefined:
    case '':
      return 'Match';
    default:
      return field
        .split('_')
        .filter(Boolean)
        .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
        .join(' ') || 'Match';
  }
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
    mode: input.mode,
    checkoutState: input.checkoutState
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

function searchCheckoutFilterLabel(checkoutState: SearchCheckoutFilter): string {
  if (checkoutState === 'checked_out') {
    return 'Checked out';
  }
  if (checkoutState === 'available') {
    return 'Available';
  }
  return 'Any';
}
