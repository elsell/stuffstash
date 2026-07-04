<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import Home from '@lucide/svelte/icons/house';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import Settings from '@lucide/svelte/icons/settings';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import { workspaceAddAvailability } from '$lib/application/workspaceAddAvailability';
  import { shellAddHref, shellModeHref, type ShellWorkspaceMode } from '$lib/application/workspaceShellNavigation';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { SettingsSection } from '$lib/application/workspaceRoute';
  import type { WorkspaceMode } from '$lib/domain/inventory';

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

  function modeHref(nextMode: ShellWorkspaceMode): string {
    return shellModeHref(nextMode, selectedTenantId || null, selectedInventoryId || null, settingsSection);
  }

  function addHref(): string {
    return shellAddHref('item', selectedTenantId || null, selectedInventoryId || null);
  }

  function openMode(event: MouseEvent, nextMode: WorkspaceMode): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onModeChange(nextMode);
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
  <Button.Root
    href={modeHref('home')}
    variant={mode === 'home' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'home' ? 'page' : undefined}
    onclick={(event) => openMode(event, 'home')}
  ><Home /> Home</Button.Root>
  <Button.Root
    href={modeHref('search')}
    variant={mode === 'search' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'search' ? 'page' : undefined}
    onclick={(event) => openMode(event, 'search')}
  ><Search /> Search</Button.Root>
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
  <Button.Root
    href={modeHref('locations')}
    variant={mode === 'locations' || mode === 'location' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'locations' || mode === 'location' ? 'page' : undefined}
    onclick={(event) => openMode(event, 'locations')}
  ><MapPin /> Places</Button.Root>
  <Button.Root
    href={modeHref('settings')}
    variant={mode === 'settings' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'settings' ? 'page' : undefined}
    onclick={(event) => openMode(event, 'settings')}
  ><Settings /> Settings</Button.Root>
</nav>
