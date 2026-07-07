<script lang="ts">
  import type { CountCell } from './importWorkspacePresentation';

  type Props = {
    cells: CountCell[];
  };

  let { cells }: Props = $props();
</script>

<div class="summary-grid">
  {#each cells as count}
    <div class:muted-count={count.muted}><strong>{count.value}</strong><span>{` ${count.label}`}</span></div>
  {/each}
</div>

<style>
  .summary-grid {
    display: grid;
    gap: 0.75rem;
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .summary-grid div {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    padding: 0.75rem;
  }

  .summary-grid div.muted-count {
    background: hsl(var(--muted) / 0.25);
  }

  .summary-grid strong {
    display: block;
    font-size: 1.35rem;
  }

  .summary-grid span {
    color: hsl(var(--muted-foreground));
    display: block;
    font-size: 0.82rem;
  }

  @media (max-width: 860px) {
    .summary-grid {
      gap: 0;
      grid-template-columns: 1fr;
    }

    .summary-grid div {
      align-items: baseline;
      border: 0;
      border-top: 1px solid hsl(var(--border));
      border-radius: 0;
      display: grid;
      gap: 0.75rem;
      grid-template-columns: minmax(7rem, auto) minmax(0, 1fr);
      padding: 0.65rem 0;
    }

    .summary-grid div:first-child {
      border-top: 0;
    }

    .summary-grid strong {
      font-size: 0.98rem;
      order: 2;
    }

    .summary-grid span {
      font-size: 0.76rem;
      font-weight: 700;
      order: 1;
      text-transform: uppercase;
    }
  }
</style>
