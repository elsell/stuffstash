<script lang="ts">
  import { tick } from 'svelte';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
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
  let activeSuggestionIndex = $state(-1);
  let addMenuOpen = $state(false);
  let searchInput = $state<HTMLInputElement | null>(null);
  let searchRegion = $state<HTMLElement | null>(null);
  let visibleSuggestions = $derived(searchFocused && query.trim().length > 0 ? suggestions.slice(0, 6) : []);
  const addKinds: AssetKind[] = ['item', 'container', 'location'];

  $effect(() => {
    if (activeSuggestionIndex >= visibleSuggestions.length) {
      activeSuggestionIndex = visibleSuggestions.length - 1;
    }
    if (visibleSuggestions.length === 0) {
      activeSuggestionIndex = -1;
    }
  });

  function chooseAddKind(event: MouseEvent, kind: AssetKind): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    addMenuOpen = false;
    onOpenAdd(kind);
  }

  function addKindHref(kind: AssetKind): string {
    return workspaceRouteHref({ action: 'add', addKind: kind }, selectedTenantId || null, selectedInventoryId || null);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }

  function openSuggestion(asset: Asset): void {
    query = asset.title;
    activeSuggestionIndex = -1;
    searchFocused = false;
    onOpenAsset(asset);
  }

  function suggestionId(index: number): string {
    return `global-search-suggestion-${index}`;
  }

  function suggestionElement(index: number): HTMLButtonElement | null {
    if (typeof document === 'undefined') {
      return null;
    }
    return document.getElementById(suggestionId(index)) as HTMLButtonElement | null;
  }

  async function focusSuggestion(index: number): Promise<void> {
    activeSuggestionIndex = index;
    await tick();
    suggestionElement(index)?.focus();
  }

  function handleSearchKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      if (visibleSuggestions.length > 0) {
        event.preventDefault();
      }
      searchFocused = false;
      activeSuggestionIndex = -1;
      return;
    }

    if (visibleSuggestions.length === 0) {
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      searchFocused = true;
      void focusSuggestion(0);
      return;
    }
  }

  function handleSuggestionKeydown(event: KeyboardEvent, index: number): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      searchInput?.focus();
      closeSearchSuggestions();
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      void focusSuggestion((index + 1) % visibleSuggestions.length);
      return;
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault();
      if (index === 0) {
        activeSuggestionIndex = -1;
        searchInput?.focus();
      } else {
        void focusSuggestion(index - 1);
      }
    }
  }

  function closeSearchSuggestions(): void {
    searchFocused = false;
    activeSuggestionIndex = -1;
  }

  function handleSearchFocusout(event: FocusEvent): void {
    const nextTarget = event.relatedTarget instanceof Node ? event.relatedTarget : null;
    if (nextTarget && searchRegion?.contains(nextTarget)) {
      return;
    }
    window.setTimeout(() => {
      const activeElement = document.activeElement;
      if (activeElement && searchRegion?.contains(activeElement)) {
        return;
      }
      closeSearchSuggestions();
    }, 120);
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
  <div bind:this={searchRegion} class="global-search-wrap" onfocusout={handleSearchFocusout}>
    <form class="global-search" onsubmit={(event) => { event.preventDefault(); onSearch(); }}>
      <Search aria-hidden="true" />
      <Input
        bind:ref={searchInput}
        bind:value={query}
        placeholder="Search this inventory"
        aria-label="Search this inventory"
        onfocus={() => { searchFocused = true; }}
        onkeydown={handleSearchKeydown}
      />
      <Button.Root type="submit" variant="ghost" size="icon-sm" aria-label="Run search"><Search /></Button.Root>
    </form>
    {#if visibleSuggestions.length > 0}
      <ul id="global-search-suggestions" class="search-suggestions" aria-label="Search suggestions">
        {#each visibleSuggestions as suggestion, index}
          <li>
            <Button.Root
              id={suggestionId(index)}
              variant="ghost"
              class="suggestion-row"
              data-active={activeSuggestionIndex === index}
              aria-label={`Open ${suggestion.title}`}
              onfocus={() => { activeSuggestionIndex = index; }}
              onkeydown={(event) => handleSuggestionKeydown(event, index)}
              onpointerenter={() => { activeSuggestionIndex = index; }}
              onclick={() => { openSuggestion(suggestion); }}
            >
              <span>
                <strong>{suggestion.title}</strong>
                <small>{assetKindLabel(suggestion.kind)}</small>
              </span>
            </Button.Root>
          </li>
        {/each}
      </ul>
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
          <Button.Root href={addKindHref(kind)} variant="ghost" class="add-menu-item" onclick={(event) => chooseAddKind(event, kind)}>
            {assetKindLabel(kind)}
          </Button.Root>
        {/each}
      </div>
    {/if}
  </div>
</header>
