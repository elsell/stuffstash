<script lang="ts">
  import Home from '@lucide/svelte/icons/house';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import Settings from '@lucide/svelte/icons/settings';
  import Upload from '@lucide/svelte/icons/upload';
  import LogOut from '@lucide/svelte/icons/log-out';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { shellModeHref, type ShellWorkspaceMode } from '$lib/application/workspaceShellNavigation';
  import type { SettingsSection } from '$lib/application/workspaceRoute';
  import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mode,
    settingsSection,
    userLabel,
    onSelectTenant,
    onSelectInventory,
    onModeChange,
    onSignOut
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    mode: WorkspaceMode;
    settingsSection: SettingsSection;
    userLabel: string;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onModeChange: (mode: WorkspaceMode) => void;
    onSignOut: () => void;
  } = $props();

  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId));

  type NavDestination = {
    mode: ShellWorkspaceMode;
    label: string;
    description: string;
    icon: Component;
  };

  const primaryDestinations: NavDestination[] = [
    { mode: 'home', label: 'Home', description: 'Recent assets and places', icon: Home },
    { mode: 'locations', label: 'Locations', description: 'Browse rooms, shelves, and places', icon: MapPin }
  ];

  const utilityDestinations: NavDestination[] = [
    { mode: 'import', label: 'Import', description: 'Bring in legacy data', icon: Upload },
    { mode: 'settings', label: 'Settings', description: 'Access, fields, and audit', icon: Settings }
  ];

  function openDestination(event: MouseEvent, destination: NavDestination): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onModeChange(destination.mode);
  }

  function destinationIsCurrent(destination: NavDestination): boolean {
    return mode === destination.mode || (destination.mode === 'locations' && mode === 'location');
  }

  function destinationHref(destination: NavDestination): string {
    return shellModeHref(destination.mode, selectedTenantId || null, selectedInventoryId || null, settingsSection);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }
</script>

<aside class="side-nav" aria-label="Workspace navigation">
  <div class="brand-lockup compact-lockup">
    <div class="brand-mark" aria-hidden="true"><span></span></div>
    <div>
      <strong>Stuff Stash</strong>
      <p>{selectedTenant?.name ?? 'Home'}</p>
    </div>
  </div>

  <WorkspaceContextSwitcher
    {tenants}
    {inventories}
    {selectedTenantId}
    {selectedInventoryId}
    {onSelectTenant}
    {onSelectInventory}
  />

  <nav class="side-nav-groups" aria-label="Inventory destinations">
    <div class="nav-section" aria-labelledby="primary-nav-label">
      <p id="primary-nav-label" class="nav-eyebrow">Inventory</p>
      <div class="nav-list">
        {#each primaryDestinations as destination}
          {@const Icon = destination.icon}
          <Button.Root
            href={destinationHref(destination)}
            variant={destinationIsCurrent(destination) ? 'secondary' : 'ghost'}
            class="nav-button"
            aria-current={destinationIsCurrent(destination) ? 'page' : undefined}
            onclick={(event) => openDestination(event, destination)}
          >
            <Icon aria-hidden="true" />
            <span>
              <strong>{destination.label}</strong>
              <small>{destination.description}</small>
            </span>
          </Button.Root>
        {/each}
      </div>
    </div>

    <div class="nav-section" aria-labelledby="utility-nav-label">
      <p id="utility-nav-label" class="nav-eyebrow">Tools</p>
      <div class="nav-list">
        {#each utilityDestinations as destination}
          {@const Icon = destination.icon}
          <Button.Root
            href={destinationHref(destination)}
            variant={destinationIsCurrent(destination) ? 'secondary' : 'ghost'}
            class="nav-button"
            aria-current={destinationIsCurrent(destination) ? 'page' : undefined}
            onclick={(event) => openDestination(event, destination)}
          >
            <Icon aria-hidden="true" />
            <span>
              <strong>{destination.label}</strong>
              <small>{destination.description}</small>
            </span>
          </Button.Root>
        {/each}
      </div>
    </div>
  </nav>

  <div class="side-nav-footer">
    <p>{userLabel}</p>
    <Button.Root variant="ghost" size="sm" onclick={onSignOut}><LogOut /> Sign out</Button.Root>
  </div>
</aside>
