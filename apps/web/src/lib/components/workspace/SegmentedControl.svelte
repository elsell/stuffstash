<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import * as Button from '$lib/components/ui/button/index.js';

  export interface SegmentedOption {
    value: string;
    label: string;
    description?: string;
    disabled?: boolean;
    href?: string;
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

  function selectOption(event: MouseEvent, option: SegmentedOption): void {
    if (option.disabled) {
      return;
    }
    if (option.href && !shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onSelect(option.value);
  }
</script>

<div class="filter-control" data-layout={layout} role="group" aria-label={label}>
  {#each options as option}
    {#if option.href}
      <Button.Root
        href={option.href}
        variant="ghost"
        class="h-11 min-w-11"
        disabled={option.disabled}
        aria-current={value === option.value ? 'page' : undefined}
        data-selected={value === option.value}
        onclick={(event) => selectOption(event, option)}
      >
        {#if option.description}
          <span>{option.label}</span>
          <small>{option.description}</small>
        {:else}
          {option.label}
        {/if}
      </Button.Root>
    {:else}
      <Button.Root
        type="button"
        variant="ghost"
        class="h-11 min-w-11"
        disabled={option.disabled}
        aria-pressed={value === option.value}
        data-selected={value === option.value}
        onclick={(event) => selectOption(event, option)}
      >
        {#if option.description}
          <span>{option.label}</span>
          <small>{option.description}</small>
        {:else}
          {option.label}
        {/if}
      </Button.Root>
    {/if}
  {/each}
</div>
