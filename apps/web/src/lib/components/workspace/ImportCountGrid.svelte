<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CircleDashed from '@lucide/svelte/icons/circle-dashed';
  import FileImage from '@lucide/svelte/icons/file-image';
  import MapPin from '@lucide/svelte/icons/map-pin';
  import PackageCheck from '@lucide/svelte/icons/package-check';
  import Rows3 from '@lucide/svelte/icons/rows-3';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { CountCell } from './importWorkspacePresentation';

  type Props = {
    cells: CountCell[];
    actionForCell?: (cell: CountCell) => void;
  };

  let { cells, actionForCell }: Props = $props();

  function tileClass(count: CountCell): string {
    const classes = ['summary-tile'];
    if (count.muted) classes.push('muted-count');
    if (count.tone) classes.push(count.tone);
    if (isActionable(count)) classes.push('actionable');
    return classes.join(' ');
  }

  function isActionable(count: CountCell): boolean {
    return Boolean(actionForCell && count.actionLabel && count.value > 0);
  }

  function iconForCell(count: CountCell) {
    const label = count.label.toLowerCase();
    if (count.tone === 'warning' || label.includes('warning')) return AlertCircle;
    if (count.tone === 'action' || label.includes('blocking')) return AlertCircle;
    if (label.includes('location')) return MapPin;
    if (label.includes('photo') || label.includes('file')) return FileImage;
    if (label.includes('asset') || label.includes('record')) return PackageCheck;
    if (label.includes('skip') || label.includes('duplicate') || count.muted) return CircleDashed;
    return Rows3;
  }
</script>

<div class="summary-grid">
  {#each cells as count}
    {@const Icon = iconForCell(count)}
    {#if isActionable(count)}
      <Button.Root variant="ghost" class={tileClass(count)} onclick={() => actionForCell?.(count)} aria-label={count.actionLabel}>
        <span class="summary-icon"><Icon size={16} aria-hidden="true" /></span>
        <strong>{count.value}</strong>
        <span>{count.label}</span>
      </Button.Root>
    {:else}
      <div class={tileClass(count)}>
        <span class="summary-icon"><Icon size={16} aria-hidden="true" /></span>
        <strong>{count.value}</strong>
        <span>{count.label}</span>
      </div>
    {/if}
  {/each}
</div>

<style>
  .summary-grid {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .summary-tile {
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--foreground);
    display: grid;
    gap: 0.25rem;
    justify-items: start;
    min-width: 0;
    padding: 0.75rem;
    text-align: left;
  }

  :global(.summary-tile[data-slot='button']) {
    height: auto;
    justify-content: start;
    white-space: normal;
  }

  :global(button.summary-tile) {
    cursor: pointer;
    font: inherit;
  }

  :global(button.summary-tile:hover) {
    background: color-mix(in oklab, var(--muted) 24%, transparent);
  }

  :global(button.summary-tile:focus-visible) {
    outline: 2px solid var(--ring);
    outline-offset: 2px;
  }

  .summary-icon {
    align-items: center;
    border-radius: 999px;
    color: var(--muted-foreground);
    display: inline-flex;
    height: 1.6rem;
    justify-content: center;
    width: 1.6rem;
  }

  .summary-tile.success .summary-icon {
    background: color-mix(in oklab, var(--primary) 8%, transparent);
    color: var(--primary);
  }

  .summary-tile.warning {
    border-color: color-mix(in oklab, var(--color-warning) 32%, var(--border));
  }

  .summary-tile.warning .summary-icon {
    background: color-mix(in oklab, var(--color-warning) 14%, transparent);
    color: var(--color-warning-foreground);
  }

  .summary-tile.action {
    border-color: color-mix(in oklab, var(--destructive) 34%, var(--border));
  }

  .summary-tile.action .summary-icon {
    background: color-mix(in oklab, var(--destructive) 10%, transparent);
    color: var(--destructive);
  }

  .summary-tile.muted,
  .summary-tile.muted-count {
    background: color-mix(in oklab, var(--muted) 25%, transparent);
  }

  .summary-grid strong {
    display: block;
    font-size: 1.35rem;
    line-height: 1.05;
  }

  .summary-tile > span:not(.summary-icon) {
    color: var(--muted-foreground);
    display: block;
    font-size: 0.82rem;
  }

  @media (max-width: 860px) {
    .summary-grid {
      gap: 0;
      grid-template-columns: 1fr;
    }

    .summary-tile {
      align-items: baseline;
      border: 0;
      border-top: 1px solid var(--border);
      border-radius: 0;
      display: grid;
      gap: 0.75rem;
      grid-template-columns: minmax(7rem, auto) minmax(0, 1fr);
      padding: 0.65rem 0;
    }

    .summary-tile:first-child {
      border-top: 0;
    }

    .summary-icon {
      display: none;
    }

    .summary-grid strong {
      font-size: 0.98rem;
      order: 2;
    }

    .summary-tile > span:not(.summary-icon) {
      font-size: 0.76rem;
      font-weight: 700;
      order: 1;
      text-transform: uppercase;
    }
  }
</style>
