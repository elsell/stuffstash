import type { AssetKind, ImportSourceType, Inventory, WorkspaceMode } from '$lib/domain/inventory';
import { workspaceRouteHref, type SettingsSection, type WorkspaceRouteState } from './workspaceRoute';

export type ShellWorkspaceMode = Extract<WorkspaceMode, 'home' | 'locations' | 'search' | 'import' | 'settings'>;

export function shellModeHref(
  mode: ShellWorkspaceMode,
  tenantId: string | null,
  inventoryId: string | null,
  settingsSection: SettingsSection = 'overview'
): string {
  const route: Partial<WorkspaceRouteState> = { mode };
  if (mode === 'settings') {
    route.settingsSection = settingsSection;
  }
  return workspaceRouteHref(route, tenantId, inventoryId);
}

export function shellAddHref(kind: AssetKind, tenantId: string | null, inventoryId: string | null): string {
  return workspaceRouteHref({ action: 'add', addKind: kind }, tenantId, inventoryId);
}

export function contextInventoryHref(inventory: Inventory): string {
  return workspaceRouteHref({ mode: 'home', tenantId: inventory.tenantId, inventoryId: inventory.id }, inventory.tenantId, inventory.id);
}

export function importSourceHref(tenantId: string, inventoryId: string | null, sourceType: ImportSourceType): string {
  return workspaceRouteHref({ mode: 'import', tenantId, inventoryId, importSourceType: sourceType }, tenantId, inventoryId);
}
