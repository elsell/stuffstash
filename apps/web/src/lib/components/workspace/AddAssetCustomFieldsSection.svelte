<script lang="ts" module>
  import { t } from '$lib/presentation/localization';
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
    { value: '', label: t('web.AddAssetCustomFieldsSection.baseAsset') },
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
      <legend>{t('web.AddAssetCustomFieldsSection.customType')}</legend>
      <ChoiceGrid
        label={t('web.AddAssetCustomFieldsSection.customAssetType')}
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
  label={t('web.AddAssetCustomFieldsSection.customFields')}
  onValueChange={onCustomFieldValueChange}
/>
