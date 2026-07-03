<script lang="ts" module>
  import type { CustomFieldDefinition } from '$lib/domain/inventory';

  export type CustomFieldControlsProps = {
    fields: CustomFieldDefinition[];
    values: Record<string, string>;
    idPrefix: string;
    label: string;
    onValueChange: (key: string, value: string) => void;
  };
</script>

<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    fields,
    values,
    idPrefix,
    label,
    onValueChange
  }: CustomFieldControlsProps = $props();

  const booleanOptions = [
    { value: '', label: 'Unset' },
    { value: 'true', label: 'Yes' },
    { value: 'false', label: 'No' }
  ];

  function inputType(field: CustomFieldDefinition): string {
    if (field.type === 'number') return 'number';
    if (field.type === 'date') return 'date';
    if (field.type === 'url') return 'url';
    return 'text';
  }
</script>

{#if fields.length > 0}
  <div class="custom-field-grid" role="group" aria-label={label}>
    {#each fields as field}
      {#if field.type === 'boolean'}
        <fieldset class="selection-field">
          <legend>{field.displayName}</legend>
          <SegmentedControl
            label={field.displayName}
            value={values[field.key] ?? ''}
            options={booleanOptions}
            onSelect={(value) => onValueChange(field.key, value)}
          />
        </fieldset>
      {:else if field.type === 'enum'}
        <fieldset class="selection-field">
          <legend>{field.displayName}</legend>
          <div class="parent-picker option-grid" role="group" aria-label={field.displayName}>
            <Button.Root
              type="button"
              variant={(values[field.key] ?? '') === '' ? 'secondary' : 'outline'}
              aria-pressed={(values[field.key] ?? '') === ''}
              onclick={() => onValueChange(field.key, '')}
            >
              Unset
            </Button.Root>
            {#each field.enumOptions as option}
              <Button.Root
                type="button"
                variant={values[field.key] === option ? 'secondary' : 'outline'}
                aria-pressed={values[field.key] === option}
                onclick={() => onValueChange(field.key, option)}
              >
                {option}
              </Button.Root>
            {/each}
          </div>
        </fieldset>
      {:else}
        <div class="field-stack">
          <Label for={`${idPrefix}-${field.key}`}>{field.displayName}</Label>
          <Input
            id={`${idPrefix}-${field.key}`}
            type={inputType(field)}
            value={values[field.key] ?? ''}
            oninput={(event) => onValueChange(field.key, event.currentTarget.value)}
          />
        </div>
      {/if}
    {/each}
  </div>
{/if}
