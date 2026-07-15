<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { tick } from 'svelte';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { workspaceAddAvailability } from '$lib/application/workspaceAddAvailability';
  import { searchAssetHref } from '$lib/application/workspaceSearch';
  import { shellAddOptions, shellModeHref, type ShellAddOption } from '$lib/application/workspaceShellNavigation';
  import type { Asset, AssetKind, Inventory, Tenant } from '$lib/domain/inventory';
  import SearchSuggestions from './SearchSuggestions.svelte';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';
  import AccountMenu from './AccountMenu.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    inventory,
    suggestions,
    query = $bindable(''),
    canCreateAsset,
    userLabel,
    showSearch = true,
    onSelectTenant,
    onSelectInventory,
    onSearch,
    onOpenAsset,
    onOpenAdd,
    onOpenSettings,
    onSignOut,
    onMobileSurfaceOpenChange
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    inventory: Inventory | null;
    suggestions: Asset[];
    query: string;
    canCreateAsset: boolean;
    userLabel: string;
    showSearch?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onSearch: () => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: (kind: AssetKind) => void;
    onOpenSettings: () => void;
    onSignOut: () => void;
    onMobileSurfaceOpenChange?: (open: boolean) => void;
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
  let showNoSuggestions = $derived(searchFocused && query.trim().length > 0 && visibleSuggestions.length === 0);
  let addAvailability = $derived(workspaceAddAvailability({ hasInventory: !!inventory, canCreateAsset }));
  let addOptions = $derived(shellAddOptions(selectedTenantId || null, selectedInventoryId || null));
  let accountSettingsHref = $derived(shellModeHref('settings', selectedTenantId || null, selectedInventoryId || null));
  const suggestionIdPrefix = 'global-search-suggestion';
  const addDeniedNoteId = 'header-add-denied';

  $effect(() => {
    if (activeSuggestionIndex >= visibleSuggestions.length) {
      activeSuggestionIndex = visibleSuggestions.length - 1;
    }
    if (visibleSuggestions.length === 0) {
      activeSuggestionIndex = -1;
    }
  });

  function chooseAddKind(event: MouseEvent, option: ShellAddOption): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    closeAddMenu(false);
    onOpenAdd(option.kind);
  }

  function openSuggestion(event: MouseEvent, asset: Asset): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    selectSuggestion(asset);
  }

  function selectSuggestion(asset: Asset): void {
    query = asset.title;
    activeSuggestionIndex = -1;
    searchFocused = false;
    onOpenAsset(asset);
  }

  function suggestionId(index: number): string {
    return `${suggestionIdPrefix}-${index}`;
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

    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault();
      searchFocused = true;
      activeSuggestionIndex = event.key === 'ArrowDown'
        ? (activeSuggestionIndex + 1) % visibleSuggestions.length
        : (activeSuggestionIndex <= 0 ? visibleSuggestions.length - 1 : activeSuggestionIndex - 1);
      return;
    }

    if (event.key === 'Enter' && activeSuggestionIndex >= 0) {
      event.preventDefault();
      selectSuggestion(visibleSuggestions[activeSuggestionIndex]!);
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
      onOpenChange={onMobileSurfaceOpenChange}
    />
  </div>
  <AccountMenu
    mobile
    {userLabel}
    settingsHref={accountSettingsHref}
    {onOpenSettings}
    {onSignOut}
    onOpenChange={onMobileSurfaceOpenChange}
  />
  {#if showSearch}
    <div bind:this={searchRegion} class="global-search-wrap" onfocusout={handleSearchFocusout}>
      <form class="global-search" onsubmit={(event) => { event.preventDefault(); closeSearchSuggestions(); onSearch(); }}>
        <Input
          bind:ref={searchInput}
          bind:value={query}
          placeholder="Search this inventory"
          aria-label="Search this inventory"
          role="combobox"
          aria-autocomplete="list"
          aria-expanded={visibleSuggestions.length > 0}
          aria-controls={visibleSuggestions.length > 0 ? 'global-search-suggestions' : undefined}
          aria-activedescendant={activeSuggestionIndex >= 0 ? suggestionId(activeSuggestionIndex) : undefined}
          onfocus={() => { searchFocused = true; }}
          oninput={() => { activeSuggestionIndex = -1; }}
          onkeydown={handleSearchKeydown}
        />
        <Button.Root type="submit" variant="ghost" size="icon" aria-label="Run search"><Search /></Button.Root>
      </form>
      <SearchSuggestions
        id="global-search-suggestions"
        idPrefix={suggestionIdPrefix}
        suggestions={visibleSuggestions}
        activeIndex={activeSuggestionIndex}
        {query}
        showEmpty={showNoSuggestions}
        assetHref={searchAssetHref}
        onFocusIndex={(index) => { activeSuggestionIndex = index; }}
        onSuggestionKeydown={() => {}}
        onOpen={openSuggestion}
      />
    </div>
  {/if}
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
        {#each addOptions as option}
          <Button.Root href={option.href} variant="ghost" class="add-menu-item" onkeydown={handleAddMenuKeydown} onclick={(event) => chooseAddKind(event, option)}>
            {option.label}
          </Button.Root>
        {/each}
      </div>
    {/if}
  </div>
</header>
