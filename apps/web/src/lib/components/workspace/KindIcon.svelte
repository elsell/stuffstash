<script lang="ts">
  import Box from '@lucide/svelte/icons/box';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import Package from '@lucide/svelte/icons/package';
  import type { AssetKind } from '$lib/domain/inventory';

  let { kind }: { kind: AssetKind } = $props();
</script>

{#if kind === 'location'}
  <MapPin aria-hidden="true" />
{:else if kind === 'container'}
  <Box aria-hidden="true" />
{:else}
  <Package aria-hidden="true" />
{/if}
