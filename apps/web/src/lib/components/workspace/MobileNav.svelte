<script lang="ts">
  import Home from '@lucide/svelte/icons/house';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import Settings from '@lucide/svelte/icons/settings';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import * as Button from '$lib/components/ui/button/index.js';
  import { workspaceRouteHref, type SettingsSection, type WorkspaceRouteState } from '$lib/application/workspaceRoute';
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

  function modeHref(nextMode: WorkspaceMode): string {
    const route: Partial<WorkspaceRouteState> = { mode: nextMode };
    if (nextMode === 'settings') {
      route.settingsSection = settingsSection;
    }
    return workspaceRouteHref(route, selectedTenantId || null, selectedInventoryId || null);
  }

  function addHref(): string {
    return workspaceRouteHref({ action: 'add', addKind: 'item' }, selectedTenantId || null, selectedInventoryId || null);
  }

  function openMode(event: MouseEvent, nextMode: WorkspaceMode): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onModeChange(nextMode);
  }

  function openAdd(event: MouseEvent): void {
    if (!canCreateAsset) {
      return;
    }
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onOpenAdd();
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
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
  <Button.Root href={addHref()} class="mobile-add" disabled={!canCreateAsset} aria-label="Add asset" onclick={openAdd}><Plus /></Button.Root>
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
