import type { AuditScope, Inventory, InvitationStatusFilter, Tenant } from '$lib/domain/inventory';
import { workspaceRouteHref, type SettingsSection } from './workspaceRoute';

export interface SettingsNavigationOption<TValue extends string = string> {
  value: TValue;
  label: string;
  href?: string;
  disabled?: boolean;
}

export type SettingsSectionIcon = 'activity' | 'boxes' | 'sliders' | 'users';

export interface SettingsSectionNavigationOption extends SettingsNavigationOption<SettingsSection> {
  description: string;
  icon: SettingsSectionIcon;
  current: boolean;
}

export interface SettingsShellPresentation {
  title: string;
  contextLabel: string;
  liveAnnouncement: string;
  overviewContextLabel: string;
  emptyState: SettingsEmptyStatePresentation | null;
}

export interface SettingsEmptyStatePresentation {
  title: string;
  message: string;
}

export interface SettingsDetailRowPresentation {
  label: string;
  value: string;
}

export interface SettingsOverviewPresentation {
  title: string;
  contextLabel: string;
  rows: SettingsDetailRowPresentation[];
}

export interface SettingsAdministrationPresentation {
  title: string;
  description: string;
}

const invitationStatusFilters: InvitationStatusFilter[] = ['all', 'pending', 'accepted', 'revoked', 'cancelled', 'expired'];

const settingsSections: Array<Omit<SettingsSectionNavigationOption, 'href' | 'current'>> = [
  { value: 'overview', label: 'Overview', description: 'Inventory context and access summary', icon: 'boxes' },
  { value: 'access', label: 'Access', description: 'Sharing, grants, and invitations', icon: 'users' },
  { value: 'fields', label: 'Fields', description: 'Custom asset types and fields', icon: 'sliders' },
  { value: 'activity', label: 'Activity', description: 'Audit history for this workspace', icon: 'activity' }
];

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
      settingsLevel: 'inventory',
      tenantId,
      inventoryId,
      settingsCollection: section === 'access' || section === 'activity' || section === 'fields' ? section : null,
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
  return workspaceRouteHref({ mode: 'settings', settingsLevel: 'inventory', tenantId, inventoryId, settingsCollection: 'access', settingsSection: 'access', invitationStatus: status }, tenantId, inventoryId);
}

export function settingsSectionOptions(input: {
  tenantId: string | null;
  inventoryId: string | null;
  section: SettingsSection;
  invitationStatus: InvitationStatusFilter;
  auditScope: AuditScope;
}): SettingsSectionNavigationOption[] {
  return settingsSections.map((option) => ({
    ...option,
    href: settingsSectionHref(input.tenantId, input.inventoryId, option.value, input.invitationStatus, input.auditScope),
    current: option.value === input.section
  }));
}

export function settingsInvitationStatusOptions(input: {
  tenantId: string | null;
  inventoryId: string | null;
}): SettingsNavigationOption<InvitationStatusFilter>[] {
  const routeBacked = !!input.inventoryId;
  return invitationStatusFilters.map((status) => ({
    value: status,
    label: invitationStatusLabel(status),
    href: routeBacked ? settingsInvitationStatusHref(input.tenantId, input.inventoryId, status) : undefined
  }));
}

export function settingsAuditScopeHref(tenantId: string | null, inventoryId: string | null, scope: AuditScope): string {
  return workspaceRouteHref({ mode: 'settings', settingsLevel: 'inventory', tenantId, inventoryId, settingsCollection: 'activity', settingsSection: 'activity', auditScope: scope }, tenantId, inventoryId);
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

export function settingsShellPresentation(input: {
  tenant: Pick<Tenant, 'name'> | null;
  inventory: Pick<Inventory, 'name'> | null;
  activeSection: Pick<SettingsSectionNavigationOption, 'label' | 'description'>;
}): SettingsShellPresentation {
  if (!input.inventory) {
    return {
      title: 'Settings',
      contextLabel: 'No inventory selected',
      liveAnnouncement: `${input.activeSection.label}: ${input.activeSection.description}`,
      overviewContextLabel: 'Not available',
      emptyState: {
        title: 'No inventory selected',
        message: 'Select or create an inventory before managing settings.'
      }
    };
  }
  return {
    title: 'Settings',
    contextLabel: `${input.inventory.name} / ${input.activeSection.label}`,
    liveAnnouncement: `${input.activeSection.label}: ${input.activeSection.description}`,
    overviewContextLabel: `${input.tenant?.name ?? 'No tenant'} / ${input.inventory.name}`,
    emptyState: null
  };
}

export function settingsOverviewPresentation(input: {
  tenantName: string | null;
  inventoryCount: number;
  accessRelationship: string;
  canEditAssets: boolean;
  contextLabel: string;
}): SettingsOverviewPresentation {
  return {
    title: 'Overview',
    contextLabel: input.contextLabel,
    rows: [
      { label: 'Tenant', value: input.tenantName ?? 'Not available' },
      { label: 'Inventories', value: String(input.inventoryCount) },
      { label: 'Access', value: input.accessRelationship },
      { label: 'Asset edits', value: input.canEditAssets ? 'Allowed' : 'View only' }
    ]
  };
}

export function settingsAdministrationPresentation(input: { canConfigureTenant: boolean }): SettingsAdministrationPresentation {
  return {
    title: 'Administration',
    description: input.canConfigureTenant
      ? 'There are no administration actions available in the web app yet.'
      : 'This account does not have access to tenant administration.'
  };
}

function invitationStatusLabel(status: InvitationStatusFilter): string {
  return status[0]?.toUpperCase() + status.slice(1);
}
