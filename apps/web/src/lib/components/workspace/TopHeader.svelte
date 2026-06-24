<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { Inventory, Tenant } from '$lib/domain/inventory';

  let {
    tenants,
    inventories,
    selectedTenantId,
    inventory,
    query = $bindable(''),
    canCreateAsset,
    onSelectTenant,
    onSelectInventory,
    onSearch,
    onOpenAdd
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    inventory: Inventory | null;
    query: string;
    canCreateAsset: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onSearch: () => void;
    onOpenAdd: () => void;
  } = $props();

  let contextOpen = $state(false);
  let showingTenants = $state(false);
  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId));
</script>

<header class="workspace-header">
  <div class="mobile-context">
    <Button.Root variant="ghost" class="mobile-context-trigger" onclick={() => { contextOpen = !contextOpen; }}>
      <span>
        <strong>{inventory?.name ?? 'No inventory'}</strong>
        <small>{selectedTenant?.name ?? 'No tenant'}</small>
      </span>
    </Button.Root>
    {#if contextOpen}
      <div class="mobile-context-menu" aria-label="Inventory context">
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
                  contextOpen = false;
                  onSelectTenant(tenant.id);
                }}
              >
                {tenant.name}
              </Button.Root>
            {/each}
          </div>
        {:else if inventories.length > 0}
          <div class="inventory-menu">
            {#each inventories as candidate}
              <Button.Root
                variant={candidate.id === inventory?.id ? 'secondary' : 'ghost'}
                class="nav-button"
                onclick={() => {
                  contextOpen = false;
                  onSelectInventory(candidate.tenantId, candidate.id);
                }}
              >
                {candidate.name}
              </Button.Root>
            {/each}
          </div>
        {:else}
          <p class="muted small-copy">No inventories in this tenant.</p>
        {/if}
      </div>
    {/if}
  </div>
  <form class="global-search" onsubmit={(event) => { event.preventDefault(); onSearch(); }}>
    <Search aria-hidden="true" />
    <Input bind:value={query} placeholder="Search this inventory" aria-label="Search this inventory" />
    <Button.Root type="submit" variant="ghost" size="icon-sm" aria-label="Run search"><Search /></Button.Root>
  </form>
  <Button.Root class="header-add" disabled={!canCreateAsset || !inventory} onclick={onOpenAdd}>
    <Plus /> Add
  </Button.Root>
</header>
