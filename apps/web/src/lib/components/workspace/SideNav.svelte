<script lang="ts">
  import Home from '@lucide/svelte/icons/house';
  import Settings from '@lucide/svelte/icons/settings';
  import Upload from '@lucide/svelte/icons/upload';
  import LogOut from '@lucide/svelte/icons/log-out';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mode,
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
    userLabel: string;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onModeChange: (mode: WorkspaceMode) => void;
    onSignOut: () => void;
  } = $props();

  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId));

  type NavDestination = {
    mode: WorkspaceMode;
    label: string;
    description: string;
    icon: Component;
  };

  const primaryDestinations: NavDestination[] = [
    { mode: 'home', label: 'Home', description: 'Browse assets and locations', icon: Home }
  ];

  const utilityDestinations: NavDestination[] = [
    { mode: 'import', label: 'Import', description: 'Bring in legacy data', icon: Upload },
    { mode: 'settings', label: 'Settings', description: 'Access, fields, and audit', icon: Settings }
  ];

  function openDestination(destination: NavDestination): void {
    onModeChange(destination.mode);
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
    onOpenSettings={() => onModeChange('settings')}
  />

  <nav class="side-nav-groups" aria-label="Inventory destinations">
    <div class="nav-section" aria-labelledby="primary-nav-label">
      <p id="primary-nav-label" class="nav-eyebrow">Inventory</p>
      <div class="nav-list">
        {#each primaryDestinations as destination}
          {@const Icon = destination.icon}
          <Button.Root
            variant={mode === destination.mode ? 'secondary' : 'ghost'}
            class="nav-button"
            aria-current={mode === destination.mode ? 'page' : undefined}
            onclick={() => openDestination(destination)}
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
            variant={mode === destination.mode ? 'secondary' : 'ghost'}
            class="nav-button"
            aria-current={mode === destination.mode ? 'page' : undefined}
            onclick={() => openDestination(destination)}
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
