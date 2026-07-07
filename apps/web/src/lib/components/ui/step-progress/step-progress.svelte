<script lang="ts" module>
  export type StepProgressStep = {
    id: string;
    label: string;
  };

  export type StepProgressState = 'complete' | 'current' | 'upcoming';
</script>

<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import * as Button from '$lib/components/ui/button/index.js';

  type Props = {
    steps: StepProgressStep[];
    current: string;
    reachableStepIds?: string[];
    ariaLabel?: string;
    onNavigateStep?: (stepId: string) => void;
  };

  let { steps, current, reachableStepIds = [], ariaLabel = 'Progress', onNavigateStep }: Props = $props();
  let currentIndex = $derived(Math.max(0, steps.findIndex((step) => step.id === current)));
  let reachableSet = $derived(new Set(reachableStepIds));

  function stepState(index: number): StepProgressState {
    if (index < currentIndex) return 'complete';
    if (index === currentIndex) return 'current';
    return 'upcoming';
  }

  function canNavigate(step: StepProgressStep): boolean {
    return Boolean(onNavigateStep && reachableSet.has(step.id));
  }
</script>

<ol class="step-progress" style:--step-count={Math.max(steps.length, 1)} aria-label={ariaLabel}>
  {#each steps as step, index}
    {@const state = stepState(index)}
    {@const navigable = canNavigate(step)}
    <li
      class={`step-progress-item ${state} ${navigable ? 'navigable' : 'locked'}`}
      data-state={state}
      data-reachable={navigable}
      aria-current={state === 'current' ? 'step' : undefined}
    >
      {#if navigable}
        <Button.Root
          type="button"
          variant="ghost"
          class="step-progress-control"
          aria-label={state === 'current' ? `${step.label}, current step` : step.label}
          onclick={() => onNavigateStep?.(step.id)}
        >
          <span class="step-progress-marker" aria-hidden="true">
            {#if state === 'complete'}
              <Check size={13} strokeWidth={3} />
            {:else if state === 'current'}
              <span class="step-progress-current-dot"></span>
            {/if}
          </span>
          <span class="step-progress-label">{step.label}</span>
        </Button.Root>
      {:else}
        <span class="step-progress-control locked-control" aria-disabled="true">
          <span class="step-progress-marker" aria-hidden="true">
            {#if state === 'complete'}
              <Check size={13} strokeWidth={3} />
            {:else if state === 'current'}
              <span class="step-progress-current-dot"></span>
            {/if}
          </span>
          <span class="step-progress-label">{step.label}</span>
        </span>
      {/if}
    </li>
  {/each}
</ol>

<style>
  .step-progress {
    align-items: start;
    color: var(--muted-foreground);
    display: grid;
    gap: 0;
    grid-template-columns: repeat(var(--step-count), minmax(0, 1fr));
    list-style: none;
    margin: 0 0 1rem;
    padding: 0;
  }

  .step-progress-item {
    display: grid;
    min-width: 0;
    position: relative;
  }

  .step-progress-item:not(:last-child)::after {
    background: var(--border);
    content: "";
    height: 2px;
    left: calc(50% + 0.675rem);
    position: absolute;
    right: calc(-50% + 0.675rem);
    top: 0.66rem;
  }

  .step-progress-item.complete:not(:last-child)::after {
    background: var(--primary);
  }

  :global(.step-progress-control),
  .step-progress-control {
    align-items: start;
    appearance: none;
    background: transparent;
    border: 0;
    color: inherit;
    display: grid;
    gap: 0.4rem;
    height: auto;
    justify-content: center;
    justify-items: center;
    min-width: 0;
    padding: 0;
    position: relative;
    text-align: center;
    width: 100%;
    z-index: 1;
  }

  :global(button.step-progress-control) {
    cursor: pointer;
    height: auto;
  }

  :global(button.step-progress-control:focus-visible) {
    border-radius: 8px;
    outline: 2px solid var(--ring);
    outline-offset: 0.35rem;
  }

  .step-progress-marker {
    background: var(--background);
    border: 2px solid var(--border);
    border-radius: 999px;
    box-shadow: 0 0 0 4px var(--step-progress-surface, var(--card));
    color: var(--primary-foreground);
    display: grid;
    height: 1.35rem;
    place-items: center;
    width: 1.35rem;
    z-index: 1;
  }

  .complete .step-progress-marker {
    background: var(--primary);
    border-color: var(--primary);
  }

  .current .step-progress-marker {
    border-color: var(--primary);
    box-shadow:
      0 0 0 4px var(--step-progress-surface, var(--card)),
      0 0 0 8px color-mix(in oklab, var(--primary) 14%, transparent);
  }

  .navigable:not(.current) :global(button.step-progress-control:hover .step-progress-marker) {
    box-shadow:
      0 0 0 4px var(--step-progress-surface, var(--card)),
      0 0 0 8px color-mix(in oklab, var(--primary) 12%, transparent);
  }

  .step-progress-current-dot {
    background: var(--primary);
    border-radius: 999px;
    display: block;
    height: 0.42rem;
    width: 0.42rem;
  }

  .step-progress-label {
    color: var(--muted-foreground);
    font-size: 0.78rem;
    font-weight: 650;
    line-height: 1.15;
    min-width: 0;
    overflow: hidden;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .complete .step-progress-label,
  .current .step-progress-label {
    color: var(--foreground);
  }

  @media (max-width: 520px) {
    .step-progress-label {
      font-size: 0.72rem;
    }
  }
</style>
