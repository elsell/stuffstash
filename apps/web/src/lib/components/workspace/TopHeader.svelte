<script lang="ts">
  import { tick } from 'svelte';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { workspaceAddAvailability } from '$lib/application/workspaceAddAvailability';
  import { searchAssetHref } from '$lib/application/workspaceSearch';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import type { Asset, AssetKind, Inventory, Tenant } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import SearchSuggestions from './SearchSuggestions.svelte';
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
  let addMenuRegion = $state<HTMLElement | null>(null);
  let addTrigger = $state<HTMLButtonElement | null>(null);
  let addMenuElement = $state<HTMLElement | null>(null);
  let visibleSuggestions = $derived(searchFocused && query.trim().length > 0 ? suggestions.slice(0, 6) : []);
  let addAvailability = $derived(workspaceAddAvailability({ hasInventory: !!inventory, canCreateAsset }));
  const suggestionIdPrefix = 'global-search-suggestion';
  const addDeniedNoteId = 'header-add-denied';
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
    closeAddMenu(false);
    onOpenAdd(kind);
  }

  function addKindHref(kind: AssetKind): string {
    return workspaceRouteHref({ action: 'add', addKind: kind }, selectedTenantId || null, selectedInventoryId || null);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }

  function openSuggestion(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    query = asset.title;
    activeSuggestionIndex = -1;
    searchFocused = false;
    onOpenAsset(asset);
  }

  function suggestionId(index: number): string {
    return `${suggestionIdPrefix}-${index}`;
  }

  function suggestionElement(index: number): HTMLElement | null {
    if (typeof document === 'undefined') {
      return null;
    }
    return document.getElementById(suggestionId(index));
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

  function toggleAddMenu(): void {
    if (addMenuOpen) {
      closeAddMenu();
      return;
    }
    addMenuOpen = true;
    void tick().then(() => firstAddMenuItem()?.focus());
  }

  function closeAddMenu(restoreFocus = true): void {
    addMenuOpen = false;
    if (restoreFocus) {
      void tick().then(() => addTrigger?.focus());
    }
  }

  function firstAddMenuItem(): HTMLElement | null {
    return addMenuElement?.querySelector<HTMLElement>('a[href], button:not([disabled])') ?? null;
  }

  function handleAddMenuKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      closeAddMenu();
    }
  }

  function handleAddMenuFocusout(event: FocusEvent): void {
    const nextTarget = event.relatedTarget instanceof Node ? event.relatedTarget : null;
    if (nextTarget && addMenuRegion?.contains(nextTarget)) {
      return;
    }
    window.setTimeout(() => {
      const activeElement = document.activeElement;
      if (activeElement && addMenuRegion?.contains(activeElement)) {
        return;
      }
      closeAddMenu(false);
    }, 0);
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
    <SearchSuggestions
      id="global-search-suggestions"
      idPrefix={suggestionIdPrefix}
      suggestions={visibleSuggestions}
      activeIndex={activeSuggestionIndex}
      assetHref={searchAssetHref}
      onFocusIndex={(index) => { activeSuggestionIndex = index; }}
      onSuggestionKeydown={handleSuggestionKeydown}
      onOpen={openSuggestion}
    />
  </div>
  <div bind:this={addMenuRegion} class="header-add-wrap" onfocusout={handleAddMenuFocusout}>
    <Button.Root
      bind:ref={addTrigger}
      class="header-add"
      disabled={!addAvailability.canOpen}
      aria-describedby={addAvailability.disabledReason ? addDeniedNoteId : undefined}
      aria-expanded={addMenuOpen}
      aria-controls="header-add-menu"
      onclick={toggleAddMenu}
      onkeydown={handleAddMenuKeydown}
    >
      <Plus /> Add
    </Button.Root>
    {#if addAvailability.disabledReason}
      <p id={addDeniedNoteId} class="visually-hidden" role="note">{addAvailability.disabledReason}</p>
    {/if}
    {#if addMenuOpen}
      <div bind:this={addMenuElement} id="header-add-menu" class="add-menu" aria-label="Add asset kind">
        {#each addKinds as kind}
          <Button.Root href={addKindHref(kind)} variant="ghost" class="add-menu-item" onkeydown={handleAddMenuKeydown} onclick={(event) => chooseAddKind(event, kind)}>
            {assetKindLabel(kind)}
          </Button.Root>
        {/each}
      </div>
    {/if}
  </div>
</header>
