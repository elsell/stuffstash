<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { Inventory, Tenant } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    inventory,
    query = $bindable(''),
    canCreateAsset,
    onSelectTenant,
    onSelectInventory,
    onOpenSettings,
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
    onOpenSettings: () => void;
    onSearch: () => void;
    onOpenAdd: () => void;
  } = $props();

  let selectedInventoryId = $derived(inventory?.id ?? '');
</script>

<header class="workspace-header">
  <div class="mobile-context">
    <WorkspaceContextSwitcher
      mobile
      {tenants}
      {inventories}
      {selectedTenantId}
      {selectedInventoryId}
      {onSelectTenant}
      {onSelectInventory}
      {onOpenSettings}
    />
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
