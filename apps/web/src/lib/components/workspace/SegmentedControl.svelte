<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';

  export interface SegmentedOption {
    value: string;
    label: string;
    description?: string;
    disabled?: boolean;
  }

  let {
    label,
    value,
    options,
    layout = 'inline',
    onSelect
  }: {
    label: string;
    value: string;
    options: SegmentedOption[];
    layout?: 'inline' | 'section-rail';
    onSelect: (value: string) => void;
  } = $props();
</script>

<div class="filter-control" data-layout={layout} role="group" aria-label={label}>
  {#each options as option}
    <Button.Root
      type="button"
      variant="ghost"
      disabled={option.disabled}
      aria-pressed={value === option.value}
      data-selected={value === option.value}
      onclick={() => onSelect(option.value)}
    >
      {#if option.description}
        <span>{option.label}</span>
        <small>{option.description}</small>
      {:else}
        {option.label}
      {/if}
    </Button.Root>
  {/each}
</div>
