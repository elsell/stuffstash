<script lang="ts">
  import Shapes from '@lucide/svelte/icons/shapes';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type {
    CustomAssetType,
    CustomFieldApplicability,
    CustomFieldDefinition,
    CustomFieldType,
    Inventory,
    Tenant
  } from '$lib/domain/inventory';
  import { hasAccessPermission } from '$lib/domain/inventory';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenant,
    inventory,
    repository,
    initialAssetTypes,
    initialFieldDefinitions,
    onSchemaChange
  }: {
    tenant: Tenant | null;
    inventory: Inventory | null;
    repository: InventoryCustomizationRepository;
    initialAssetTypes: CustomAssetType[];
    initialFieldDefinitions: CustomFieldDefinition[];
    onSchemaChange: (assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]) => void;
  } = $props();

  let assetTypes = $state<CustomAssetType[]>([]);
  let fieldDefinitions = $state<CustomFieldDefinition[]>([]);
  let busy = $state(false);
  let error = $state('');
  let typeScope = $state<'tenant' | 'inventory'>('inventory');
  let typeKey = $state('');
  let typeName = $state('');
  let typeDescription = $state('');
  let fieldScope = $state<'tenant' | 'inventory'>('inventory');
  let fieldKey = $state('');
  let fieldName = $state('');
  let fieldType = $state<CustomFieldType>('text');
  let fieldApplicability = $state<CustomFieldApplicability>('all_assets');
  let fieldTargets = $state<string[]>([]);
  let enumOptions = $state('');
  const scopeOptions = [
    { value: 'inventory', label: 'Inventory', disabled: false },
    { value: 'tenant', label: 'Tenant', disabled: false }
  ];
  const fieldTypeOptions = ['text', 'number', 'boolean', 'date', 'url', 'enum'].map((option) => ({ value: option, label: option }));
  const applicabilityOptions = [
    { value: 'all_assets', label: 'All assets' },
    { value: 'custom_asset_types', label: 'Types only' }
  ];

  let canConfigureInventory = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canManage = $derived(canConfigureInventory || canConfigureTenant);
  let activeAssetTypes = $derived(assetTypes.filter((assetType) => assetType.lifecycleState === 'active'));
  let activeFieldDefinitions = $derived(fieldDefinitions.filter((definition) => definition.lifecycleState === 'active'));
  let targetableAssetTypes = $derived(activeAssetTypes.filter((assetType) => fieldScope === 'tenant' ? assetType.scope === 'tenant' : true));
  let selectedTargetCount = $derived(fieldTargets.filter((id) => targetableAssetTypes.some((assetType) => assetType.id === id)).length);

  $effect(() => {
    assetTypes = initialAssetTypes;
    fieldDefinitions = initialFieldDefinitions;
  });

  async function createAssetType(): Promise<void> {
    if (!tenant || !inventory || !typeKey.trim() || !typeName.trim() || !canScope(typeScope)) {
      return;
    }
    busy = true;
    error = '';
    try {
      const created = await repository.createCustomAssetType(tenant.id, inventory.id, {
        scope: typeScope,
        key: typeKey.trim(),
        displayName: typeName.trim(),
        description: typeDescription.trim()
      });
      const nextAssetTypes = [created, ...assetTypes];
      assetTypes = nextAssetTypes;
      onSchemaChange(nextAssetTypes, fieldDefinitions);
      typeKey = '';
      typeName = '';
      typeDescription = '';
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Unable to create custom asset type.';
    } finally {
      busy = false;
    }
  }

  async function createFieldDefinition(): Promise<void> {
    if (!tenant || !inventory || !fieldKey.trim() || !fieldName.trim() || !canScope(fieldScope)) {
      return;
    }
    if (fieldApplicability === 'custom_asset_types' && fieldTargets.length === 0) {
      error = 'Select at least one custom type for this field.';
      return;
    }
    busy = true;
    error = '';
    try {
      const created = await repository.createCustomFieldDefinition(tenant.id, inventory.id, {
        scope: fieldScope,
        key: fieldKey.trim(),
        displayName: fieldName.trim(),
        type: fieldType,
        enumOptions: fieldType === 'enum' ? splitOptions(enumOptions) : [],
        applicability: fieldApplicability,
        customAssetTypeIds: fieldApplicability === 'custom_asset_types' ? fieldTargets : []
      });
      const nextFieldDefinitions = [created, ...fieldDefinitions];
      fieldDefinitions = nextFieldDefinitions;
      onSchemaChange(assetTypes, nextFieldDefinitions);
      fieldKey = '';
      fieldName = '';
      fieldType = 'text';
      fieldApplicability = 'all_assets';
      fieldTargets = [];
      enumOptions = '';
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Unable to create custom field.';
    } finally {
      busy = false;
    }
  }

  async function archiveAssetType(assetType: CustomAssetType): Promise<void> {
    if (!tenant || !inventory || !canScope(assetType.scope)) return;
    busy = true;
    error = '';
    try {
      const archived = await repository.archiveCustomAssetType(tenant.id, inventory.id, assetType.id, assetType.scope);
      const nextAssetTypes = assetTypes.map((candidate) => candidate.id === archived.id ? archived : candidate);
      const nextFieldTargets = fieldTargets.filter((id) => id !== assetType.id);
      assetTypes = nextAssetTypes;
      fieldTargets = nextFieldTargets;
      onSchemaChange(nextAssetTypes, fieldDefinitions);
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Unable to archive custom asset type.';
    } finally {
      busy = false;
    }
  }

  async function archiveFieldDefinition(definition: CustomFieldDefinition): Promise<void> {
    if (!tenant || !inventory || !canScope(definition.scope)) return;
    busy = true;
    error = '';
    try {
      const archived = await repository.archiveCustomFieldDefinition(tenant.id, inventory.id, definition.id, definition.scope);
      const nextFieldDefinitions = fieldDefinitions.map((candidate) => candidate.id === archived.id ? archived : candidate);
      fieldDefinitions = nextFieldDefinitions;
      onSchemaChange(assetTypes, nextFieldDefinitions);
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Unable to archive custom field.';
    } finally {
      busy = false;
    }
  }

  function toggleTarget(assetTypeId: string): void {
    fieldTargets = fieldTargets.includes(assetTypeId)
      ? fieldTargets.filter((candidate) => candidate !== assetTypeId)
      : [...fieldTargets, assetTypeId];
  }

  function selectFieldScope(scope: 'tenant' | 'inventory'): void {
    fieldScope = scope;
    fieldTargets = fieldTargets.filter((id) =>
      assetTypes.some(
        (assetType) =>
          assetType.id === id &&
          assetType.lifecycleState === 'active' &&
          (scope === 'inventory' || assetType.scope === 'tenant')
      )
    );
  }

  function canScope(scope: 'tenant' | 'inventory'): boolean {
    return scope === 'tenant' ? canConfigureTenant : canConfigureInventory;
  }

  function splitOptions(value: string): string[] {
    return value.split(',').map((option) => option.trim()).filter(Boolean);
  }
</script>

<section class="settings-panel wide" aria-labelledby="settings-customization">
  <div class="settings-panel-heading">
    <Shapes aria-hidden="true" />
    <div>
      <h2 id="settings-customization">Custom fields</h2>
      <p>Types and fields available to this inventory.</p>
    </div>
  </div>

  {#if !tenant || !inventory}
    <p class="denied-note">Select an inventory before managing fields.</p>
  {:else if !canManage}
    <p class="denied-note">Custom fields require tenant or inventory configuration access.</p>
  {:else}
    <div class="customization-grid">
      <div class="customization-column">
        <h3>Asset types</h3>
        <SegmentedControl
          label="Custom type scope"
          value={typeScope}
          options={scopeOptions.map((option) => ({
            ...option,
            disabled: option.value === 'inventory' ? !canConfigureInventory : !canConfigureTenant
          }))}
          onSelect={(value) => { typeScope = value as 'tenant' | 'inventory'; }}
        />
        <div class="field-stack">
          <Label for="custom-type-key">Key</Label>
          <Input id="custom-type-key" bind:value={typeKey} placeholder="medicine" />
        </div>
        <div class="field-stack">
          <Label for="custom-type-name">Display name</Label>
          <Input id="custom-type-name" bind:value={typeName} placeholder="Medicine" />
        </div>
        <div class="field-stack">
          <Label for="custom-type-description">Description</Label>
          <Textarea id="custom-type-description" bind:value={typeDescription} placeholder="Optional" />
        </div>
        <Button.Root disabled={busy || !typeKey.trim() || !typeName.trim()} onclick={() => { void createAssetType(); }}>Create type</Button.Root>

        <div class="schema-list" aria-label="Custom asset types">
          {#each activeAssetTypes as assetType}
            <article class="schema-row">
              <div>
                <strong>{assetType.displayName}</strong>
                <small>{assetType.key}</small>
              </div>
              <div class="audit-meta">
                <Badge variant="outline">{assetType.scope}</Badge>
                <Button.Root variant="ghost" size="icon-xs" aria-label={`Archive ${assetType.displayName}`} disabled={busy || !canScope(assetType.scope)} onclick={() => { void archiveAssetType(assetType); }}>
                  <Trash2 />
                </Button.Root>
              </div>
            </article>
          {/each}
        </div>
      </div>

      <div class="customization-column">
        <h3>Field definitions</h3>
        <SegmentedControl
          label="Custom field scope"
          value={fieldScope}
          options={scopeOptions.map((option) => ({
            ...option,
            disabled: option.value === 'inventory' ? !canConfigureInventory : !canConfigureTenant
          }))}
          onSelect={(value) => selectFieldScope(value as 'tenant' | 'inventory')}
        />
        <div class="field-stack">
          <Label for="custom-field-key">Key</Label>
          <Input id="custom-field-key" bind:value={fieldKey} placeholder="expiration-date" />
        </div>
        <div class="field-stack">
          <Label for="custom-field-name">Display name</Label>
          <Input id="custom-field-name" bind:value={fieldName} placeholder="Expiration date" />
        </div>
        <SegmentedControl
          label="Custom field type"
          value={fieldType}
          options={fieldTypeOptions}
          onSelect={(value) => { fieldType = value as CustomFieldType; }}
        />
        {#if fieldType === 'enum'}
          <div class="field-stack">
            <Label for="custom-field-options">Options</Label>
            <Input id="custom-field-options" bind:value={enumOptions} placeholder="new, open, closed" />
          </div>
        {/if}
        <SegmentedControl
          label="Field applicability"
          value={fieldApplicability}
          options={applicabilityOptions}
          onSelect={(value) => { fieldApplicability = value as CustomFieldApplicability; }}
        />
        {#if fieldApplicability === 'custom_asset_types'}
          <fieldset class="selection-field">
            <legend>Field custom type targets</legend>
            <p class="selection-summary">
              {selectedTargetCount === 0
                ? 'No custom types selected'
                : `${selectedTargetCount} custom ${selectedTargetCount === 1 ? 'type' : 'types'} selected`}
            </p>
            {#if targetableAssetTypes.length === 0}
              <p class="muted-note">No eligible custom asset types for this scope.</p>
            {:else}
              <div class="parent-picker option-grid" role="group" aria-label="Field custom type targets">
                {#each targetableAssetTypes as assetType}
                  <Button.Root
                    variant={fieldTargets.includes(assetType.id) ? 'secondary' : 'outline'}
                    aria-pressed={fieldTargets.includes(assetType.id)}
                    onclick={() => toggleTarget(assetType.id)}
                  >
                    {assetType.displayName}
                  </Button.Root>
                {/each}
              </div>
            {/if}
          </fieldset>
        {/if}
        <Button.Root disabled={busy || !fieldKey.trim() || !fieldName.trim()} onclick={() => { void createFieldDefinition(); }}>Create field</Button.Root>

        <div class="schema-list" aria-label="Custom field definitions">
          {#each activeFieldDefinitions as definition}
            <article class="schema-row">
              <div>
                <strong>{definition.displayName}</strong>
                <small>{definition.key} / {definition.type}</small>
              </div>
              <div class="audit-meta">
                <Badge variant="outline">{definition.scope}</Badge>
                <Button.Root variant="ghost" size="icon-xs" aria-label={`Archive ${definition.displayName}`} disabled={busy || !canScope(definition.scope)} onclick={() => { void archiveFieldDefinition(definition); }}>
                  <Trash2 />
                </Button.Root>
              </div>
            </article>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  {#if error}
    <p class="denied-note" role="alert">{error}</p>
  {/if}
</section>
