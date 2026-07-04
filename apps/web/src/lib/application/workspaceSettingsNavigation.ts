import type { AuditScope, InvitationStatusFilter } from '$lib/domain/inventory';
import { workspaceRouteHref, type SettingsSection } from './workspaceRoute';

export function settingsSectionHref(
  tenantId: string | null,
  inventoryId: string | null,
  section: SettingsSection,
  invitationStatus: InvitationStatusFilter,
  auditScope: AuditScope
): string {
  return workspaceRouteHref(
    {
      mode: 'settings',
      settingsSection: section,
      invitationStatus: section === 'access' ? invitationStatus : 'all',
      auditScope: section === 'activity' ? auditScope : 'inventory'
    },
    tenantId,
    inventoryId
  );
}

export function settingsInvitationStatusHref(
  tenantId: string | null,
  inventoryId: string | null,
  status: InvitationStatusFilter
): string {
  return workspaceRouteHref({ mode: 'settings', settingsSection: 'access', invitationStatus: status }, tenantId, inventoryId);
}

export function settingsAuditScopeHref(tenantId: string | null, inventoryId: string | null, scope: AuditScope): string {
  return workspaceRouteHref({ mode: 'settings', settingsSection: 'activity', auditScope: scope }, tenantId, inventoryId);
}
