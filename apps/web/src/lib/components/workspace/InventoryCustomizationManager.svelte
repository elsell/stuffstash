<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { tick } from 'svelte';
  import Shapes from '@lucide/svelte/icons/shapes';
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import {
    customizationArchiveAssetTypeHref,
    customizationArchiveFieldDefinitionHref,
    customizationFieldsHref,
    customizationManagerAccessStatus,
    customizationManagerOperationStatus
  } from '$lib/application/workspaceCustomizationActions';
  import {
    customizationApplicabilityOptions,
    customizationFieldTypeOptions,
    customizationScopeOptions,
    customizationTargetAssetTypeOptions
  } from '$lib/application/workspaceCustomizationPresentation';
  import type { CustomizationRouteAction } from '$lib/application/workspaceRoute';
  import type {
    CustomAssetType,
    CustomDefinitionScope,
    CustomFieldApplicability,
    CustomFieldDefinition,
    CustomFieldType,
    Inventory,
    Tenant
  } from '$lib/domain/inventory';
  import { hasAccessPermission } from '$lib/domain/inventory';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import ChoiceGrid from './ChoiceGrid.svelte';
  import InventoryCustomizationArchivePanel from './InventoryCustomizationArchivePanel.svelte';
  import SegmentedControl from './SegmentedControl.svelte';

  let {
    tenant,
    inventory,
    repository,
    initialAssetTypes,
    initialFieldDefinitions,
    archiveAction = null,
    archiveAssetTypeId = null,
    archiveFieldDefinitionId = null,
    onArchiveActionOpen = () => {},
    onArchiveActionClose = () => {},
    onSchemaChange
  }: {
    tenant: Tenant | null;
    inventory: Inventory | null;
    repository: InventoryCustomizationRepository;
    initialAssetTypes: CustomAssetType[];
    initialFieldDefinitions: CustomFieldDefinition[];
    archiveAction?: CustomizationRouteAction;
    archiveAssetTypeId?: string | null;
    archiveFieldDefinitionId?: string | null;
    onArchiveActionOpen?: (action: CustomizationRouteAction, id: string) => void;
    onArchiveActionClose?: () => void;
    onSchemaChange: (assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]) => void;
  } = $props();

  let assetTypes = $state<CustomAssetType[]>([]);
  let fieldDefinitions = $state<CustomFieldDefinition[]>([]);
  let busy = $state(false);
  let error = $state('');
  let typeScope = $state<CustomDefinitionScope>('inventory');
  let typeKey = $state('');
  let typeName = $state('');
  let typeDescription = $state('');
  let fieldScope = $state<CustomDefinitionScope>('inventory');
  let fieldKey = $state('');
  let fieldName = $state('');
  let fieldType = $state<CustomFieldType>('text');
  let fieldApplicability = $state<CustomFieldApplicability>('all_assets');
  let fieldTargets = $state<string[]>([]);
  let enumOptions = $state('');
  let archiveConfirmationElement = $state<HTMLElement | null>(null);
  let lastArchiveRouteKey = $state('');
  const fieldTypeOptions = customizationFieldTypeOptions();
  const applicabilityOptions = customizationApplicabilityOptions();

  let canConfigureInventory = $derived(hasAccessPermission(inventory?.access, 'configure'));
  let canConfigureTenant = $derived(hasAccessPermission(tenant?.access, 'configure'));
  let canManage = $derived(canConfigureInventory || canConfigureTenant);
  let accessStatus = $derived(customizationManagerAccessStatus({ hasTenant: Boolean(tenant), hasInventory: Boolean(inventory), canManage }));
  let operationStatus = $derived(customizationManagerOperationStatus(error));
  let activeAssetTypes = $derived(assetTypes.filter((assetType) => assetType.lifecycleState === 'active'));
  let activeFieldDefinitions = $derived(fieldDefinitions.filter((definition) => definition.lifecycleState === 'active'));
  let scopeOptions = $derived(customizationScopeOptions({ canConfigureInventory, canConfigureTenant }));
  let firstAvailableScope = $derived(firstAllowedScope(canConfigureInventory, canConfigureTenant));
  let targetableAssetTypeOptions = $derived(customizationTargetAssetTypeOptions({ assetTypes, fieldScope }));
  let selectedTargetCount = $derived(fieldTargets.filter((id) => targetableAssetTypeOptions.some((option) => option.value === id)).length);
  let routeArchiveAssetType = $derived(
    archiveAction === 'archive_asset_type'
      ? activeAssetTypes.find((assetType) => assetType.id === archiveAssetTypeId) ?? null
      : null
  );
  let routeArchiveFieldDefinition = $derived(
    archiveAction === 'archive_field_definition'
      ? activeFieldDefinitions.find((definition) => definition.id === archiveFieldDefinitionId) ?? null
      : null
  );
  let hasArchiveRoute = $derived(archiveAction === 'archive_asset_type' || archiveAction === 'archive_field_definition');
  let archiveRouteKey = $derived(
    archiveAction === 'archive_asset_type'
      ? `${archiveAction}:${archiveAssetTypeId ?? ''}`
      : archiveAction === 'archive_field_definition'
        ? `${archiveAction}:${archiveFieldDefinitionId ?? ''}`
        : ''
  );

  $effect(() => {
    assetTypes = initialAssetTypes;
    fieldDefinitions = initialFieldDefinitions;
  });

  $effect(() => {
    const fallbackScope = firstAvailableScope;
    if (!fallbackScope) {
      return;
    }
    if (!canScope(typeScope)) {
      typeScope = fallbackScope;
    }
    if (!canScope(fieldScope)) {
      selectFieldScope(fallbackScope);
    }
  });

  $effect(() => {
    const routeKey = archiveRouteKey;
    if (!routeKey) {
      lastArchiveRouteKey = '';
      return;
    }
    if (routeKey === lastArchiveRouteKey) {
      return;
    }
    lastArchiveRouteKey = routeKey;
    void tick().then(() => archiveConfirmationElement?.focus());
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
      onArchiveActionClose();
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
      onArchiveActionClose();
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

  function selectFieldScope(scope: CustomDefinitionScope): void {
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

  function canScope(scope: CustomDefinitionScope): boolean {
    return scope === 'tenant' ? canConfigureTenant : canConfigureInventory;
  }

  function firstAllowedScope(canUseInventoryScope: boolean, canUseTenantScope: boolean): CustomDefinitionScope | null {
    if (canUseInventoryScope) return 'inventory';
    if (canUseTenantScope) return 'tenant';
    return null;
  }

  function splitOptions(value: string): string[] {
    return value.split(',').map((option) => option.trim()).filter(Boolean);
  }

  function fieldsHref(): string {
    return customizationFieldsHref(tenant?.id ?? inventory?.tenantId ?? null, inventory?.id ?? null);
  }

  function archiveAssetTypeHref(assetType: CustomAssetType): string {
    return customizationArchiveAssetTypeHref(tenant?.id ?? inventory?.tenantId ?? null, inventory?.id ?? null, assetType);
  }

  function archiveFieldDefinitionHref(definition: CustomFieldDefinition): string {
    return customizationArchiveFieldDefinitionHref(
      tenant?.id ?? inventory?.tenantId ?? null,
      inventory?.id ?? null,
      definition
    );
  }

  function openArchiveAction(event: MouseEvent, action: Exclude<CustomizationRouteAction, null>, id: string): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onArchiveActionOpen(action, id);
  }

  function closeArchiveAction(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
    onArchiveActionClose();
  }
</script>

<section class="settings-panel wide customization-panel" aria-labelledby="settings-customization">
  <div class="settings-panel-heading">
    <Shapes aria-hidden="true" />
    <div>
      <h2 id="settings-customization">Custom fields</h2>
      <p>Types and fields available to this inventory.</p>
    </div>
  </div>

  {#if accessStatus}
    <p class="denied-note" role={accessStatus.alert ? 'alert' : undefined}>{accessStatus.message}</p>
  {:else}
    {#if hasArchiveRoute}
      <InventoryCustomizationArchivePanel
        assetType={routeArchiveAssetType}
        fieldDefinition={routeArchiveFieldDefinition}
        {busy}
        fieldsHref={fieldsHref()}
        bind:panelElement={archiveConfirmationElement}
        canArchiveScope={canScope}
        onClose={closeArchiveAction}
        onArchiveAssetType={archiveAssetType}
        onArchiveFieldDefinition={archiveFieldDefinition}
      />
    {/if}

    <div class="customization-grid">
      <section class="customization-column customization-surface" aria-labelledby="custom-asset-types-title">
        <div class="customization-surface-heading">
          <h3 id="custom-asset-types-title">Asset types</h3>
          <span aria-label={`${activeAssetTypes.length} custom asset types`}>{activeAssetTypes.length}</span>
        </div>
        <SegmentedControl
          label="Custom type scope"
          value={typeScope}
          options={scopeOptions}
          onSelect={(value) => { typeScope = value as CustomDefinitionScope; }}
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
        <Button.Root disabled={busy || !typeKey.trim() || !typeName.trim() || !canScope(typeScope)} onclick={() => { void createAssetType(); }}>Create type</Button.Root>

        <div class="schema-list" aria-label="Custom asset types">
          {#if activeAssetTypes.length === 0}
            <p class="schema-empty">No custom asset types yet.</p>
          {:else}
          {#each activeAssetTypes as assetType}
            <article class="schema-row">
              <div>
                <strong>{assetType.displayName}</strong>
                <small>{assetType.key}</small>
              </div>
              <div class="audit-meta">
                <Badge variant="outline">{assetType.scope}</Badge>
                <Button.Root
                  href={archiveAssetTypeHref(assetType)}
                  variant="ghost"
                  size="icon-xs"
                  aria-label={`Archive ${assetType.displayName}`}
                  disabled={busy || !canScope(assetType.scope)}
                  onclick={(event) => openArchiveAction(event, 'archive_asset_type', assetType.id)}
                >
                  <Trash2 />
                </Button.Root>
              </div>
            </article>
          {/each}
          {/if}
        </div>
      </section>

      <section class="customization-column customization-surface" aria-labelledby="custom-field-definitions-title">
        <div class="customization-surface-heading">
          <h3 id="custom-field-definitions-title">Field definitions</h3>
          <span aria-label={`${activeFieldDefinitions.length} custom fields`}>{activeFieldDefinitions.length}</span>
        </div>
        <SegmentedControl
          label="Custom field scope"
          value={fieldScope}
          options={scopeOptions}
          onSelect={(value) => selectFieldScope(value as CustomDefinitionScope)}
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
            <ChoiceGrid
              label="Field custom type targets"
              options={targetableAssetTypeOptions}
              selectedValues={fieldTargets}
              emptyMessage="No eligible custom asset types for this scope."
              onSelect={toggleTarget}
            />
          </fieldset>
        {/if}
        <Button.Root disabled={busy || !fieldKey.trim() || !fieldName.trim() || !canScope(fieldScope)} onclick={() => { void createFieldDefinition(); }}>Create field</Button.Root>

        <div class="schema-list" aria-label="Custom field definitions">
          {#if activeFieldDefinitions.length === 0}
            <p class="schema-empty">No custom fields yet.</p>
          {:else}
          {#each activeFieldDefinitions as definition}
            <article class="schema-row">
              <div>
                <strong>{definition.displayName}</strong>
                <small>{definition.key} / {definition.type}</small>
              </div>
              <div class="audit-meta">
                <Badge variant="outline">{definition.scope}</Badge>
                <Button.Root
                  href={archiveFieldDefinitionHref(definition)}
                  variant="ghost"
                  size="icon-xs"
                  aria-label={`Archive ${definition.displayName}`}
                  disabled={busy || !canScope(definition.scope)}
                  onclick={(event) => openArchiveAction(event, 'archive_field_definition', definition.id)}
                >
                  <Trash2 />
                </Button.Root>
              </div>
            </article>
          {/each}
          {/if}
        </div>
      </section>
    </div>
  {/if}

  {#if operationStatus}
    <p class="denied-note" role={operationStatus.alert ? 'alert' : undefined}>{operationStatus.message}</p>
  {/if}
</section>
