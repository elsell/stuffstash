<script lang="ts" module>
  import type { Component } from 'svelte';

  export type BusyButtonContentIcon = Component;
</script>

<script lang="ts">
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';

  type Props = {
    busy: boolean;
    label: string;
    busyLabel: string;
    icon?: BusyButtonContentIcon;
  };

  let { busy, label, busyLabel, icon }: Props = $props();
  let Icon = $derived(busy ? LoaderCircle : icon);
</script>

<span class="busy-button-content">
  {#if Icon}
    <Icon class={busy ? 'busy-button-spinner' : undefined} size={16} aria-hidden="true" />
  {/if}
  <span>{busy ? busyLabel : label}</span>
</span>

<style>
  .busy-button-content {
    align-items: center;
    display: inline-flex;
    gap: 0.4rem;
    min-width: 0;
  }

  :global(.busy-button-spinner) {
    animation: busy-button-spin 1s linear infinite;
  }

  @keyframes busy-button-spin {
    to {
      transform: rotate(360deg);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    :global(.busy-button-spinner) {
      animation: none;
    }
  }
</style>
