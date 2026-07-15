<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import Home from '@lucide/svelte/icons/house';
  import Plus from '@lucide/svelte/icons/plus';
  import Compass from '@lucide/svelte/icons/compass';
  import Settings from '@lucide/svelte/icons/settings';
  import Upload from '@lucide/svelte/icons/upload';
  import { workspaceAddAvailability } from '$lib/application/workspaceAddAvailability';
  import { mobileShellNavigationItems, shellAddHref, type ShellNavigationDestination, type ShellNavigationIcon } from '$lib/application/workspaceShellNavigation';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { SettingsSection } from '$lib/application/workspaceRoute';
  import type { WorkspaceMode } from '$lib/domain/inventory';
  import type { Component } from 'svelte';

  let {
    mode,
    selectedTenantId,
    selectedInventoryId,
    settingsSection,
    canCreateAsset,
    onModeChange,
    onOpenAdd
  }: {
    mode: WorkspaceMode;
    selectedTenantId: string;
    selectedInventoryId: string;
    settingsSection: SettingsSection;
    canCreateAsset: boolean;
    onModeChange: (mode: WorkspaceMode) => void;
    onOpenAdd: () => void;
  } = $props();

  const addDeniedNoteId = 'mobile-add-denied';
  let addAvailability = $derived(workspaceAddAvailability({ hasInventory: selectedInventoryId.length > 0, canCreateAsset }));
  let navigationItems = $derived(
    mobileShellNavigationItems({
      mode,
      tenantId: selectedTenantId || null,
      inventoryId: selectedInventoryId || null,
      settingsSection
    })
  );

  const destinationIcons: Record<ShellNavigationIcon, Component> = {
    home: Home,
    browse: Compass,
    import: Upload,
    settings: Settings
  };

  function addHref(): string {
    return shellAddHref('item', selectedTenantId || null, selectedInventoryId || null);
  }

  function openMode(event: MouseEvent, destination: ShellNavigationDestination): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onModeChange(destination.mode);
  }

  function openAdd(event: MouseEvent): void {
    if (!addAvailability.canOpen) {
      return;
    }
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onOpenAdd();
  }
</script>

<nav class="mobile-nav" aria-label="Mobile navigation">
  {#each navigationItems.slice(0, 2) as destination}
    {@const Icon = destinationIcons[destination.icon]}
    <Button.Root
      href={destination.href}
      variant={destination.current ? 'secondary' : 'ghost'}
      size="sm"
      aria-current={destination.current ? 'page' : undefined}
      onclick={(event) => openMode(event, destination)}
    ><Icon /> {destination.label}</Button.Root>
  {/each}
  <Button.Root
    href={addHref()}
    class="mobile-add"
    disabled={!addAvailability.canOpen}
    aria-label="Add asset"
    aria-describedby={addAvailability.disabledReason ? addDeniedNoteId : undefined}
    onclick={openAdd}
  ><Plus /></Button.Root>
  {#if addAvailability.disabledReason}
    <p id={addDeniedNoteId} class="visually-hidden" role="note">{addAvailability.disabledReason}</p>
  {/if}
  {#each navigationItems.slice(2) as destination}
    {@const Icon = destinationIcons[destination.icon]}
    <Button.Root
      href={destination.href}
      variant={destination.current ? 'secondary' : 'ghost'}
      size="sm"
      aria-current={destination.current ? 'page' : undefined}
      onclick={(event) => openMode(event, destination)}
    ><Icon /> {destination.label}</Button.Root>
  {/each}
</nav>
