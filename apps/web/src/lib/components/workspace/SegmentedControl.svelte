<script lang="ts">
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
    if (option.href && !shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    onSelect(option.value);
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey && !event.defaultPrevented;
  }
</script>

<div class="filter-control" data-layout={layout} role="group" aria-label={label}>
  {#each options as option}
    {#if option.href && !option.disabled}
      <Button.Root
        href={option.href}
        variant="ghost"
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
