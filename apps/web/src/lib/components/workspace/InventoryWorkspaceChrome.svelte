<script lang="ts" module>
  import type { Snippet } from 'svelte';
  import type { SettingsSection } from '$lib/application/workspaceRoute';
  import type { Asset, AssetKind, Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';

  export type InventoryWorkspaceChromeProps = {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    selectedInventory: Inventory | null;
    mode: WorkspaceMode;
    settingsSection: SettingsSection;
    userLabel: string;
    searchSuggestions: Asset[];
    searchQuery: string;
    canCreateAsset: boolean;
    modalOpen?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onModeChange: (mode: WorkspaceMode) => void;
    onSearch: () => void;
    onOpenSearchAsset: (asset: Asset) => void;
    onOpenAdd: (kind: AssetKind) => void;
    onOpenAccountSettings: () => void;
    onSignOut: () => void;
    children?: Snippet;
  };
</script>

<script lang="ts">
  import MobileNav from './MobileNav.svelte';
  import SideNav from './SideNav.svelte';
  import TopHeader from './TopHeader.svelte';

  let mobileSurfaceOpen = $state(false);

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    selectedInventory,
    mode,
    settingsSection,
    userLabel,
    searchSuggestions,
    searchQuery = $bindable(''),
    canCreateAsset,
    modalOpen = false,
    onSelectTenant,
    onSelectInventory,
    onModeChange,
    onSearch,
    onOpenSearchAsset,
    onOpenAdd,
    onOpenAccountSettings,
    onSignOut,
    children
  }: InventoryWorkspaceChromeProps = $props();

  let showMobileNavigation = $derived(mode !== 'asset' && mode !== 'location');
</script>

<div class="product-shell" inert={modalOpen ? true : undefined} aria-hidden={modalOpen ? 'true' : undefined}>
  <SideNav
    {tenants}
    {inventories}
    {selectedTenantId}
    {selectedInventoryId}
    {mode}
    {settingsSection}
    {userLabel}
    {onSelectTenant}
    {onSelectInventory}
    {onModeChange}
    {onOpenAccountSettings}
    {onSignOut}
  />

  <div class="workspace-column">
    <TopHeader
      {tenants}
      {inventories}
      {selectedTenantId}
      inventory={selectedInventory}
      suggestions={searchSuggestions}
      bind:query={searchQuery}
      {canCreateAsset}
      {userLabel}
      showSearch={mode !== 'search' && mode !== 'browse'}
      {onSelectTenant}
      {onSelectInventory}
      {onSearch}
      onOpenAsset={onOpenSearchAsset}
      {onOpenAdd}
      onOpenSettings={onOpenAccountSettings}
      {onSignOut}
      onMobileSurfaceOpenChange={(open) => { mobileSurfaceOpen = open; }}
    />

    <main class="workspace-route-content" inert={mobileSurfaceOpen ? true : undefined} aria-hidden={mobileSurfaceOpen ? 'true' : undefined}>
      {@render children?.()}
    </main>
  </div>

  {#if showMobileNavigation}
    <div class="mobile-nav-shell" inert={mobileSurfaceOpen ? true : undefined} aria-hidden={mobileSurfaceOpen ? 'true' : undefined}>
      <MobileNav
        {mode}
        {selectedTenantId}
        {selectedInventoryId}
        {settingsSection}
        {canCreateAsset}
        {onModeChange}
        onOpenAdd={() => onOpenAdd('item')}
      />
    </div>
  {/if}
</div>
