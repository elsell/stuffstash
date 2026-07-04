<script lang="ts" module>
  export type ChoiceGridOption = {
    value: string;
    label: string;
    description?: string;
    disabled?: boolean;
  };
</script>

<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    label,
    options,
    selectedValues,
    emptyMessage = 'No choices available.',
    onSelect
  }: {
    label: string;
    options: ChoiceGridOption[];
    selectedValues: string[];
    emptyMessage?: string;
    onSelect: (value: string) => void;
  } = $props();

  function isSelected(value: string): boolean {
    return selectedValues.includes(value);
  }
</script>

{#if options.length === 0}
  <p class="muted-note">{emptyMessage}</p>
{:else}
  <div class="choice-grid parent-picker option-grid" role="group" aria-label={label}>
    {#each options as option}
      <Button.Root
        type="button"
        variant={isSelected(option.value) ? 'secondary' : 'outline'}
        disabled={option.disabled}
        aria-pressed={isSelected(option.value)}
        onclick={() => onSelect(option.value)}
      >
        {#if option.description}
          <span>
            <strong>{option.label}</strong>
            <small>{option.description}</small>
          </span>
        {:else}
          {option.label}
        {/if}
      </Button.Root>
    {/each}
  </div>
{/if}
