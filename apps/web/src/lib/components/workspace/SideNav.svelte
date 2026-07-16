<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import Home from '@lucide/svelte/icons/house';
  import Compass from '@lucide/svelte/icons/compass';
  import Settings from '@lucide/svelte/icons/settings';
  import Upload from '@lucide/svelte/icons/upload';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { desktopShellNavigationGroups, shellModeHref, type ShellNavigationDestination, type ShellNavigationIcon } from '$lib/application/workspaceShellNavigation';
  import type { SettingsSection } from '$lib/application/workspaceRoute';
  import type { Inventory, Tenant, WorkspaceMode } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';
  import AccountMenu from './AccountMenu.svelte';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mode,
    settingsSection,
    userLabel,
    disableAccountPortal = false,
    onSelectTenant,
    onSelectInventory,
    onCreateTenantWithInventory,
    onCreateInventory,
    onModeChange,
    onOpenAccountSettings,
    onSignOut
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    mode: WorkspaceMode;
    settingsSection: SettingsSection;
    userLabel: string;
    disableAccountPortal?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onCreateTenantWithInventory?: (input: { tenantName: string; inventoryName: string }) => Promise<void>;
    onCreateInventory?: (tenantId: string, inventoryName: string) => Promise<void>;
    onModeChange: (mode: WorkspaceMode) => void;
    onOpenAccountSettings: () => void;
    onSignOut: () => void;
  } = $props();

  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId));
  let selectedInventory = $derived(inventories.find((inventory) => inventory.id === selectedInventoryId) ?? null);
  let navigationGroups = $derived(
    desktopShellNavigationGroups({
      mode,
      tenantId: selectedTenantId || null,
      inventoryId: selectedInventoryId || null,
      inventory: selectedInventory,
      settingsSection
    })
  );
  let accountSettingsHref = $derived(shellModeHref('settings', selectedTenantId || null, selectedInventoryId || null));

  const destinationIcons: Record<ShellNavigationIcon, Component> = {
    home: Home,
    browse: Compass,
    import: Upload,
    settings: Settings
  };

  function openDestination(event: MouseEvent, destination: ShellNavigationDestination): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onModeChange(destination.mode);
  }
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
    {onCreateTenantWithInventory}
    {onCreateInventory}
  />

  <nav class="side-nav-groups" aria-label="Inventory destinations">
    {#each navigationGroups as group}
      <div class="nav-section" aria-labelledby={`${group.id}-nav-label`}>
      <p id={`${group.id}-nav-label`} class="nav-eyebrow">{group.label}</p>
      <div class="nav-list">
        {#each group.destinations as destination}
          {@const Icon = destinationIcons[destination.icon]}
          <Button.Root
            href={destination.href}
            variant={destination.current ? 'secondary' : 'ghost'}
            class="nav-button"
            aria-current={destination.current ? 'page' : undefined}
            onclick={(event) => openDestination(event, destination)}
          >
            <Icon aria-hidden="true" />
            <span>
              <strong>{destination.label}</strong>
              <small>{destination.description}</small>
            </span>
          </Button.Root>
        {/each}
      </div>
      </div>
    {/each}
  </nav>

  <div class="side-nav-footer">
    <AccountMenu
      {userLabel}
      settingsHref={accountSettingsHref}
      onOpenSettings={onOpenAccountSettings}
      {onSignOut}
      disablePortal={disableAccountPortal}
    />
  </div>
</aside>
