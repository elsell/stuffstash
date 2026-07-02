<script lang="ts">
  import Home from '@lucide/svelte/icons/house';
  import Settings from '@lucide/svelte/icons/settings';
  import Upload from '@lucide/svelte/icons/upload';
  import LogOut from '@lucide/svelte/icons/log-out';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mode,
    userLabel,
    onSelectTenant,
    onSelectInventory,
    onModeChange,
    onSignOut
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    mode: WorkspaceMode;
    userLabel: string;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onModeChange: (mode: WorkspaceMode) => void;
    onSignOut: () => void;
  } = $props();

  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId));
</script>

<aside class="side-nav" aria-label="Workspace navigation">
  <div class="brand-lockup compact-lockup">
    <div class="brand-mark" aria-hidden="true"><span></span></div>
    <div>
      <strong>Stuff Stash</strong>
      <p>{selectedTenant?.name ?? 'Home'}</p>
    </div>
  </div>

  <WorkspaceContextSwitcher
    {tenants}
    {inventories}
    {selectedTenantId}
    {selectedInventoryId}
    {onSelectTenant}
    {onSelectInventory}
    onOpenSettings={() => onModeChange('settings')}
  />

  <nav class="nav-list" aria-label="Inventory destinations">
    <p class="nav-eyebrow">Inventory</p>
    <Button.Root variant={mode === 'home' ? 'secondary' : 'ghost'} class="nav-button" onclick={() => onModeChange('home')}>
      <Home /> Home
    </Button.Root>
    <Button.Root variant={mode === 'import' ? 'secondary' : 'ghost'} class="nav-button" onclick={() => onModeChange('import')}>
      <Upload /> Import
    </Button.Root>
    <Button.Root variant={mode === 'settings' ? 'secondary' : 'ghost'} class="nav-button" onclick={() => onModeChange('settings')}>
      <Settings /> Settings
    </Button.Root>
  </nav>

  <div class="side-nav-footer">
    <p>{userLabel}</p>
    <Button.Root variant="ghost" size="sm" onclick={onSignOut}><LogOut /> Sign out</Button.Root>
  </div>
</aside>
