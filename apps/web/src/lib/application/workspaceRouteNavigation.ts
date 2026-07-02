import type { Inventory, WorkspaceData } from '$lib/domain/inventory';
import { canEditAsset, type Asset } from '$lib/domain/inventory';
import {
  defaultWorkspaceRoute,
  parseWorkspaceRoute,
  workspaceRouteHref,
  type AssetRouteAction,
  type WorkspaceRouteState
} from './workspaceRoute';

export type BrowserRouteTarget = Pick<Window, 'history' | 'location'>;

function browserRouteTarget(): BrowserRouteTarget | null {
  return typeof window === 'undefined' ? null : window;
}

export function currentWorkspaceRoute(target: BrowserRouteTarget | null = browserRouteTarget()): WorkspaceRouteState {
  return target ? parseWorkspaceRoute(new URL(target.location.href)) : { ...defaultWorkspaceRoute };
}

export function pushWorkspaceRoute(
  route: Partial<WorkspaceRouteState>,
  selectedTenantId: string | null,
  selectedInventoryId: string | null,
  target: BrowserRouteTarget | null = browserRouteTarget()
): WorkspaceRouteState {
  if (!target) {
    return { ...defaultWorkspaceRoute, ...route };
  }
  target.history.pushState({}, '', workspaceRouteHref(route, selectedTenantId, selectedInventoryId));
  return currentWorkspaceRoute(target);
}

export function replaceWorkspaceRoute(
  route: Partial<WorkspaceRouteState>,
  selectedTenantId: string | null,
  selectedInventoryId: string | null,
  target: BrowserRouteTarget | null = browserRouteTarget()
): void {
  target?.history.replaceState({}, '', workspaceRouteHref(route, selectedTenantId, selectedInventoryId));
}

export function shouldCanonicalizeWorkspaceAlias(route: WorkspaceRouteState): boolean {
  return !!route.inventoryId && !route.tenantId;
}

export function replaceCanonicalWorkspaceAlias(
  route: WorkspaceRouteState,
  selectedTenantId: string | null,
  selectedInventoryId: string | null,
  target: BrowserRouteTarget | null = browserRouteTarget()
): void {
  if (!shouldCanonicalizeWorkspaceAlias(route) || !selectedTenantId || !selectedInventoryId) {
    return;
  }
  replaceWorkspaceRoute(
    {
      ...route,
      tenantId: selectedTenantId,
      inventoryId: selectedInventoryId
    },
    selectedTenantId,
    selectedInventoryId,
    target
  );
}

export function findRouteTenant(data: WorkspaceData, route: WorkspaceRouteState): string | null {
  if (!route.tenantId || route.tenantId === data.context.selectedTenantId) {
    return null;
  }
  return data.context.tenants.some((tenant) => tenant.id === route.tenantId) ? route.tenantId : null;
}

export function findRouteInventory(data: WorkspaceData, route: WorkspaceRouteState): Inventory | null {
  if (!route.inventoryId || route.inventoryId === data.context.selectedInventoryId) {
    return null;
  }
  return (
    data.context.inventories.find(
      (inventory) =>
        inventory.id === route.inventoryId &&
        (route.tenantId ? inventory.tenantId === route.tenantId : inventory.tenantId === data.context.selectedTenantId)
    ) ?? null
  );
}

export function assetRouteActionIsAvailable(
  action: AssetRouteAction,
  inventory: Inventory | null,
  asset: Asset | null
): boolean {
  if (!action) {
    return true;
  }
  if (!canEditAsset(inventory)) {
    return false;
  }
  return action === 'delete' || asset?.lifecycleState === 'active';
}
