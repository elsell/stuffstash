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
  import ChoiceGrid, { type ChoiceGridOption } from './ChoiceGrid.svelte';
  import CustomFieldControls from './CustomFieldControls.svelte';

  let {
    activeCustomAssetTypes,
    applicableFields,
    customAssetTypeId,
    customFieldValues,
    onCustomAssetTypeSelect,
    onCustomFieldValueChange
  }: AddAssetCustomFieldsSectionProps = $props();

  let customTypeOptions = $derived<ChoiceGridOption[]>([
    { value: '', label: 'Base asset' },
    ...activeCustomAssetTypes.map((assetType) => ({
      value: assetType.id,
      label: assetType.displayName,
      description: assetType.description || undefined
    }))
  ]);
</script>

{#if activeCustomAssetTypes.length > 0}
  <div class="field-stack">
    <fieldset class="selection-field">
      <legend>Custom type</legend>
      <ChoiceGrid
        label="Custom asset type"
        options={customTypeOptions}
        selectedValues={[customAssetTypeId]}
        onSelect={onCustomAssetTypeSelect}
      />
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
