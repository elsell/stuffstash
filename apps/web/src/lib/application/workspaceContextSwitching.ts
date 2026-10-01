import { t } from '$lib/presentation/localization';
import type { Inventory, Tenant } from '$lib/domain/inventory';
import { contextInventoryHref } from './workspaceShellNavigation';

export interface TenantContextOption {
  id: string;
  name: string;
  inventoryCountLabel: string;
  selected: boolean;
}

export interface InventoryContextOption {
  id: string;
  tenantId: string;
  name: string;
  tenantName: string;
  relationshipLabel: string;
  href: string;
  selected: boolean;
}

export interface ContextSwitcherPresentation {
  triggerInventoryLabel: string;
  triggerTenantLabel: string;
  activeTenantLabel: string;
  emptyInventoryMessage: string;
}

export function tenantContextOptions(input: {
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
}): TenantContextOption[] {
  return input.tenants.map((tenant) => ({
    id: tenant.id,
    name: tenant.name,
    inventoryCountLabel: inventoryCountLabel(input.inventories, tenant.id),
    selected: tenant.id === input.selectedTenantId
  }));
}

export function inventoryContextOptions(input: {
  tenants: Tenant[];
  inventories: Inventory[];
  selectedTenantId: string;
  selectedInventoryId: string;
}): InventoryContextOption[] {
  const selectedTenant = input.tenants.find((tenant) => tenant.id === input.selectedTenantId);
  return input.inventories
    .filter((inventory) => inventory.tenantId === input.selectedTenantId)
    .map((inventory) => ({
      id: inventory.id,
      tenantId: inventory.tenantId,
      name: inventory.name,
      tenantName: selectedTenant?.name ?? t('web.workspaceContextSwitching.inventory'),
      relationshipLabel: relationshipLabel(inventory.access.relationship),
      href: contextInventoryHref(inventory),
      selected: inventory.id === input.selectedInventoryId
    }));
}

export function inventoryCountLabel(inventories: Inventory[], tenantId: string): string {
  const count = inventories.filter((inventory) => inventory.tenantId === tenantId).length;
  return `${count} ${count === 1 ? 'inventory' : 'inventories'}`;
}

export function relationshipLabel(relationship: string | undefined): string {
  if (!relationship) {
    return t('web.workspaceContextSwitching.member');
  }
  return relationship
    .split(/[\s_-]+/)
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase() + part.slice(1))
    .join(' ');
}

export function contextSwitcherPresentation(input: {
  selectedTenant: Pick<Tenant, 'name'> | null;
  selectedInventory: Pick<Inventory, 'name'> | null;
}): ContextSwitcherPresentation {
  const tenantLabel = input.selectedTenant?.name ?? t('web.workspaceContextSwitching.noTenant');
  return {
    triggerInventoryLabel: input.selectedInventory?.name ?? t('web.workspaceContextSwitching.noInventory'),
    triggerTenantLabel: tenantLabel,
    activeTenantLabel: tenantLabel,
    emptyInventoryMessage: t('web.workspaceContextSwitching.noInventoriesInThisTenant')
  };
}
