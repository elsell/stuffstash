import type { AssetKind, Inventory, WorkspaceMode } from '$lib/domain/inventory';
import { assetKindLabel, assetKinds, canViewImportJobs } from '$lib/domain/inventory';
import { workspaceRouteHref, type SettingsSection, type WorkspaceRouteState } from './workspaceRoute';

export type ShellWorkspaceMode = Extract<WorkspaceMode, 'home' | 'locations' | 'search' | 'import' | 'settings'>;
export type ShellNavigationIcon = 'home' | 'locations' | 'search' | 'import' | 'settings';

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

type ShellNavigationDefinition = Omit<ShellNavigationDestination, 'href' | 'current'>;

const desktopPrimaryDestinations: ShellNavigationDefinition[] = [
  { mode: 'home', label: 'Home', description: 'Recent assets and places', icon: 'home' },
  { mode: 'locations', label: 'Locations', description: 'Browse rooms, shelves, and places', icon: 'locations' }
];

const desktopUtilityDestinations: ShellNavigationDefinition[] = [
  { mode: 'import', label: 'Import', description: 'Bring in legacy data', icon: 'import' },
  { mode: 'settings', label: 'Settings', description: 'Access, fields, and audit', icon: 'settings' }
];

const mobileDestinations: ShellNavigationDefinition[] = [
  { mode: 'home', label: 'Home', description: 'Inventory home', icon: 'home' },
  { mode: 'search', label: 'Search', description: 'Find assets', icon: 'search' },
  { mode: 'locations', label: 'Places', description: 'Browse places', icon: 'locations' },
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
  return currentMode === destinationMode || (destinationMode === 'locations' && currentMode === 'location');
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
