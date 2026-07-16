<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { workspaceAddAvailability } from '$lib/application/workspaceAddAvailability';
  import { searchAssetHref } from '$lib/application/workspaceSearch';
  import { shellModeHref } from '$lib/application/workspaceShellNavigation';
  import type { Asset, AssetKind, Inventory, Tenant } from '$lib/domain/inventory';
  import SearchSuggestions from './SearchSuggestions.svelte';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';
  import AccountMenu from './AccountMenu.svelte';
  import WorkspaceAddMenu from './WorkspaceAddMenu.svelte';

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
    disablePortal = false,
    onSelectTenant,
    onSelectInventory,
    onCreateTenantWithInventory,
    onCreateInventory,
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
    disablePortal?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onCreateTenantWithInventory?: (input: { tenantName: string; inventoryName: string }) => Promise<void>;
    onCreateInventory?: (tenantId: string, inventoryName: string) => Promise<void>;
    onSearch: () => void;
    onOpenAsset: (asset: Asset) => void;
    onOpenAdd: (kind: AssetKind, parentAssetId?: string | null, opener?: HTMLElement | null) => void;
    onOpenSettings: () => void;
    onSignOut: () => void;
    onMobileSurfaceOpenChange?: (open: boolean) => void;
  } = $props();

  let selectedInventoryId = $derived(inventory?.id ?? '');
  let searchFocused = $state(false);
  let activeSuggestionIndex = $state(-1);
  let searchInput = $state<HTMLInputElement | null>(null);
  let searchRegion = $state<HTMLElement | null>(null);
  let visibleSuggestions = $derived(searchFocused && query.trim().length > 0 ? suggestions.slice(0, 6) : []);
  let showNoSuggestions = $derived(searchFocused && query.trim().length > 0 && visibleSuggestions.length === 0);
  let addAvailability = $derived(workspaceAddAvailability({ hasInventory: !!inventory, canCreateAsset }));
  let accountSettingsHref = $derived(shellModeHref('settings', selectedTenantId || null, selectedInventoryId || null));
  const suggestionIdPrefix = 'global-search-suggestion';

  $effect(() => {
    if (activeSuggestionIndex >= visibleSuggestions.length) {
      activeSuggestionIndex = visibleSuggestions.length - 1;
    }
    if (visibleSuggestions.length === 0) {
      activeSuggestionIndex = -1;
    }
  });

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

<header class:contextual-toolbar={!showSearch} class="workspace-header">
  <div class="mobile-context">
    <WorkspaceContextSwitcher
      mobile
      {tenants}
      {inventories}
      {selectedTenantId}
      {selectedInventoryId}
      {onSelectTenant}
      {onSelectInventory}
      {onCreateTenantWithInventory}
      {onCreateInventory}
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
  {:else if inventory}
    <p class="desktop-header-context" aria-label={`Current inventory: ${inventory.name}`}>
      <small>Current inventory</small>
      <strong>{inventory.name}</strong>
    </p>
  {/if}
  <WorkspaceAddMenu
    tenantId={selectedTenantId || null}
    inventoryId={selectedInventoryId || null}
    canOpen={addAvailability.canOpen}
    disabledReason={addAvailability.disabledReason}
    {disablePortal}
    {onOpenAdd}
  />
</header>
