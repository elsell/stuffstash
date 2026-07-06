<script lang="ts">
  type StepID = 'source' | 'connect' | 'preview' | 'run';

  type Step = {
    id: StepID;
    label: string;
  };

  type Props = {
    current: StepID;
  };

  const steps: Step[] = [
    { id: 'source', label: 'Source' },
    { id: 'connect', label: 'Connect' },
    { id: 'preview', label: 'Preview' },
    { id: 'run', label: 'Run' }
  ];

  let { current }: Props = $props();
  let currentIndex = $derived(steps.findIndex((step) => step.id === current));

  function stepState(index: number): 'complete' | 'current' | 'upcoming' {
    if (index < currentIndex) return 'complete';
    if (index === currentIndex) return 'current';
    return 'upcoming';
  }
</script>

<ol class="flow-stepper" aria-label="Import progress">
  {#each steps as step, index}
    {@const state = stepState(index)}
    <li class={state} aria-current={state === 'current' ? 'step' : undefined}>
      <span class="marker" aria-hidden="true"></span>
      <span class="label">{step.label}</span>
    </li>
  {/each}
</ol>

<style>
  .flow-stepper {
    align-items: start;
    color: hsl(var(--muted-foreground));
    display: grid;
    gap: 0;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    list-style: none;
    margin: 0 0 1rem;
    padding: 0;
  }

  .flow-stepper li {
    display: grid;
    gap: 0.4rem;
    min-width: 0;
    position: relative;
  }

  .flow-stepper li:not(:last-child)::after {
    background: hsl(var(--border));
    content: "";
    height: 2px;
    left: calc(0.75rem + 6px);
    position: absolute;
    right: 0.4rem;
    top: 0.34rem;
  }

  .flow-stepper li.complete:not(:last-child)::after {
    background: hsl(var(--primary));
  }

  .marker {
    background: hsl(var(--background));
    border: 2px solid hsl(var(--border));
    border-radius: 999px;
    display: block;
    height: 0.75rem;
    width: 0.75rem;
    z-index: 1;
  }

  .complete .marker,
  .current .marker {
    border-color: hsl(var(--primary));
  }

  .complete .marker {
    background: hsl(var(--primary));
  }

  .current .marker {
    background: hsl(var(--background));
    box-shadow: 0 0 0 4px hsl(var(--primary) / 0.12);
  }

  .label {
    font-size: 0.78rem;
    font-weight: 650;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .complete .label,
  .current .label {
    color: hsl(var(--foreground));
  }

  @media (max-width: 520px) {
    .label {
      font-size: 0.72rem;
    }
  }
</style>
