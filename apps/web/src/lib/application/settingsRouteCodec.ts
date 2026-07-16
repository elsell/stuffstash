import type { AuditScope, CustomDefinitionLifecycleState, InvitationStatusFilter } from '$lib/domain/inventory';

export type SettingsLevel = 'overview' | 'account' | 'tenant' | 'inventory';
export type SettingsCollection = 'access' | 'activity' | 'fields' | 'asset-types' | 'tags' | null;
export type SettingsResourceAction = 'new' | 'edit' | 'archive' | 'restore' | 'delete' | null;

export interface SettingsRouteState {
  settingsLevel: SettingsLevel;
  settingsCollection: SettingsCollection;
  settingsLifecycle: CustomDefinitionLifecycleState;
  settingsResourceId: string | null;
  settingsResourceAction: SettingsResourceAction;
  invitationStatus: InvitationStatusFilter;
  accessInvitationAction: 'expire' | 'cancel' | 'delete' | null;
  accessInvitationId: string | null;
  auditScope: AuditScope;
  tenantId: string | null;
  inventoryId: string | null;
}

export function formatSettingsRouteHref(state: SettingsRouteState): string {
  let path = '/settings';
  if (state.settingsLevel === 'account') return `${path}/account/general`;
  if ((state.settingsLevel === 'tenant' || state.settingsLevel === 'inventory') && state.tenantId) {
    path += `/tenants/${encodeURIComponent(state.tenantId)}`;
    if (state.settingsLevel === 'inventory' && state.inventoryId) path += `/inventories/${encodeURIComponent(state.inventoryId)}`;
    if (state.settingsCollection) {
      path += `/${state.settingsCollection}`;
      if (state.settingsCollection === 'access' && state.accessInvitationAction && state.accessInvitationId) {
        path += `/invitations/${encodeURIComponent(state.accessInvitationId)}/${state.accessInvitationAction}`;
      } else if (state.settingsResourceAction === 'new') path += '/new';
      else if (state.settingsResourceId) {
        path += `/${encodeURIComponent(state.settingsResourceId)}`;
        if (state.settingsResourceAction) path += `/${state.settingsResourceAction}`;
      }
    }
  }
  const search = new URLSearchParams();
  if (state.settingsLifecycle === 'archived' && state.settingsCollection !== 'tags') search.set('lifecycle', 'archived');
  if (state.settingsCollection === 'access' && state.invitationStatus !== 'all') search.set('invitationStatus', state.invitationStatus);
  if (state.settingsCollection === 'activity' && state.auditScope !== 'inventory') search.set('auditScope', state.auditScope);
  return search.size ? `${path}?${search}` : path;
}
