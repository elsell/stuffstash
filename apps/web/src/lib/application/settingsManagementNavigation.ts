import { t } from '$lib/presentation/localization';
import type { AuditScope, Inventory, InvitationStatusFilter, Tenant } from '$lib/domain/inventory';
import { workspaceRouteHref, type AccessInvitationRouteAction, type SettingsCollection, type SettingsResourceAction } from './workspaceRoute';

export type SettingsDestinationIcon = 'account' | 'tenant' | 'inventory' | 'access' | 'activity' | 'fields' | 'asset-types' | 'tags' | 'conversations' | 'notifications' | 'printing';

export interface SettingsDestination {
  label: string;
  eyebrow: string;
  description: string;
  href: string;
  icon: SettingsDestinationIcon;
}

export function settingsOverviewDestinations(input: {
  tenant: Pick<Tenant, 'id' | 'name'> | null;
  inventory: Pick<Inventory, 'id' | 'tenantId' | 'name'> | null;
}): SettingsDestination[] {
  const rows: SettingsDestination[] = [{
    label: t('web.settingsManagementNavigation.accountAndApp'), eyebrow: t('web.settingsManagementNavigation.personal'), description: t('web.settingsManagementNavigation.accountConnectionAndAppInformation'),
    href: '/settings/account/general', icon: 'account'
  }];
  if (input.tenant) rows.push({
    label: input.tenant.name, eyebrow: t('web.settingsManagementNavigation.tenantSettings'), description: t('web.settingsManagementNavigation.fieldsAndAssetTypesSharedWithItsInventories'),
    href: settingsResourceHref({ level: 'tenant', tenantId: input.tenant.id }), icon: 'tenant'
  });
  if (input.inventory) rows.push({
    label: input.inventory.name, eyebrow: t('web.settingsManagementNavigation.inventorySettings'), description: input.tenant ? t('web.settingsManagementNavigation.belongsTo', { value: input.tenant.name }) : t('web.settingsManagementNavigation.belongsToSelectedTenant'),
    href: settingsResourceHref({ level: 'inventory', tenantId: input.inventory.tenantId, inventoryId: input.inventory.id }), icon: 'inventory'
  });
  return rows;
}

export function settingsResourceHref(input: {
  level: 'tenant' | 'inventory'; tenantId: string; inventoryId?: string; collection?: SettingsCollection;
  lifecycle?: 'active' | 'archived'; resourceId?: string; action?: SettingsResourceAction;
  invitationStatus?: InvitationStatusFilter; auditScope?: AuditScope;
  accessInvitationId?: string; accessInvitationAction?: AccessInvitationRouteAction;
}): string {
  return workspaceRouteHref({
    mode: 'settings', settingsLevel: input.level, tenantId: input.tenantId, inventoryId: input.inventoryId ?? null,
    settingsCollection: input.collection ?? null, settingsLifecycle: input.lifecycle ?? 'active',
    settingsResourceId: input.resourceId ?? null, settingsResourceAction: input.action ?? null,
    invitationStatus: input.invitationStatus ?? 'all', auditScope: input.auditScope ?? 'inventory',
    accessInvitationId: input.accessInvitationId ?? null, accessInvitationAction: input.accessInvitationAction ?? null
  }, null, null);
}

export function tenantSettingsDestinations(tenant: Pick<Tenant, 'id'>): SettingsDestination[] {
  return [
    { label: t('web.settingsManagementNavigation.conversations'), eyebrow: t('web.settingsManagementNavigation.voiceAndModels'), description: t('web.settingsManagementNavigation.tuneWorkflowsAndTestRealisticInventoryRequests'), icon: 'conversations', href: settingsResourceHref({ level: 'tenant', tenantId: tenant.id, collection: 'conversations' }) },
    { label: t('web.settingsManagementNavigation.customFields'), eyebrow: t('web.settingsManagementNavigation.sharedSchema'), description: t('web.settingsManagementNavigation.fieldsAvailableToEveryInventory'), icon: 'fields', href: settingsResourceHref({ level: 'tenant', tenantId: tenant.id, collection: 'fields' }) },
    { label: t('web.settingsManagementNavigation.assetTypes'), eyebrow: t('web.settingsManagementNavigation.sharedSchema'), description: t('web.settingsManagementNavigation.typesAvailableToEveryInventory'), icon: 'asset-types', href: settingsResourceHref({ level: 'tenant', tenantId: tenant.id, collection: 'asset-types' }) }
  ];
}

export function inventorySettingsDestinations(inventory: Pick<Inventory, 'id' | 'tenantId'>): SettingsDestination[] {
  const base = { level: 'inventory' as const, tenantId: inventory.tenantId, inventoryId: inventory.id };
  return [
    { label: t('web.Printing.title'), eyebrow: t('web.Printing.shared'), description: t('web.Printing.description'), icon: 'printing', href: settingsResourceHref({ ...base, collection: 'printing' }) },
    { label: t('web.settingsManagementNavigation.notifications'), eyebrow: t('web.settingsManagementNavigation.personal'), description: t('web.settingsManagementNavigation.expirationRemindersAndAssetTypeOverrides'), icon: 'notifications', href: settingsResourceHref({ ...base, collection: 'notifications' }) },
    { label: t('web.settingsManagementNavigation.sharing'), eyebrow: t('web.settingsManagementNavigation.people'), description: t('web.settingsManagementNavigation.accessAndInvitations'), icon: 'access', href: settingsResourceHref({ ...base, collection: 'access' }) },
    { label: t('web.settingsManagementNavigation.tags'), eyebrow: t('web.settingsManagementNavigation.organization'), description: t('web.settingsManagementNavigation.reusableLabelsForThisInventory'), icon: 'tags', href: settingsResourceHref({ ...base, collection: 'tags' }) },
    { label: t('web.settingsManagementNavigation.customFields'), eyebrow: t('web.settingsManagementNavigation.schema'), description: t('web.settingsManagementNavigation.inheritedAndInventoryOnlyFields'), icon: 'fields', href: settingsResourceHref({ ...base, collection: 'fields' }) },
    { label: t('web.settingsManagementNavigation.assetTypes'), eyebrow: t('web.settingsManagementNavigation.schema'), description: t('web.settingsManagementNavigation.inheritedAndInventoryOnlyTypes'), icon: 'asset-types', href: settingsResourceHref({ ...base, collection: 'asset-types' }) },
    { label: t('web.settingsManagementNavigation.activity'), eyebrow: t('web.settingsManagementNavigation.history'), description: t('web.settingsManagementNavigation.auditHistoryForThisInventory'), icon: 'activity', href: settingsResourceHref({ ...base, collection: 'activity' }) }
  ];
}
