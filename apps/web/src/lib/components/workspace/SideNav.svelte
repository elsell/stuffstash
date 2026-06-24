<script lang="ts">
  import Home from '@lucide/svelte/icons/house';
  import Settings from '@lucide/svelte/icons/settings';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import LogOut from '@lucide/svelte/icons/log-out';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';

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
  let selectedInventory = $derived(inventories.find((inventory) => inventory.id === selectedInventoryId));
  let showingTenants = $state(false);
</script>

<aside class="side-nav" aria-label="Workspace navigation">
  <div class="brand-lockup compact-lockup">
    <div class="brand-mark" aria-hidden="true"><span></span></div>
    <div>
      <strong>Stuff Stash</strong>
      <p>{selectedTenant?.name ?? 'Home'}</p>
    </div>
  </div>

  <section class="context-switcher" aria-label="Inventory context">
    <p class="eyebrow">Inventory</p>
    <strong>{selectedInventory?.name ?? 'No inventory'}</strong>
    <div class="tenant-row">
      <span>{selectedTenant?.name ?? 'No tenant'}</span>
      {#if tenants.length > 1}
        <Button.Root variant="ghost" size="sm" onclick={() => { showingTenants = !showingTenants; }}>
          {showingTenants ? 'Inventories' : 'Switch tenant'}
        </Button.Root>
      {/if}
    </div>
    {#if showingTenants}
      <div class="inventory-menu" aria-label="Tenants">
        {#each tenants as tenant}
          <Button.Root
            variant={tenant.id === selectedTenantId ? 'secondary' : 'ghost'}
            class="nav-button"
            onclick={() => {
              showingTenants = false;
              onSelectTenant(tenant.id);
            }}
          >
            {tenant.name}
          </Button.Root>
        {/each}
      </div>
    {:else if inventories.length > 0}
      <div class="inventory-menu">
        {#each inventories as inventory}
          <Button.Root
            variant={inventory.id === selectedInventoryId ? 'secondary' : 'ghost'}
            class="nav-button"
            onclick={() => onSelectInventory(inventory.tenantId, inventory.id)}
          >
            {inventory.name}
          </Button.Root>
        {/each}
      </div>
    {:else}
      <p class="muted small-copy">No inventories in this tenant.</p>
    {/if}
  </section>

  <nav class="nav-list">
    <Button.Root variant={mode === 'home' ? 'secondary' : 'ghost'} class="nav-button" onclick={() => onModeChange('home')}>
      <Home /> Home
    </Button.Root>
    <Button.Root variant={mode === 'location' ? 'secondary' : 'ghost'} class="nav-button" onclick={() => onModeChange('home')}>
      <MapPin /> Locations
    </Button.Root>
    <Button.Root variant={mode === 'settings' ? 'secondary' : 'ghost'} class="nav-button" onclick={() => onModeChange('settings')}>
      <Settings /> Settings
    </Button.Root>
  </nav>

  <div class="side-nav-footer">
    <p>{userLabel}</p>
    <Button.Root variant="ghost" size="sm" onclick={onSignOut}><LogOut /> Sign out</Button.Root>
  </div>
</aside>
