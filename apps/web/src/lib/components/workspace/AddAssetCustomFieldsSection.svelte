<script lang="ts" module>
  import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';

  export type AddAssetCustomFieldsSectionProps = {
    activeCustomAssetTypes: CustomAssetType[];
    applicableFields: CustomFieldDefinition[];
    customAssetTypeId: string;
    customFieldValues: Record<string, string>;
    onCustomAssetTypeSelect: (id: string) => void;
    onCustomFieldValueChange: (key: string, value: string) => void;
  };
</script>

<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    activeCustomAssetTypes,
    applicableFields,
    customAssetTypeId,
    customFieldValues,
    onCustomAssetTypeSelect,
    onCustomFieldValueChange
  }: AddAssetCustomFieldsSectionProps = $props();

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

{#if activeCustomAssetTypes.length > 0}
  <div class="field-stack">
    <fieldset class="selection-field">
      <legend>Custom type</legend>
      <div class="parent-picker option-grid" role="group" aria-label="Custom asset type">
        <Button.Root
          type="button"
          variant={customAssetTypeId === '' ? 'secondary' : 'outline'}
          aria-pressed={customAssetTypeId === ''}
          onclick={() => onCustomAssetTypeSelect('')}
        >
          Base asset
        </Button.Root>
        {#each activeCustomAssetTypes as assetType}
          <Button.Root
            type="button"
            variant={customAssetTypeId === assetType.id ? 'secondary' : 'outline'}
            aria-pressed={customAssetTypeId === assetType.id}
            onclick={() => onCustomAssetTypeSelect(assetType.id)}
          >
            {assetType.displayName}
          </Button.Root>
        {/each}
      </div>
    </fieldset>
  </div>
{/if}

{#if applicableFields.length > 0}
  <div class="custom-field-grid" aria-label="Custom fields">
    {#each applicableFields as field}
      {#if field.type === 'boolean'}
        <fieldset class="selection-field">
          <legend>{field.displayName}</legend>
          <SegmentedControl
            label={field.displayName}
            value={customFieldValues[field.key] ?? ''}
            options={booleanOptions}
            onSelect={(value) => onCustomFieldValueChange(field.key, value)}
          />
        </fieldset>
      {:else if field.type === 'enum'}
        <fieldset class="selection-field">
          <legend>{field.displayName}</legend>
          <div class="parent-picker option-grid" role="group" aria-label={field.displayName}>
            <Button.Root
              type="button"
              variant={(customFieldValues[field.key] ?? '') === '' ? 'secondary' : 'outline'}
              aria-pressed={(customFieldValues[field.key] ?? '') === ''}
              onclick={() => onCustomFieldValueChange(field.key, '')}
            >
              Unset
            </Button.Root>
            {#each field.enumOptions as option}
              <Button.Root
                type="button"
                variant={customFieldValues[field.key] === option ? 'secondary' : 'outline'}
                aria-pressed={customFieldValues[field.key] === option}
                onclick={() => onCustomFieldValueChange(field.key, option)}
              >
                {option}
              </Button.Root>
            {/each}
          </div>
        </fieldset>
      {:else}
        <div class="field-stack">
          <Label for={`custom-field-${field.key}`}>{field.displayName}</Label>
          <Input
            id={`custom-field-${field.key}`}
            type={inputType(field)}
            value={customFieldValues[field.key] ?? ''}
            oninput={(event) => onCustomFieldValueChange(field.key, event.currentTarget.value)}
          />
        </div>
      {/if}
    {/each}
  </div>
{/if}
