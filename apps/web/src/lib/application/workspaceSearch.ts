import type {
  Asset, AssetTag, BrowseScope, BrowseSort, BrowseSurface, SearchCheckoutFilter,
  SearchLifecycleFilter, SearchMode, SearchResult
} from '$lib/domain/inventory';
import type { BrowseAssetsRequest, InventoryBrowseRepository } from '$lib/ports/inventoryBrowseRepository';
import { defaultWorkspaceRoute, workspaceRouteHref, type WorkspaceRouteState } from './workspaceRoute';
import { filterAssets } from './workspace';
import { compareNaturalText } from './textCollation';

export interface BrowseSearchState {
  query: string;
  lifecycleState: SearchLifecycleFilter;
  mode: SearchMode;
  checkoutState: SearchCheckoutFilter;
  surface?: BrowseSurface;
  scope: BrowseScope;
  sort: BrowseSort;
  selectedTagIds: string[];
}

export type AppliedBrowseFilter = { key: string; label: string };

export type BrowseStateChange = Partial<Pick<
  BrowseSearchState,
  'surface' | 'scope' | 'lifecycleState' | 'checkoutState' | 'sort' | 'selectedTagIds'
>>;

export interface BrowsePageLoadInput {
  tenantId: string;
  inventoryId: string;
  query: string;
  selectedTagIds: string[];
  lifecycleState: SearchLifecycleFilter;
  checkoutState: SearchCheckoutFilter;
  scope: BrowseScope;
  sort: BrowseSort;
  mode: SearchMode;
  append: boolean;
  cursor?: string;
  currentAssets: Asset[];
  currentSearchResults: SearchResult[];
  currentInventoryEmpty: boolean;
}

export interface BrowsePageState {
  assets: Asset[];
  searchResults: SearchResult[];
  nextCursor: string | null;
  hasMore: boolean;
  inventoryEmpty: boolean;
  submitted: boolean;
}

export type BrowseEmptyPresentation = {
  kind: 'inventory' | 'query' | 'filters';
  title: string;
  description: string;
  showCreateActions: boolean;
  showClearSearch: boolean;
};

export const browseFilterOptions = {
  scopes: [
    { value: 'all', label: 'All' }, { value: 'places', label: 'Places' },
    { value: 'containers', label: 'Containers' }, { value: 'items', label: 'Items' }
  ] satisfies Array<{ value: BrowseScope; label: string }>,
  lifecycle: [
    { value: 'active', label: 'Active' }, { value: 'archived', label: 'Archived' },
    { value: 'all', label: 'All' }
  ] satisfies Array<{ value: SearchLifecycleFilter; label: string }>,
  availability: [
    { value: 'any', label: 'Any' }, { value: 'available', label: 'Available' },
    { value: 'checked_out', label: 'Checked out' }
  ] satisfies Array<{ value: SearchCheckoutFilter; label: string }>
} as const;

export function browseSearchRoute(
  tenantId: string | null,
  inventoryId: string | null,
  state: BrowseSearchState
): WorkspaceRouteState {
  return {
    ...defaultWorkspaceRoute,
    mode: 'browse', tenantId, inventoryId,
    searchQuery: state.query.trim(),
    searchLifecycleState: state.lifecycleState,
    searchMode: state.mode,
    searchCheckoutState: state.checkoutState,
    browseSurface: state.surface ?? 'list',
    browseScope: state.scope,
    browseSort: state.sort,
    browseTagIds: normalizeBrowseTagIds(state.selectedTagIds),
    compatibilityAlias: false
  };
}

export function browseSearchHref(
  tenantId: string | null,
  inventoryId: string | null,
  state: BrowseSearchState
): string {
  return workspaceRouteHref(browseSearchRoute(tenantId, inventoryId, state), tenantId, inventoryId);
}

export async function executeBrowseSearch(
  repository: InventoryBrowseRepository,
  request: Omit<BrowseAssetsRequest, 'limit' | 'query'> & { query: string; limit?: number }
) {
  return repository.browseAssets({
    ...request,
    query: request.query.trim(),
    limit: request.limit ?? 20
  });
}

export function normalizeBrowseTagIds(selectedTagIds: string[]): string[] {
  return Array.from(new Set(selectedTagIds.map((id) => id.trim()).filter(Boolean)));
}

export function mergeBrowseSearchState(current: BrowseSearchState, next: BrowseStateChange): BrowseSearchState {
  return {
    ...current,
    ...next,
    selectedTagIds: normalizeBrowseTagIds(next.selectedTagIds ?? current.selectedTagIds)
  };
}

export async function loadBrowsePage(
  repository: InventoryBrowseRepository,
  input: BrowsePageLoadInput
): Promise<BrowsePageState> {
  const selectedTagIds = normalizeBrowseTagIds(input.selectedTagIds);
  const page = await executeBrowseSearch(repository, {
    tenantId: input.tenantId,
    inventoryId: input.inventoryId,
    query: input.query,
    tagIds: selectedTagIds,
    lifecycleState: input.lifecycleState,
    checkoutState: input.checkoutState,
    scope: input.scope,
    sort: input.sort,
    mode: input.mode,
    cursor: input.cursor
  });
  const checksDefaultInventoryEmptiness = !input.append && !input.query.trim() && selectedTagIds.length === 0 &&
    input.scope === 'all' && input.lifecycleState === 'active' && input.checkoutState === 'any';
  const inventoryEmpty = checksDefaultInventoryEmptiness && page.assets.length === 0
    ? !(await repository.hasAnyAssets(input.tenantId, input.inventoryId))
    : page.assets.length > 0 ? false : input.currentInventoryEmpty;
  return {
    assets: input.append ? [...input.currentAssets, ...page.assets] : page.assets,
    searchResults: input.append ? [...input.currentSearchResults, ...page.searchResults] : page.searchResults,
    nextCursor: page.nextCursor,
    hasMore: page.hasMore,
    inventoryEmpty,
    submitted: !!input.query.trim() || selectedTagIds.length > 0
  };
}

export function filterBrowseAssets(
  assets: Asset[],
  filters: Pick<BrowseSearchState, 'scope' | 'lifecycleState' | 'checkoutState' | 'selectedTagIds'>
): Asset[] {
  const selected = normalizeBrowseTagIds(filters.selectedTagIds);
  return assets.filter((asset) => {
    if (filters.scope === 'places' && asset.kind !== 'location') return false;
    if (filters.scope === 'containers' && asset.kind !== 'container') return false;
    if (filters.scope === 'items' && asset.kind !== 'item') return false;
    if (filters.lifecycleState !== 'all' && asset.lifecycleState !== filters.lifecycleState) return false;
    if (filters.checkoutState === 'checked_out' && !asset.currentCheckout) return false;
    if (filters.checkoutState === 'available' && asset.currentCheckout) return false;
    if (selected.length > 0 && !selected.every((id) => asset.tags?.some((tag) => tag.id === id))) return false;
    return true;
  });
}

export function browseFilterCount(
  lifecycleState: SearchLifecycleFilter,
  checkoutState: SearchCheckoutFilter,
  selectedTagIds: string[]
): number {
  return (lifecycleState === 'active' ? 0 : 1) + (checkoutState === 'any' ? 0 : 1) +
    normalizeBrowseTagIds(selectedTagIds).length;
}

export function browseFiltersAreDirty(
  draftLifecycleState: SearchLifecycleFilter,
  lifecycleState: SearchLifecycleFilter,
  draftCheckoutState: SearchCheckoutFilter,
  checkoutState: SearchCheckoutFilter,
  draftTagIds: string[],
  selectedTagIds: string[]
): boolean {
  return draftLifecycleState !== lifecycleState || draftCheckoutState !== checkoutState ||
    normalizeBrowseTagIds(draftTagIds).sort().join(',') !== normalizeBrowseTagIds(selectedTagIds).sort().join(',');
}

export function browseEmptyPresentation(
  inventoryEmpty: boolean,
  query: string,
  scope: BrowseScope,
  lifecycleState: SearchLifecycleFilter,
  checkoutState: SearchCheckoutFilter,
  selectedTagIds: string[],
  canCreateAsset: boolean
): BrowseEmptyPresentation {
  const normalizedQuery = query.trim();
  const defaultEmptyInventory = inventoryEmpty && !normalizedQuery && scope === 'all' && lifecycleState === 'active' &&
    checkoutState === 'any' && normalizeBrowseTagIds(selectedTagIds).length === 0;
  if (defaultEmptyInventory) {
    return {
      kind: 'inventory', title: 'No stuff here yet',
      description: canCreateAsset ? 'Add an item or location to start this inventory.' : 'This inventory is empty.',
      showCreateActions: canCreateAsset, showClearSearch: false
    };
  }
  return normalizedQuery
    ? { kind: 'query', title: `No results for “${normalizedQuery}”`, description: 'Try another search term or clear a filter.', showCreateActions: false, showClearSearch: true }
    : { kind: 'filters', title: 'Nothing matches these filters', description: 'Try another scope or clear a filter.', showCreateActions: false, showClearSearch: false };
}

export function buildAppliedBrowseFilters(
  lifecycleState: SearchLifecycleFilter,
  checkoutState: SearchCheckoutFilter,
  selectedTagIds: string[],
  assetTags: AssetTag[]
): AppliedBrowseFilter[] {
  const lifecycleLabel = browseFilterOptions.lifecycle.find((option) => option.value === lifecycleState)?.label;
  const availabilityLabel = browseFilterOptions.availability.find((option) => option.value === checkoutState)?.label;
  const normalizedTagIds = normalizeBrowseTagIds(selectedTagIds);
  const tagsById = new Map(assetTags.map((tag) => [tag.id, tag]));
  const selectedTags = assetTags
    .filter((tag) => normalizedTagIds.includes(tag.id))
    .sort((left, right) => compareNaturalText(left.displayName, right.displayName));
  const unavailableTagIds = normalizedTagIds.filter((id) => !tagsById.has(id)).sort(compareNaturalText);
  return [
    ...(lifecycleState === 'active' ? [] : [{ key: 'lifecycle', label: `Status: ${lifecycleLabel}` }]),
    ...(checkoutState === 'any' ? [] : [{ key: 'availability', label: `Availability: ${availabilityLabel}` }]),
    ...selectedTags.map((tag) => ({ key: `tag:${tag.id}`, label: `Tag: ${tag.displayName}` })),
    ...unavailableTagIds.map((id) => ({ key: `tag:${id}`, label: `Unavailable tag: ${id}` }))
  ];
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

  return workspaceRouteHref(
    { mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id },
    asset.tenantId,
    asset.inventoryId
  );
}
