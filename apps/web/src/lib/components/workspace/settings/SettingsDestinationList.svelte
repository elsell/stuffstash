<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import CircleUserRound from '@lucide/svelte/icons/circle-user-round';
  import Building2 from '@lucide/svelte/icons/building-2';
  import Boxes from '@lucide/svelte/icons/boxes';
  import Users from '@lucide/svelte/icons/users';
  import Activity from '@lucide/svelte/icons/activity';
  import ListChecks from '@lucide/svelte/icons/list-checks';
  import Shapes from '@lucide/svelte/icons/shapes';
  import Tags from '@lucide/svelte/icons/tags';
  import type { Component } from 'svelte';
  import type { SettingsDestination, SettingsDestinationIcon } from '$lib/application/settingsManagementNavigation';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';

  let { label, destinations, onNavigate }: { label: string; destinations: SettingsDestination[]; onNavigate: (href: string) => void } = $props();
  const icons: Record<SettingsDestinationIcon, Component> = {
    account: CircleUserRound, tenant: Building2, inventory: Boxes, access: Users, activity: Activity,
    fields: ListChecks, 'asset-types': Shapes, tags: Tags
  };
  function navigate(event: MouseEvent, href: string): void {
    if (!shouldHandleWorkspaceLinkClick(event)) return;
    event.preventDefault();
    onNavigate(href);
  }
</script>

<nav class="settings-destination-list" aria-label={label}>
  {#each destinations as destination}
    {@const Icon = icons[destination.icon]}
    <a class="settings-destination-row" href={destination.href} onclick={(event) => navigate(event, destination.href)}>
      <span class="settings-destination-icon"><Icon aria-hidden="true" /></span>
      <span class="settings-destination-copy">
        <small>{destination.eyebrow}</small>
        <strong>{destination.label}</strong>
        <span>{destination.description}</span>
      </span>
      <ChevronRight aria-hidden="true" />
    </a>
  {/each}
</nav>
