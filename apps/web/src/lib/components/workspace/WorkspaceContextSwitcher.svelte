<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { tick } from 'svelte';
  import Building2 from '@lucide/svelte/icons/building-2';
  import Check from '@lucide/svelte/icons/check';
  import Package from '@lucide/svelte/icons/package';
  import {
    contextSwitcherPresentation,
    inventoryContextOptions,
    tenantContextOptions,
    type InventoryContextOption
  } from '$lib/application/workspaceContextSwitching';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { Inventory, Tenant } from '$lib/domain/inventory';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mobile = false,
    onSelectTenant,
    onSelectInventory,
    onOpenChange
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    mobile?: boolean;
    onSelectTenant: (tenantId: string) => void;
    onSelectInventory: (tenantId: string, inventoryId: string) => void;
    onOpenChange?: (open: boolean) => void;
  } = $props();

  let open = $state(false);
  let showingTenants = $state(false);
  let selectedTenant = $derived(tenants.find((tenant) => tenant.id === selectedTenantId) ?? null);
  let selectedInventory = $derived(inventories.find((inventory) => inventory.id === selectedInventoryId) ?? null);
  let tenantOptions = $derived(tenantContextOptions({ tenants, inventories, selectedTenantId }));
  let inventoryOptions = $derived(inventoryContextOptions({ tenants, inventories, selectedTenantId, selectedInventoryId }));
  let presentation = $derived(contextSwitcherPresentation({ selectedTenant, selectedInventory }));
  let modalContext = $derived(mobile && Boolean(onOpenChange));

  let rootElement: HTMLDivElement | null = $state(null);
  let triggerElement: HTMLButtonElement | null = $state(null);
  let panelElement: HTMLDivElement | null = $state(null);
  let lastFocusedTenantId = $state(initialSelectedTenantId());
  let restoreFocusElement: HTMLElement | null = null;

  $effect(() => {
    if (selectedTenantId === lastFocusedTenantId) {
      return;
    }
    lastFocusedTenantId = selectedTenantId;
    showingTenants = false;
    if (open) {
      void focusPanel();
    }
  });

  function toggleContext(): void {
    if (open) {
      closeContext();
      return;
    }
    restoreFocusElement = document.activeElement instanceof HTMLElement ? document.activeElement : triggerElement;
    open = true;
    onOpenChange?.(true);
    showingTenants = false;
    void focusPanel();
  }

  function closeContext(restoreFocus = true): void {
    const focusTarget = restoreFocusElement ?? triggerElement;
    open = false;
    onOpenChange?.(false);
    showingTenants = false;
    restoreFocusElement = null;
    if (restoreFocus) {
      void tick().then(() => {
        focusTarget?.focus();
      });
    }
  }

  async function focusPanel(): Promise<void> {
    await tick();
    const selectedOption = panelElement?.querySelector<HTMLElement>('[aria-current="page"], [aria-pressed="true"]');
    const firstFocusable = panelElement
      ?.querySelector<HTMLElement>('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    (selectedOption ?? firstFocusable ?? panelElement)?.focus();
  }

  function initialSelectedTenantId(): string {
    return selectedTenantId;
  }

  function chooseTenant(tenantId: string): void {
    const tenantAlreadySelected = tenantId === selectedTenantId;
    showingTenants = false;
    onSelectTenant(tenantId);
    if (tenantAlreadySelected) {
      void focusPanel();
    }
  }

  function chooseInventory(event: MouseEvent, inventory: InventoryContextOption): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    closeContext();
    onSelectInventory(inventory.tenantId, inventory.id);
  }

  function handlePanelKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      closeContext();
      return;
    }

    if (!modalContext || event.key !== 'Tab' || !panelElement) {
      return;
    }
    const focusable = Array.from(
      panelElement.querySelectorAll<HTMLElement>('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])')
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

  function handleContextFocusout(event: FocusEvent): void {
    if (mobile || !open) {
      return;
    }
    const nextTarget = event.relatedTarget instanceof Node ? event.relatedTarget : null;
    if (nextTarget && rootElement?.contains(nextTarget)) {
      return;
    }
    window.setTimeout(() => {
      const activeElement = document.activeElement;
      if (activeElement && rootElement?.contains(activeElement)) {
        return;
      }
      closeContext(false);
    }, 0);
  }
</script>

<div bind:this={rootElement} class:mobile-context={mobile} class:context-switcher={!mobile} onfocusout={handleContextFocusout}>
  <Button.Root
    bind:ref={triggerElement}
    variant="ghost"
    class={mobile ? 'context-trigger mobile-context-trigger' : 'context-trigger'}
    aria-haspopup="dialog"
    aria-expanded={open}
    onclick={toggleContext}
  >
    <span class="identity-icon" data-kind="inventory" aria-hidden="true"><Package /></span>
    <span class="identity-copy">
      <strong>{presentation.triggerInventoryLabel}</strong>
      <small>{presentation.triggerTenantLabel}</small>
    </span>
  </Button.Root>

  {#if open}
    {#if mobile}
      <Button.Root variant="ghost" class="sheet-backdrop" aria-label="Close inventory switcher" onclick={() => closeContext()}></Button.Root>
    {/if}
    <div
      bind:this={panelElement}
      class:context-popover={!mobile}
      class:mobile-context-menu={mobile}
      role="dialog"
      aria-modal={modalContext ? 'true' : undefined}
      aria-label="Inventory context"
      tabindex={-1}
      onkeydown={handlePanelKeydown}
    >
      <div class="context-header">
        <span class="identity-label">
          <span class="identity-icon" data-kind="tenant" aria-hidden="true"><Building2 /></span>
          <span>{presentation.activeTenantLabel}</span>
        </span>
        {#if tenants.length > 0}
          <Button.Root variant="ghost" class={mobile ? 'context-switch-action' : undefined} onclick={() => { showingTenants = !showingTenants; }}>
            {showingTenants ? 'Back' : 'Switch tenant'}
          </Button.Root>
        {/if}
      </div>

      {#if showingTenants}
        <p class="context-section-label">Tenants</p>
        <div class="context-option-list" aria-label="Tenants">
          {#each tenantOptions as tenant}
            <Button.Root
              variant={tenant.selected ? 'secondary' : 'ghost'}
              class="context-option"
              aria-pressed={tenant.selected}
              onclick={() => chooseTenant(tenant.id)}
            >
              <span class="context-option-check" aria-hidden="true">{#if tenant.selected}<Check />{/if}</span>
              <span class="context-option-copy">
                <strong>{tenant.name}</strong>
                <small>{tenant.inventoryCountLabel}</small>
              </span>
            </Button.Root>
          {/each}
        </div>
      {:else if inventoryOptions.length > 0}
        <p class="context-section-label">Inventories</p>
        <div class="context-option-list" aria-label="Inventories">
          {#each inventoryOptions as inventory}
            <Button.Root
              href={inventory.href}
              variant={inventory.selected ? 'secondary' : 'ghost'}
              class="context-option"
              aria-current={inventory.selected ? 'page' : undefined}
              onclick={(event) => chooseInventory(event, inventory)}
            >
              <span class="context-option-check" aria-hidden="true">{#if inventory.selected}<Check />{/if}</span>
              <span class="context-option-copy">
                <strong>{inventory.name}</strong>
                <small>{inventory.tenantName}</small>
              </span>
              <span class="context-option-pill">{inventory.relationshipLabel}</span>
            </Button.Root>
          {/each}
        </div>
      {:else}
        <p class="muted small-copy">{presentation.emptyInventoryMessage}</p>
      {/if}
    </div>
  {/if}
</div>
