import type { AssetKind, AssetLifecycleFilter, SearchLifecycleFilter, SearchMode, WorkspaceMode } from '$lib/domain/inventory';

export type WorkspaceAction = 'add' | 'edit' | null;

export interface WorkspaceRouteState {
  mode: WorkspaceMode;
  tenantId: string | null;
  inventoryId: string | null;
  locationId: string | null;
  assetId: string | null;
  action: WorkspaceAction;
  addKind: AssetKind | null;
  lifecycleState: AssetLifecycleFilter;
  searchQuery: string;
  searchLifecycleState: SearchLifecycleFilter;
  searchMode: SearchMode;
}

export const defaultWorkspaceRoute: WorkspaceRouteState = {
  mode: 'home',
  tenantId: null,
  inventoryId: null,
  locationId: null,
  assetId: null,
  action: null,
  addKind: null,
  lifecycleState: 'active',
  searchQuery: '',
  searchLifecycleState: 'active',
  searchMode: 'fuzzy'
};

const assetKinds = new Set<AssetKind>(['item', 'container', 'location']);
const lifecycleFilters = new Set<AssetLifecycleFilter>(['active', 'archived']);
const searchLifecycleFilters = new Set<SearchLifecycleFilter>(['active', 'archived', 'all']);
const searchModes = new Set<SearchMode>(['fuzzy', 'exact']);

export function parseWorkspaceRoute(url: URL): WorkspaceRouteState {
  const segments = safePathSegments(url.pathname);
  if (!segments) {
    return { ...defaultWorkspaceRoute };
  }
  const route = {
    ...defaultWorkspaceRoute,
    lifecycleState: parseLifecycle(url.searchParams.get('lifecycle')),
    searchLifecycleState: parseSearchLifecycle(url.searchParams.get('lifecycle')),
    searchQuery: url.searchParams.get('q') ?? '',
    searchMode: parseSearchMode(url.searchParams.get('mode'))
  };

  if (segments.length === 0) {
    return route;
  }

  const inventoryOffset = parseInventoryOffset(segments);
  if (!inventoryOffset) {
    return route;
  }

  route.tenantId = inventoryOffset.tenantId;
  route.inventoryId = inventoryOffset.inventoryId;
  if (segments.length === inventoryOffset.nextIndex) {
    return route;
  }

  const section = segments[inventoryOffset.nextIndex];
  if (section === 'locations' && segments[inventoryOffset.nextIndex + 1]) {
    return { ...route, mode: 'location', locationId: segments[inventoryOffset.nextIndex + 1] };
  }
  if (section === 'assets' && segments[inventoryOffset.nextIndex + 1]) {
    return {
      ...route,
      mode: 'asset',
      assetId: segments[inventoryOffset.nextIndex + 1],
      action: segments[inventoryOffset.nextIndex + 2] === 'edit' ? 'edit' : null
    };
  }
  if (section === 'search') {
    return { ...route, mode: 'search' };
  }
  if (section === 'settings') {
    return { ...route, mode: 'settings' };
  }
  if (section === 'add' && assetKinds.has(segments[inventoryOffset.nextIndex + 1] as AssetKind)) {
    return { ...route, action: 'add', addKind: segments[inventoryOffset.nextIndex + 1] as AssetKind };
  }

  return route;
}

function parseInventoryOffset(
  segments: string[]
): { tenantId: string | null; inventoryId: string; nextIndex: number } | null {
  if (segments[0] === 'tenants' && segments[1] && segments[2] === 'inventories' && segments[3]) {
    return { tenantId: segments[1], inventoryId: segments[3], nextIndex: 4 };
  }
  if (segments[0] === 'inventories' && segments[1]) {
    return { tenantId: null, inventoryId: segments[1], nextIndex: 2 };
  }
  return null;
}

function safePathSegments(pathname: string): string[] | null {
  try {
    return pathname.split('/').filter(Boolean).map(decodeURIComponent);
  } catch {
    return null;
  }
}

export function workspaceRouteHref(
  state: Partial<WorkspaceRouteState>,
  selectedTenantId: string | null,
  selectedInventoryId: string | null
): string {
  const next = { ...defaultWorkspaceRoute, ...state };
  const tenantId = next.tenantId ?? selectedTenantId;
  const inventoryId = next.inventoryId ?? selectedInventoryId;
  const search = new URLSearchParams();
  let path = '/';

  if (tenantId && inventoryId) {
    path = `/tenants/${encodeURIComponent(tenantId)}/inventories/${encodeURIComponent(inventoryId)}`;
  } else if (inventoryId) {
    path = `/inventories/${encodeURIComponent(inventoryId)}`;
  }

  if (inventoryId && next.mode === 'location' && next.locationId) {
    path += `/locations/${encodeURIComponent(next.locationId)}`;
  } else if (inventoryId && next.mode === 'asset' && next.assetId) {
    path += `/assets/${encodeURIComponent(next.assetId)}`;
    if (next.action === 'edit') {
      path += '/edit';
    }
  } else if (inventoryId && next.mode === 'search') {
    path += '/search';
    if (next.searchQuery.trim()) {
      search.set('q', next.searchQuery.trim());
    }
    if (next.searchLifecycleState !== 'active') {
      search.set('lifecycle', next.searchLifecycleState);
    }
    if (next.searchMode !== 'fuzzy') {
      search.set('mode', next.searchMode);
    }
  } else if (inventoryId && next.mode === 'settings') {
    path += '/settings';
  } else if (inventoryId && next.action === 'add' && next.addKind) {
    path += `/add/${next.addKind}`;
  } else if (next.lifecycleState !== 'active') {
    search.set('lifecycle', next.lifecycleState);
  }

  const query = search.toString();
  return query ? `${path}?${query}` : path;
}

function parseLifecycle(value: string | null): AssetLifecycleFilter {
  return lifecycleFilters.has(value as AssetLifecycleFilter) ? (value as AssetLifecycleFilter) : 'active';
}

function parseSearchLifecycle(value: string | null): SearchLifecycleFilter {
  return searchLifecycleFilters.has(value as SearchLifecycleFilter) ? (value as SearchLifecycleFilter) : 'active';
}

function parseSearchMode(value: string | null): SearchMode {
  return searchModes.has(value as SearchMode) ? (value as SearchMode) : 'fuzzy';
}
