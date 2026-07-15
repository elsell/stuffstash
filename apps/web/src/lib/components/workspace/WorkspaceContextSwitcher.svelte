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
  import * as Popover from '$lib/components/ui/popover/index.js';
  import * as Sheet from '$lib/components/ui/sheet/index.js';
  import type { Inventory, Tenant } from '$lib/domain/inventory';
  import { cn } from '$lib/utils.js';

  let {
    tenants,
    inventories,
    selectedTenantId,
    selectedInventoryId,
    mobile = false,
    disablePortal = false,
    onSelectTenant,
    onSelectInventory,
    onOpenChange: notifyOpenChange
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    selectedTenantId: string;
    selectedInventoryId: string;
    mobile?: boolean;
    disablePortal?: boolean;
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

  let panelElement: HTMLDivElement | null = $state(null);
  let lastFocusedTenantId = $state(initialSelectedTenantId());

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

  function handleOpenChange(nextOpen: boolean): void {
    open = nextOpen;
    notifyOpenChange?.(nextOpen);
    if (!nextOpen) {
      showingTenants = false;
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
    handleOpenChange(false);
    onSelectInventory(inventory.tenantId, inventory.id);
  }

  function contextTriggerClass(primitiveClass: unknown, mobileTrigger = false): string {
    return cn(
      typeof primitiveClass === 'string' ? primitiveClass : undefined,
      'context-trigger',
      mobileTrigger && 'mobile-context-trigger'
    );
  }
</script>

{#snippet trigger()}
  <span class="identity-icon" data-kind="inventory" aria-hidden="true"><Package /></span>
  <span class="identity-copy">
    <strong>{presentation.triggerInventoryLabel}</strong>
    <small>{presentation.triggerTenantLabel}</small>
  </span>
{/snippet}

{#snippet contextOptions()}
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
{/snippet}

{#if mobile}
  <Sheet.Root bind:open onOpenChange={handleOpenChange}>
    <Sheet.Trigger>
      {#snippet child({ props })}
        <Button.Root {...props} variant="ghost" class={contextTriggerClass(props.class, true)}>
          {@render trigger()}
        </Button.Root>
      {/snippet}
    </Sheet.Trigger>
    {#if open || !disablePortal}
      <Sheet.Content
        bind:ref={panelElement}
        side="bottom"
        forceMount={disablePortal}
        class="max-h-[min(80dvh,42rem)] gap-0 rounded-t-2xl"
        portalProps={{ disabled: disablePortal }}
        onOpenAutoFocus={(event) => { event.preventDefault(); void focusPanel(); }}
      >
        <Sheet.Header class="context-sheet-header">
          <Sheet.Title>Inventory context</Sheet.Title>
          <Sheet.Description class="sr-only">Choose the tenant and inventory to use.</Sheet.Description>
        </Sheet.Header>
        <div class="grid min-h-0 gap-2.5 overflow-y-auto p-4 pb-[max(1rem,env(safe-area-inset-bottom))]">
          {@render contextOptions()}
        </div>
      </Sheet.Content>
    {/if}
  </Sheet.Root>
{:else}
  <Popover.Root bind:open onOpenChange={handleOpenChange}>
    <Popover.Trigger>
      {#snippet child({ props })}
        <Button.Root {...props} variant="ghost" class={contextTriggerClass(props.class)}>
          {@render trigger()}
        </Button.Root>
      {/snippet}
    </Popover.Trigger>
    {#if open || !disablePortal}
      <Popover.Content
        bind:ref={panelElement}
        align="start"
        sideOffset={8}
        forceMount={disablePortal}
        class="w-[min(360px,calc(100vw-32px))] gap-2.5 p-2.5"
        portalProps={{ disabled: disablePortal }}
        onOpenAutoFocus={(event) => { event.preventDefault(); void focusPanel(); }}
      >
        {@render contextOptions()}
      </Popover.Content>
    {/if}
  </Popover.Root>
{/if}
