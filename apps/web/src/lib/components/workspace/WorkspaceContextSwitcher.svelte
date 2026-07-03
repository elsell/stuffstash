<script lang="ts">
  import Settings from '@lucide/svelte/icons/settings';
  import * as Button from '$lib/components/ui/button/index.js';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import type { Inventory, Tenant } from '$lib/domain/inventory';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mobile = false,
    onSelectTenant,
    onSelectInventory,
    onOpenSettings
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    mobile?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onOpenSettings: () => void;
  } = $props();

  let open = $state(false);
  let showingTenants = $state(false);
  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId) ?? null);
  let selectedInventory = $derived(inventories.find((inventory) => inventory.id === selectedInventoryId) ?? null);

  let sheetElement: HTMLDivElement | null = $state(null);

  function chooseTenant(tenantId: string): void {
    showingTenants = false;
    onSelectTenant(tenantId);
    if (!mobile) {
      open = false;
    }
  }

  function chooseInventory(inventory: Inventory): void {
    open = false;
    onSelectInventory(inventory.tenantId, inventory.id);
  }

  function settings(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    open = false;
    onOpenSettings();
  }

  function settingsHref(): string {
    return workspaceRouteHref({ mode: 'settings' }, selectedTenantId || null, selectedInventoryId || null);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }

  function handleSheetKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      open = false;
      return;
    }
    if (event.key !== 'Tab' || !sheetElement) {
      return;
    }
    const focusable = Array.from(
      sheetElement.querySelectorAll<HTMLElement>('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])')
    ).filter((element) => !element.hasAttribute('disabled'));
    const first = focusable[0];
    const last = focusable.at(-1);
    if (!first || !last) {
      return;
    }
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }
</script>

<div class:mobile-context={mobile} class:context-switcher={!mobile}>
  {#if mobile}
    <Button.Root variant="ghost" class="mobile-context-trigger" aria-expanded={open} onclick={() => { open = !open; }}>
      <span>
        <strong>{selectedInventory?.name ?? 'No inventory'}</strong>
        <small>{selectedTenant?.name ?? 'No tenant'}</small>
      </span>
    </Button.Root>
  {:else}
    <p class="eyebrow">Inventory</p>
    <strong>{selectedInventory?.name ?? 'No inventory'}</strong>
    <div class="tenant-row">
      <span>{selectedTenant?.name ?? 'No tenant'}</span>
      {#if tenants.length > 1}
        <Button.Root variant="ghost" size="sm" onclick={() => { showingTenants = !showingTenants; }}>
          {showingTenants ? 'Inventories' : 'Switch tenant'}
        </Button.Root>
      {/if}
    </div>
  {/if}

  {#if !mobile || open}
    {#if mobile}
      <Button.Root variant="ghost" class="sheet-backdrop" tabindex={-1} aria-hidden="true" onclick={() => { open = false; }}></Button.Root>
      <div
        bind:this={sheetElement}
        class="mobile-context-menu"
        role="dialog"
        aria-modal="true"
        aria-label="Inventory context"
        tabindex={-1}
        onkeydown={handleSheetKeydown}
      >
        <div class="tenant-row">
          <span>{selectedTenant?.name ?? 'No tenant'}</span>
          {#if tenants.length > 1}
            <Button.Root variant="ghost" size="sm" onclick={() => { showingTenants = !showingTenants; }}>
              {showingTenants ? 'Inventories' : 'Switch tenant'}
            </Button.Root>
          {/if}
        </div>
        <div class="context-caption">{selectedInventory?.name ?? 'No inventory selected'}</div>
        {#if showingTenants}
          <div class="inventory-menu" aria-label="Tenants">
            {#each tenants as tenant}
              <Button.Root
                variant={tenant.id === selectedTenantId ? 'secondary' : 'ghost'}
                class="nav-button"
                onclick={() => chooseTenant(tenant.id)}
              >
                {tenant.name}
              </Button.Root>
            {/each}
          </div>
        {:else if inventories.length > 0}
          <div class="inventory-menu" aria-label="Inventories">
            {#each inventories as inventory}
              <Button.Root
                variant={inventory.id === selectedInventoryId ? 'secondary' : 'ghost'}
                class="nav-button"
                onclick={() => chooseInventory(inventory)}
              >
                {inventory.name}
              </Button.Root>
            {/each}
          </div>
        {:else}
          <p class="muted small-copy">No inventories in this tenant.</p>
        {/if}
        <Button.Root href={settingsHref()} variant="outline" class="nav-button" onclick={settings}><Settings /> Inventory settings</Button.Root>
      </div>
    {:else}
      {#if showingTenants}
        <div class="inventory-menu" aria-label="Tenants">
          {#each tenants as tenant}
            <Button.Root
              variant={tenant.id === selectedTenantId ? 'secondary' : 'ghost'}
              class="nav-button"
              onclick={() => chooseTenant(tenant.id)}
            >
              {tenant.name}
            </Button.Root>
          {/each}
        </div>
      {:else if inventories.length > 0}
        <div class="inventory-menu" aria-label="Inventories">
          {#each inventories as inventory}
            <Button.Root
              variant={inventory.id === selectedInventoryId ? 'secondary' : 'ghost'}
              class="nav-button"
              onclick={() => chooseInventory(inventory)}
            >
              {inventory.name}
            </Button.Root>
          {/each}
        </div>
      {:else}
        <p class="muted small-copy">No inventories in this tenant.</p>
      {/if}
      <Button.Root href={settingsHref()} variant="outline" class="nav-button" onclick={settings}><Settings /> Inventory settings</Button.Root>
    {/if}
  {/if}
</div>
