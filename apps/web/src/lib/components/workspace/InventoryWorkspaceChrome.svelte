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
    disablePortal?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onCreateTenantWithInventory?: (input: { tenantName: string; inventoryName: string }) => Promise<void>;
    onCreateInventory?: (tenantId: string, inventoryName: string) => Promise<void>;
    onModeChange: (mode: WorkspaceMode) => void;
    onSearch: () => void;
    onOpenSearchAsset: (asset: Asset) => void;
    onOpenAdd: (kind: AssetKind, parentAssetId?: string | null, opener?: HTMLElement | null) => void;
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
    disablePortal = false,
    onSelectTenant,
    onSelectInventory,
    onCreateTenantWithInventory,
    onCreateInventory,
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
    {onCreateTenantWithInventory}
    {onCreateInventory}
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
      {disablePortal}
      showSearch={mode !== 'browse'}
      {onSelectTenant}
      {onSelectInventory}
      {onCreateTenantWithInventory}
      {onCreateInventory}
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
        onOpenAdd={(opener) => onOpenAdd('item', null, opener)}
      />
    </div>
  {/if}
</div>
