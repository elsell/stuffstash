<script lang="ts">
  import InventoryWorkspaceChrome, { type InventoryWorkspaceChromeProps } from './InventoryWorkspaceChrome.svelte';

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
    modalOpen,
    disablePortal,
    onSelectTenant,
    onSelectInventory,
    onModeChange,
    onSearch,
    onOpenSearchAsset,
    onOpenAdd,
    onOpenAccountSettings,
    onSignOut
  }: InventoryWorkspaceChromeProps = $props();
</script>

<output data-testid="bound-search-query">{searchQuery}</output>

<InventoryWorkspaceChrome
  {tenants}
  {inventories}
  {selectedTenantId}
  {selectedInventoryId}
  {selectedInventory}
  {mode}
  {settingsSection}
  {userLabel}
  {searchSuggestions}
  bind:searchQuery
  {canCreateAsset}
  {modalOpen}
  {disablePortal}
  {onSelectTenant}
  {onSelectInventory}
  {onModeChange}
  {onSearch}
  {onOpenSearchAsset}
  {onOpenAdd}
  {onOpenAccountSettings}
  {onSignOut}
>
  <section class="workspace-main" aria-labelledby="fixture-main-title">
    <h1 id="fixture-main-title">Fixture workspace</h1>
  </section>
</InventoryWorkspaceChrome>
