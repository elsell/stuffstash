import type { AuditScope, InvitationStatusFilter } from '$lib/domain/inventory';
import { workspaceRouteHref, type SettingsSection } from './workspaceRoute';

export interface SettingsNavigationOption<TValue extends string = string> {
  value: TValue;
  label: string;
  href?: string;
  disabled?: boolean;
}

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

export function settingsAuditScopeOptions(input: {
  tenantId: string | null;
  inventoryId: string | null;
  hasTenant: boolean;
  hasInventory: boolean;
}): SettingsNavigationOption<AuditScope>[] {
  const routeBacked = !!input.inventoryId;
  return [
    {
      value: 'inventory',
      label: 'Inventory',
      href: routeBacked ? settingsAuditScopeHref(input.tenantId, input.inventoryId, 'inventory') : undefined,
      disabled: !input.hasInventory
    },
    {
      value: 'tenant',
      label: 'Tenant',
      href: routeBacked ? settingsAuditScopeHref(input.tenantId, input.inventoryId, 'tenant') : undefined,
      disabled: !input.hasTenant
    }
  ];
}
