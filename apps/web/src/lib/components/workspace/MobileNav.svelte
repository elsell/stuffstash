<script lang="ts">
  import Home from '@lucide/svelte/icons/house';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import Settings from '@lucide/svelte/icons/settings';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { WorkspaceMode } from '$lib/domain/inventory';

  let {
    mode,
    canCreateAsset,
    onModeChange,
    onOpenAdd
  }: {
    mode: WorkspaceMode;
    canCreateAsset: boolean;
    onModeChange: (mode: WorkspaceMode) => void;
    onOpenAdd: () => void;
  } = $props();
</script>

<nav class="mobile-nav" aria-label="Mobile navigation">
  <Button.Root
    variant={mode === 'home' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'home' ? 'page' : undefined}
    onclick={() => onModeChange('home')}
  ><Home /> Home</Button.Root>
  <Button.Root
    variant={mode === 'search' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'search' ? 'page' : undefined}
    onclick={() => onModeChange('search')}
  ><Search /> Search</Button.Root>
  <Button.Root class="mobile-add" disabled={!canCreateAsset} aria-label="Add asset" onclick={onOpenAdd}><Plus /></Button.Root>
  <Button.Root
    variant={mode === 'locations' || mode === 'location' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'locations' || mode === 'location' ? 'page' : undefined}
    onclick={() => onModeChange('locations')}
  ><MapPin /> Places</Button.Root>
  <Button.Root
    variant={mode === 'settings' ? 'secondary' : 'ghost'}
    size="sm"
    aria-current={mode === 'settings' ? 'page' : undefined}
    onclick={() => onModeChange('settings')}
  ><Settings /> Settings</Button.Root>
</nav>
