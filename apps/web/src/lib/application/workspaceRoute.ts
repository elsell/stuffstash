import type {
  AssetKind,
  AssetLifecycleFilter,
  AuditScope,
  InvitationStatusFilter,
  SearchCheckoutFilter,
  SearchLifecycleFilter,
  SearchMode,
  WorkspaceMode
} from '$lib/domain/inventory';

export type WorkspaceAction = 'add' | 'edit' | null;
export type AssetRouteAction = 'edit' | 'move' | 'archive' | 'restore' | 'delete' | 'checkout' | 'return' | null;
export type AttachmentRouteAction = 'delete' | null;
export type AccessInvitationRouteAction = 'expire' | 'cancel' | 'delete' | null;
export type CustomizationRouteAction = 'archive_asset_type' | 'archive_field_definition' | null;
export type SettingsSection = 'overview' | 'access' | 'fields' | 'activity' | 'administration';
export type ImportSourceRoute = 'homebox' | 'homebox-csv' | null;
export type ImportDetailTabRoute = 'overview' | 'issues' | 'plan' | 'records' | 'timeline';

export interface WorkspaceRouteState {
  mode: WorkspaceMode;
  tenantId: string | null;
  inventoryId: string | null;
  locationId: string | null;
  assetId: string | null;
  action: WorkspaceAction;
  assetAction: AssetRouteAction;
  attachmentId: string | null;
  attachmentAction: AttachmentRouteAction;
  addKind: AssetKind | null;
  addParentAssetId: string | null;
  settingsSection: SettingsSection;
  invitationStatus: InvitationStatusFilter;
  accessInvitationAction: AccessInvitationRouteAction;
  accessInvitationId: string | null;
  auditScope: AuditScope;
  customizationAction: CustomizationRouteAction;
  customAssetTypeId: string | null;
  customFieldDefinitionId: string | null;
  importSource: ImportSourceRoute;
  importJobId: string | null;
  importTab: ImportDetailTabRoute | null;
  lifecycleState: AssetLifecycleFilter;
  searchQuery: string;
  searchLifecycleState: SearchLifecycleFilter;
  searchMode: SearchMode;
  searchCheckoutState: SearchCheckoutFilter;
}

export const defaultWorkspaceRoute: WorkspaceRouteState = {
  mode: 'home',
  tenantId: null,
  inventoryId: null,
  locationId: null,
  assetId: null,
  action: null,
  assetAction: null,
  attachmentId: null,
  attachmentAction: null,
  addKind: null,
  addParentAssetId: null,
  settingsSection: 'overview',
  invitationStatus: 'all',
  accessInvitationAction: null,
  accessInvitationId: null,
  auditScope: 'inventory',
  customizationAction: null,
  customAssetTypeId: null,
  customFieldDefinitionId: null,
  importSource: null,
  importJobId: null,
  importTab: null,
  lifecycleState: 'active',
  searchQuery: '',
  searchLifecycleState: 'active',
  searchMode: 'fuzzy',
  searchCheckoutState: 'any'
};

const assetKinds = new Set<AssetKind>(['item', 'container', 'location']);
const assetActions = new Set<AssetRouteAction>(['edit', 'move', 'archive', 'restore', 'delete', 'checkout', 'return']);
const attachmentActions = new Set<AttachmentRouteAction>(['delete']);
const accessInvitationActions = new Set<AccessInvitationRouteAction>(['expire', 'cancel', 'delete']);
const settingsSections = new Set<SettingsSection>(['overview', 'access', 'fields', 'activity', 'administration']);
const invitationStatuses = new Set<InvitationStatusFilter>(['all', 'pending', 'accepted', 'revoked', 'cancelled', 'expired']);
const auditScopes = new Set<AuditScope>(['inventory', 'tenant']);
const importSources = new Set<Exclude<ImportSourceRoute, null>>(['homebox', 'homebox-csv']);
const importDetailTabs = new Set<ImportDetailTabRoute>(['overview', 'issues', 'plan', 'records', 'timeline']);
const lifecycleFilters = new Set<AssetLifecycleFilter>(['active', 'archived']);
const searchLifecycleFilters = new Set<SearchLifecycleFilter>(['active', 'archived', 'all']);
const searchModes = new Set<SearchMode>(['fuzzy', 'exact']);
const searchCheckoutFilters = new Set<SearchCheckoutFilter>(['any', 'checked_out', 'available']);

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
    searchMode: parseSearchMode(url.searchParams.get('mode')),
    searchCheckoutState: parseSearchCheckout(url.searchParams.get('checkout'))
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
  const remaining = segments.length - inventoryOffset.nextIndex;
  if (section === 'browse' && remaining === 1) {
    return { ...route, mode: 'browse' };
  }
  if (section === 'locations' && remaining === 1) {
    return { ...route, mode: 'locations' };
  }
  if (section === 'locations' && (remaining === 2 || remaining === 3) && segments[inventoryOffset.nextIndex + 1]) {
    const action = parseLocationAction(segments[inventoryOffset.nextIndex + 2]);
    if (remaining === 3 && !action) {
      return route;
    }
    return {
      ...route,
      mode: action === 'edit' ? 'asset' : 'location',
      locationId: segments[inventoryOffset.nextIndex + 1],
      assetId: action === 'edit' ? segments[inventoryOffset.nextIndex + 1] : null,
      action,
      assetAction: action
    };
  }
  if (
    section === 'assets' &&
    remaining === 5 &&
    segments[inventoryOffset.nextIndex + 1] &&
    segments[inventoryOffset.nextIndex + 2] === 'attachments' &&
    segments[inventoryOffset.nextIndex + 3]
  ) {
    const attachmentAction = parseAttachmentAction(segments[inventoryOffset.nextIndex + 4]);
    if (!attachmentAction) {
      return route;
    }
    return {
      ...route,
      mode: 'asset',
      assetId: segments[inventoryOffset.nextIndex + 1],
      attachmentId: segments[inventoryOffset.nextIndex + 3],
      attachmentAction
    };
  }
  if (section === 'assets' && (remaining === 2 || remaining === 3) && segments[inventoryOffset.nextIndex + 1]) {
    const action = parseAssetAction(segments[inventoryOffset.nextIndex + 2]);
    if (remaining === 3 && !action) {
      return route;
    }
    return {
      ...route,
      mode: 'asset',
      assetId: segments[inventoryOffset.nextIndex + 1],
      action: action === 'edit' ? 'edit' : null,
      assetAction: action
    };
  }
  if (section === 'search') {
    return remaining === 1 ? { ...route, mode: 'search', lifecycleState: 'active' } : route;
  }
  if (section === 'settings') {
    if (remaining > 5) {
      return route;
    }
    const settingsSection = parseSettingsSection(segments[inventoryOffset.nextIndex + 1]);
    if (settingsSection === 'access' && remaining === 5) {
      const resource = segments[inventoryOffset.nextIndex + 2];
      const resourceId = segments[inventoryOffset.nextIndex + 3];
      const action = parseAccessInvitationAction(segments[inventoryOffset.nextIndex + 4]);
      if (resource === 'invitations' && resourceId && action) {
        return {
          ...route,
          mode: 'settings',
          settingsSection,
          invitationStatus: parseInvitationStatus(url.searchParams.get('invitationStatus')),
          accessInvitationAction: action,
          accessInvitationId: resourceId
        };
      }
      return route;
    }
    if (settingsSection === 'fields' && remaining === 5) {
      const resource = segments[inventoryOffset.nextIndex + 2];
      const resourceId = segments[inventoryOffset.nextIndex + 3];
      const action = segments[inventoryOffset.nextIndex + 4];
      if (resource === 'asset-types' && resourceId && action === 'archive') {
        return {
          ...route,
          mode: 'settings',
          settingsSection,
          customizationAction: 'archive_asset_type',
          customAssetTypeId: resourceId
        };
      }
      if (resource === 'field-definitions' && resourceId && action === 'archive') {
        return {
          ...route,
          mode: 'settings',
          settingsSection,
          customizationAction: 'archive_field_definition',
          customFieldDefinitionId: resourceId
        };
      }
      return route;
    }
    if (remaining > 2) {
      return route;
    }
    return {
      ...route,
      mode: 'settings',
      settingsSection,
      invitationStatus: settingsSection === 'access' ? parseInvitationStatus(url.searchParams.get('invitationStatus')) : 'all',
      auditScope: settingsSection === 'activity' ? parseAuditScope(url.searchParams.get('auditScope')) : 'inventory'
    };
  }
  if (section === 'import') {
    if (!route.tenantId) {
      return route;
    }
    if (remaining === 1) {
      return { ...route, mode: 'import', importSource: null };
    }
    if (remaining === 3 && segments[inventoryOffset.nextIndex + 1] === 'jobs' && segments[inventoryOffset.nextIndex + 2]) {
      return {
        ...route,
        mode: 'import',
        importSource: null,
        importJobId: segments[inventoryOffset.nextIndex + 2],
        importTab: parseImportTab(url.searchParams.get('tab'))
      };
    }
    if (remaining === 2 && importSources.has(segments[inventoryOffset.nextIndex + 1] as Exclude<ImportSourceRoute, null>)) {
      return {
        ...route,
        mode: 'import',
        importSource: segments[inventoryOffset.nextIndex + 1] as Exclude<ImportSourceRoute, null>
      };
    }
    return route;
  }
  if (section === 'add' && remaining === 2 && assetKinds.has(segments[inventoryOffset.nextIndex + 1] as AssetKind)) {
    return {
      ...route,
      action: 'add',
      addKind: segments[inventoryOffset.nextIndex + 1] as AssetKind,
      addParentAssetId: url.searchParams.get('parent') || null
    };
  }

  return route;
}

function parseAssetAction(value: string | undefined): AssetRouteAction {
  return assetActions.has(value as AssetRouteAction) ? (value as AssetRouteAction) : null;
}

function parseAttachmentAction(value: string | undefined): AttachmentRouteAction {
  return attachmentActions.has(value as AttachmentRouteAction) ? (value as AttachmentRouteAction) : null;
}

function parseAccessInvitationAction(value: string | undefined): AccessInvitationRouteAction {
  return accessInvitationActions.has(value as AccessInvitationRouteAction) ? (value as AccessInvitationRouteAction) : null;
}

function parseLocationAction(value: string | undefined): Extract<AssetRouteAction, 'edit'> | null {
  return value === 'edit' ? 'edit' : null;
}

function parseSettingsSection(value: string | undefined): SettingsSection {
  return settingsSections.has(value as SettingsSection) ? (value as SettingsSection) : 'overview';
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

  if (inventoryId && next.mode === 'browse') {
    path += '/browse';
    if (next.searchQuery.trim()) {
      search.set('q', next.searchQuery.trim());
    }
    if (next.searchLifecycleState !== 'active') {
      search.set('lifecycle', next.searchLifecycleState);
    }
    if (next.searchMode !== 'fuzzy') {
      search.set('mode', next.searchMode);
    }
    if (next.searchCheckoutState !== 'any') {
      search.set('checkout', next.searchCheckoutState);
    }
  } else if (inventoryId && next.mode === 'locations') {
    path += '/locations';
  } else if (inventoryId && next.mode === 'location' && next.locationId) {
    path += `/locations/${encodeURIComponent(next.locationId)}`;
    if ((next.assetAction ?? next.action) === 'edit') {
      path += '/edit';
    }
  } else if (inventoryId && next.mode === 'asset' && next.assetId) {
    const action = next.assetAction ?? (next.action === 'edit' ? 'edit' : null);
    const isLocationEdit = next.locationId === next.assetId && action === 'edit';
    if (isLocationEdit) {
      path += `/locations/${encodeURIComponent(next.assetId)}/edit`;
    } else {
      path += `/assets/${encodeURIComponent(next.assetId)}`;
    }
    if (next.attachmentId && next.attachmentAction) {
      path += `/attachments/${encodeURIComponent(next.attachmentId)}/${next.attachmentAction}`;
      const query = search.toString();
      return query ? `${path}?${query}` : path;
    }
    if (action && !isLocationEdit) {
      path += `/${action}`;
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
    if (next.searchCheckoutState !== 'any') {
      search.set('checkout', next.searchCheckoutState);
    }
  } else if (inventoryId && next.mode === 'settings') {
    path += '/settings';
    if (next.settingsSection !== 'overview') {
      path += `/${next.settingsSection}`;
    }
    if (next.settingsSection === 'access' && next.accessInvitationAction && next.accessInvitationId) {
      path += `/invitations/${encodeURIComponent(next.accessInvitationId)}/${next.accessInvitationAction}`;
    }
    if (next.settingsSection === 'fields' && next.customizationAction === 'archive_asset_type' && next.customAssetTypeId) {
      path += `/asset-types/${encodeURIComponent(next.customAssetTypeId)}/archive`;
    }
    if (
      next.settingsSection === 'fields' &&
      next.customizationAction === 'archive_field_definition' &&
      next.customFieldDefinitionId
    ) {
      path += `/field-definitions/${encodeURIComponent(next.customFieldDefinitionId)}/archive`;
    }
    if (next.settingsSection === 'access' && next.invitationStatus !== 'all') {
      search.set('invitationStatus', next.invitationStatus);
    }
    if (next.settingsSection === 'activity' && next.auditScope !== 'inventory') {
      search.set('auditScope', next.auditScope);
    }
  } else if (inventoryId && next.mode === 'import') {
    path += '/import';
    if (next.importJobId) {
      path += `/jobs/${encodeURIComponent(next.importJobId)}`;
      if (next.importTab) {
        search.set('tab', next.importTab);
      }
    } else if (next.importSource) {
      path += `/${next.importSource}`;
    }
  } else if (inventoryId && next.action === 'add' && next.addKind) {
    path += `/add/${next.addKind}`;
    if (next.addParentAssetId) {
      search.set('parent', next.addParentAssetId);
    }
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

function parseSearchCheckout(value: string | null): SearchCheckoutFilter {
  return searchCheckoutFilters.has(value as SearchCheckoutFilter) ? (value as SearchCheckoutFilter) : 'any';
}

function parseInvitationStatus(value: string | null): InvitationStatusFilter {
  return invitationStatuses.has(value as InvitationStatusFilter) ? (value as InvitationStatusFilter) : 'all';
}

function parseAuditScope(value: string | null): AuditScope {
  return auditScopes.has(value as AuditScope) ? (value as AuditScope) : 'inventory';
}

function parseImportTab(value: string | null): ImportDetailTabRoute | null {
  return importDetailTabs.has(value as ImportDetailTabRoute) ? (value as ImportDetailTabRoute) : null;
}
