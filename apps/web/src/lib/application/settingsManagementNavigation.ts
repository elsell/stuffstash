import type { AuditScope, Inventory, InvitationStatusFilter, Tenant } from '$lib/domain/inventory';
import { workspaceRouteHref, type AccessInvitationRouteAction, type SettingsCollection, type SettingsResourceAction } from './workspaceRoute';

export type SettingsDestinationIcon = 'account' | 'tenant' | 'inventory' | 'access' | 'activity' | 'fields' | 'asset-types' | 'tags';

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
    label: 'Account and app', eyebrow: 'Personal', description: 'Account, connection, and app information',
    href: '/settings/account/general', icon: 'account'
  }];
  if (input.tenant) rows.push({
    label: input.tenant.name, eyebrow: 'Tenant settings', description: 'Fields and asset types shared with its inventories',
    href: settingsResourceHref({ level: 'tenant', tenantId: input.tenant.id }), icon: 'tenant'
  });
  if (input.inventory) rows.push({
    label: input.inventory.name, eyebrow: 'Inventory settings', description: `Belongs to ${input.tenant?.name ?? 'the selected tenant'}`,
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
    { label: 'Custom fields', eyebrow: 'Shared schema', description: 'Fields available to every inventory', icon: 'fields', href: settingsResourceHref({ level: 'tenant', tenantId: tenant.id, collection: 'fields' }) },
    { label: 'Asset types', eyebrow: 'Shared schema', description: 'Types available to every inventory', icon: 'asset-types', href: settingsResourceHref({ level: 'tenant', tenantId: tenant.id, collection: 'asset-types' }) }
  ];
}

export function inventorySettingsDestinations(inventory: Pick<Inventory, 'id' | 'tenantId'>): SettingsDestination[] {
  const base = { level: 'inventory' as const, tenantId: inventory.tenantId, inventoryId: inventory.id };
  return [
    { label: 'Sharing', eyebrow: 'People', description: 'Access and invitations', icon: 'access', href: settingsResourceHref({ ...base, collection: 'access' }) },
    { label: 'Tags', eyebrow: 'Organization', description: 'Reusable labels for this inventory', icon: 'tags', href: settingsResourceHref({ ...base, collection: 'tags' }) },
    { label: 'Custom fields', eyebrow: 'Schema', description: 'Inherited and inventory-only fields', icon: 'fields', href: settingsResourceHref({ ...base, collection: 'fields' }) },
    { label: 'Asset types', eyebrow: 'Schema', description: 'Inherited and inventory-only types', icon: 'asset-types', href: settingsResourceHref({ ...base, collection: 'asset-types' }) },
    { label: 'Activity', eyebrow: 'History', description: 'Audit history for this inventory', icon: 'activity', href: settingsResourceHref({ ...base, collection: 'activity' }) }
  ];
}
