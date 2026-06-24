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
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onModeChange: (mode: WorkspaceMode) => void;
    onSignOut: () => void;
  } = $props();

  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId));
  let selectedInventory = $derived(inventories.find((inventory) => inventory.id === selectedInventoryId));
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
    {#if inventories.length > 0}
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
