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
  import CustomFieldControls from './CustomFieldControls.svelte';

  let {
    activeCustomAssetTypes,
    applicableFields,
    customAssetTypeId,
    customFieldValues,
    onCustomAssetTypeSelect,
    onCustomFieldValueChange
  }: AddAssetCustomFieldsSectionProps = $props();

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

<CustomFieldControls
  fields={applicableFields}
  values={customFieldValues}
  idPrefix="custom-field"
  label="Custom fields"
  onValueChange={onCustomFieldValueChange}
/>
