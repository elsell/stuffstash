<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';

  export interface SegmentedOption {
    value: string;
    label: string;
    disabled?: boolean;
  }

  let {
    label,
    value,
    options,
    onSelect
  }: {
    label: string;
    value: string;
    options: SegmentedOption[];
    onSelect: (value: string) => void;
  } = $props();
</script>

<div class="filter-control" role="group" aria-label={label}>
  {#each options as option}
    <Button.Root
      type="button"
      variant="ghost"
      disabled={option.disabled}
      aria-pressed={value === option.value}
      data-selected={value === option.value}
      onclick={() => onSelect(option.value)}
    >
      {option.label}
    </Button.Root>
  {/each}
</div>
