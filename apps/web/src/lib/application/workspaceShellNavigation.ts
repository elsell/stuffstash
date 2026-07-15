import type { AssetKind, Inventory, Principal, WorkspaceMode } from '$lib/domain/inventory';
import { assetKindLabel, assetKinds, canViewImportJobs } from '$lib/domain/inventory';
import { workspaceRouteHref, type SettingsSection, type WorkspaceRouteState } from './workspaceRoute';

export type ShellWorkspaceMode = Extract<WorkspaceMode, 'home' | 'browse' | 'import' | 'settings'>;
export type ShellNavigationIcon = 'home' | 'browse' | 'import' | 'settings';

export interface ShellNavigationDestination {
  mode: ShellWorkspaceMode;
  label: string;
  description: string;
  icon: ShellNavigationIcon;
  href: string;
  current: boolean;
}

export interface ShellNavigationGroup {
  id: string;
  label: string;
  destinations: ShellNavigationDestination[];
}

export interface ShellNavigationInput {
  mode: WorkspaceMode;
  tenantId: string | null;
  inventoryId: string | null;
  inventory?: Inventory | null;
  settingsSection?: SettingsSection;
}

export interface ShellAddOption {
  kind: AssetKind;
  label: string;
  href: string;
}

export function accountDisplayLabel(principal: Principal): string {
  const email = principal.email?.trim();
  return email || 'Signed-in account';
}

type ShellNavigationDefinition = Omit<ShellNavigationDestination, 'href' | 'current'>;

const desktopPrimaryDestinations: ShellNavigationDefinition[] = [
  { mode: 'home', label: 'Home', description: 'Recent assets and places', icon: 'home' },
  { mode: 'browse', label: 'Browse', description: 'Find and explore your inventory', icon: 'browse' }
];

const desktopUtilityDestinations: ShellNavigationDefinition[] = [
  { mode: 'import', label: 'Import', description: 'Bring in outside data', icon: 'import' },
  { mode: 'settings', label: 'Settings', description: 'Access, fields, and audit', icon: 'settings' }
];

const mobileDestinations: ShellNavigationDefinition[] = [
  { mode: 'home', label: 'Home', description: 'Inventory home', icon: 'home' },
  { mode: 'browse', label: 'Browse', description: 'Find and explore', icon: 'browse' },
  { mode: 'settings', label: 'Settings', description: 'Inventory settings', icon: 'settings' }
];

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

export function shellAddOptions(tenantId: string | null, inventoryId: string | null): ShellAddOption[] {
  return assetKinds.map((kind) => ({
    kind,
    label: assetKindLabel(kind),
    href: shellAddHref(kind, tenantId, inventoryId)
  }));
}

export function desktopShellNavigationGroups(input: ShellNavigationInput): ShellNavigationGroup[] {
  return [
    {
      id: 'primary',
      label: 'Inventory',
      destinations: shellDestinations(desktopPrimaryDestinations, input)
    },
    {
      id: 'utility',
      label: 'Tools',
      destinations: shellDestinations(desktopUtilityDestinations, input)
    }
  ];
}

export function mobileShellNavigationItems(input: ShellNavigationInput): ShellNavigationDestination[] {
  return shellDestinations(mobileDestinations, input);
}

export function shellModeIsCurrent(currentMode: WorkspaceMode, destinationMode: ShellWorkspaceMode): boolean {
  return currentMode === destinationMode || (destinationMode === 'browse' && (currentMode === 'location' || currentMode === 'locations' || currentMode === 'search'));
}

export function contextInventoryHref(inventory: Inventory): string {
  return workspaceRouteHref({ mode: 'home', tenantId: inventory.tenantId, inventoryId: inventory.id }, inventory.tenantId, inventory.id);
}

function shellDestinations(definitions: ShellNavigationDefinition[], input: ShellNavigationInput): ShellNavigationDestination[] {
  return definitions
    .filter((destination) => destination.mode !== 'import' || canViewImportJobs(input.inventory))
    .map((destination) => ({
      ...destination,
      href: shellModeHref(destination.mode, input.tenantId, input.inventoryId, input.settingsSection ?? 'overview'),
      current: shellModeIsCurrent(input.mode, destination.mode)
    }));
}
