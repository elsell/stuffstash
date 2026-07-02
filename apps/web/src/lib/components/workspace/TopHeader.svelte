<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { Asset, AssetKind, Inventory, Tenant } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    inventory,
    suggestions,
    query = $bindable(''),
    canCreateAsset,
    onSelectTenant,
    onSelectInventory,
    onOpenSettings,
    onSearch,
    onOpenAsset,
    onOpenAdd
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    inventory: Inventory | null;
    suggestions: Asset[];
    query: string;
    canCreateAsset: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onOpenSettings: () => void;
    onSearch: () => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: (kind: AssetKind) => void;
  } = $props();

  let selectedInventoryId = $derived(inventory?.id ?? '');
  let searchFocused = $state(false);
  let addMenuOpen = $state(false);
  let visibleSuggestions = $derived(searchFocused && query.trim().length > 0 ? suggestions.slice(0, 6) : []);
  const addKinds: AssetKind[] = ['item', 'container', 'location'];

  function chooseAddKind(kind: AssetKind): void {
    addMenuOpen = false;
    onOpenAdd(kind);
  }
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
  <div class="global-search-wrap">
    <form class="global-search" onsubmit={(event) => { event.preventDefault(); onSearch(); }}>
      <Search aria-hidden="true" />
      <Input
        bind:value={query}
        placeholder="Search this inventory"
        aria-label="Search this inventory"
        onfocus={() => { searchFocused = true; }}
        onblur={() => { window.setTimeout(() => { searchFocused = false; }, 120); }}
      />
      <Button.Root type="submit" variant="ghost" size="icon-sm" aria-label="Run search"><Search /></Button.Root>
    </form>
    {#if visibleSuggestions.length > 0}
      <div id="global-search-suggestions" class="search-suggestions" aria-label="Search suggestions">
        {#each visibleSuggestions as suggestion}
          <Button.Root
            variant="ghost"
            class="suggestion-row"
            aria-label={`Open ${suggestion.title}`}
            onclick={() => {
              query = suggestion.title;
              searchFocused = false;
              onOpenAsset(suggestion);
            }}
          >
            <span>
              <strong>{suggestion.title}</strong>
              <small>{assetKindLabel(suggestion.kind)}</small>
            </span>
          </Button.Root>
        {/each}
      </div>
    {/if}
  </div>
  <div class="header-add-wrap">
    <Button.Root
      class="header-add"
      disabled={!canCreateAsset || !inventory}
      aria-expanded={addMenuOpen}
      aria-controls="header-add-menu"
      onclick={() => { addMenuOpen = !addMenuOpen; }}
    >
      <Plus /> Add
    </Button.Root>
    {#if addMenuOpen}
      <div id="header-add-menu" class="add-menu" aria-label="Add asset kind">
        {#each addKinds as kind}
          <Button.Root variant="ghost" class="add-menu-item" onclick={() => chooseAddKind(kind)}>
            {assetKindLabel(kind)}
          </Button.Root>
        {/each}
      </div>
    {/if}
  </div>
</header>
