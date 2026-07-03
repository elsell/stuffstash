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
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onModeChange: (mode: WorkspaceMode) => void;
    onSearch: () => void;
    onOpenSearchAsset: (asset: Asset) => void;
    onOpenAdd: (kind: AssetKind) => void;
    onSignOut: () => void;
    children?: Snippet;
  };
</script>

<script lang="ts">
  import MobileNav from './MobileNav.svelte';
  import SideNav from './SideNav.svelte';
  import TopHeader from './TopHeader.svelte';

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
    onSelectTenant,
    onSelectInventory,
    onModeChange,
    onSearch,
    onOpenSearchAsset,
    onOpenAdd,
    onSignOut,
    children
  }: InventoryWorkspaceChromeProps = $props();
</script>

<div class="product-shell">
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
      {onSelectTenant}
      {onSelectInventory}
      {onSearch}
      onOpenAsset={onOpenSearchAsset}
      {onOpenAdd}
    />

    {@render children?.()}
  </div>

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
