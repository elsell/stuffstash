<script lang="ts" module>
  export type StepProgressStep = {
    id: string;
    label: string;
    description?: string;
  };

  export type StepProgressState = 'complete' | 'current' | 'upcoming';
  export type StepProgressOrientation = 'horizontal' | 'vertical';
  export type StepProgressDensity = 'compact' | 'comfortable';
</script>

<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import * as Button from '$lib/components/ui/button/index.js';

  type Props = {
    steps: StepProgressStep[];
    current: string;
    reachableStepIds?: string[];
    ariaLabel?: string;
    orientation?: StepProgressOrientation;
    density?: StepProgressDensity;
    onNavigateStep?: (stepId: string) => void;
  };

  let {
    steps,
    current,
    reachableStepIds = [],
    ariaLabel = 'Progress',
    orientation = 'horizontal',
    density = 'compact',
    onNavigateStep
  }: Props = $props();
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

  function stateLabel(state: StepProgressState): string {
    if (state === 'complete') return 'Completed';
    if (state === 'current') return 'Current';
    return 'Not started';
  }

  function navigationLabel(step: StepProgressStep, state: StepProgressState): string {
    const prefix = state === 'current' ? `${step.label}, current step` : `Go to ${step.label}, ${stateLabel(state).toLowerCase()} step`;
    return step.description ? `${prefix}. ${step.description}` : prefix;
  }
</script>

<ol
  class="step-progress"
  style:--step-count={Math.max(steps.length, 1)}
  aria-label={ariaLabel}
  data-orientation={orientation}
  data-density={density}
>
  {#each steps as step, index}
    {@const state = stepState(index)}
    {@const navigable = canNavigate(step)}
    <li
      class={`step-progress-item ${state} ${navigable ? 'navigable' : 'locked'}`}
      data-state={state}
      data-reachable={navigable}
      data-step-id={step.id}
      aria-current={state === 'current' ? 'step' : undefined}
    >
      {#if navigable}
        <Button.Root
          type="button"
          variant="ghost"
          class="step-progress-control"
          aria-label={navigationLabel(step, state)}
          onclick={() => onNavigateStep?.(step.id)}
        >
          <span class="step-progress-marker" aria-hidden="true">
            {#if state === 'complete'}
              <Check size={13} strokeWidth={3} />
            {:else if state === 'current'}
              <span class="step-progress-current-dot"></span>
            {/if}
          </span>
          <span class="step-progress-copy">
            <span class="step-progress-label">{step.label}</span>
            {#if step.description}
              <span class="step-progress-description">{step.description}</span>
            {/if}
            <span class="step-progress-state">{stateLabel(state)}</span>
          </span>
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
          <span class="step-progress-copy">
            <span class="step-progress-label">{step.label}</span>
            {#if step.description}
              <span class="step-progress-description">{step.description}</span>
            {/if}
            <span class="step-progress-state">{stateLabel(state)}</span>
          </span>
        </span>
      {/if}
    </li>
  {/each}
</ol>

<style>
  .step-progress {
    --step-progress-marker-size: 1.35rem;
    --step-progress-marker-ring: 0.25rem;
    --step-progress-rail-size: 2px;
    --step-progress-rail-offset: calc(var(--step-progress-marker-size) / 2 - var(--step-progress-rail-size) / 2);
    align-items: start;
    color: var(--muted-foreground);
    display: grid;
    gap: 0;
    grid-template-columns: repeat(var(--step-count), minmax(0, 1fr));
    list-style: none;
    margin: 0 0 1rem;
    padding: 0;
  }

  .step-progress[data-density='comfortable'] {
    --step-progress-marker-size: 1.55rem;
    --step-progress-marker-ring: 0.3rem;
  }

  .step-progress[data-orientation='vertical'] {
    gap: 0.75rem;
    grid-template-columns: minmax(0, 1fr);
  }

  .step-progress-item {
    display: grid;
    min-width: 0;
    position: relative;
  }

  .step-progress[data-orientation='horizontal'] .step-progress-item:not(:last-child)::after {
    background: var(--border);
    content: "";
    height: var(--step-progress-rail-size);
    left: calc(50% + var(--step-progress-marker-size) / 2);
    position: absolute;
    right: calc(-50% + var(--step-progress-marker-size) / 2);
    top: var(--step-progress-rail-offset);
  }

  .step-progress[data-orientation='vertical'] .step-progress-item:not(:last-child)::after {
    background: var(--border);
    bottom: calc(-0.75rem - var(--step-progress-marker-ring));
    content: "";
    left: var(--step-progress-rail-offset);
    position: absolute;
    top: var(--step-progress-marker-size);
    width: var(--step-progress-rail-size);
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

  .step-progress[data-orientation='vertical'] :global(.step-progress-control),
  .step-progress[data-orientation='vertical'] .step-progress-control {
    align-items: start;
    gap: 0.65rem;
    grid-template-columns: var(--step-progress-marker-size) minmax(0, 1fr);
    justify-content: start;
    justify-items: start;
    text-align: left;
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
    box-shadow: 0 0 0 var(--step-progress-marker-ring) var(--step-progress-surface, var(--card));
    color: var(--primary-foreground);
    display: grid;
    height: var(--step-progress-marker-size);
    place-items: center;
    width: var(--step-progress-marker-size);
    z-index: 1;
  }

  .complete .step-progress-marker {
    background: var(--primary);
    border-color: var(--primary);
  }

  .current .step-progress-marker {
    border-color: var(--primary);
    box-shadow:
      0 0 0 var(--step-progress-marker-ring) var(--step-progress-surface, var(--card)),
      0 0 0 calc(var(--step-progress-marker-ring) * 2) color-mix(in oklab, var(--primary) 14%, transparent);
  }

  .navigable:not(.current) :global(button.step-progress-control:hover .step-progress-marker) {
    box-shadow:
      0 0 0 var(--step-progress-marker-ring) var(--step-progress-surface, var(--card)),
      0 0 0 calc(var(--step-progress-marker-ring) * 2) color-mix(in oklab, var(--primary) 12%, transparent);
  }

  .step-progress-current-dot {
    background: var(--primary);
    border-radius: 999px;
    display: block;
    height: 0.42rem;
    width: 0.42rem;
  }

  .step-progress-copy {
    display: grid;
    gap: 0.18rem;
    min-width: 0;
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

  .step-progress-description {
    color: var(--muted-foreground);
    display: none;
    font-size: 0.75rem;
    line-height: 1.2;
  }

  .step-progress[data-orientation='vertical'] .step-progress-description {
    display: block;
  }

  .step-progress[data-orientation='vertical'] .step-progress-label,
  .step-progress[data-orientation='vertical'] .step-progress-description {
    overflow: visible;
    text-align: left;
    text-overflow: clip;
    white-space: normal;
  }

  .step-progress-state {
    border: 0;
    clip: rect(0 0 0 0);
    height: 1px;
    margin: -1px;
    overflow: hidden;
    padding: 0;
    position: absolute;
    white-space: nowrap;
    width: 1px;
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
