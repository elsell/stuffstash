<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
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
  import InventoryCustomizationArchivePanel, { customizationArchiveFocusTarget } from './InventoryCustomizationArchivePanel.svelte';
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
  let archiveActionTrigger = $state<HTMLElement | null>(null);
  let customizationHeading = $state<HTMLHeadingElement | null>(null);
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
      error = safeWorkspaceErrorMessage(caught, 'Custom asset type could not be created. Try again.');
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
      error = safeWorkspaceErrorMessage(caught, 'Custom field could not be created. Try again.');
    } finally {
      busy = false;
    }
  }

  async function archiveAssetType(assetType: CustomAssetType): Promise<boolean> {
    if (!tenant || !inventory || !canScope(assetType.scope)) return false;
    busy = true;
    error = '';
    try {
      const archived = await repository.archiveCustomAssetType(tenant.id, inventory.id, assetType.id, assetType.scope);
      const nextAssetTypes = assetTypes.map((candidate) => candidate.id === archived.id ? archived : candidate);
      const nextFieldTargets = fieldTargets.filter((id) => id !== assetType.id);
      assetTypes = nextAssetTypes;
      fieldTargets = nextFieldTargets;
      onSchemaChange(nextAssetTypes, fieldDefinitions);
      return true;
    } catch (caught) {
      error = safeWorkspaceErrorMessage(caught, 'Custom asset type could not be archived. Try again.');
      return false;
    } finally {
      busy = false;
    }
  }

  async function archiveFieldDefinition(definition: CustomFieldDefinition): Promise<boolean> {
    if (!tenant || !inventory || !canScope(definition.scope)) return false;
    busy = true;
    error = '';
    try {
      const archived = await repository.archiveCustomFieldDefinition(tenant.id, inventory.id, definition.id, definition.scope);
      const nextFieldDefinitions = fieldDefinitions.map((candidate) => candidate.id === archived.id ? archived : candidate);
      fieldDefinitions = nextFieldDefinitions;
      onSchemaChange(assetTypes, nextFieldDefinitions);
      return true;
    } catch (caught) {
      error = safeWorkspaceErrorMessage(caught, 'Custom field could not be archived. Try again.');
      return false;
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
    archiveActionTrigger = event.currentTarget instanceof HTMLElement ? event.currentTarget : null;
    onArchiveActionOpen(action, id);
  }

  function restoreArchiveActionFocus(event: Event): void {
    event.preventDefault();
    const trigger = archiveActionTrigger;
    archiveActionTrigger = null;
    customizationArchiveFocusTarget(trigger, customizationHeading)?.focus();
  }

  function closeArchiveAction(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
      return;
    }
    event.preventDefault();
  }
</script>

<section class="settings-panel wide customization-panel" aria-labelledby="settings-customization">
  <div class="settings-panel-heading">
    <Shapes aria-hidden="true" />
    <div>
      <h2 id="settings-customization" bind:this={customizationHeading} tabindex="-1">{t('web.InventoryCustomizationManager.customFields')}</h2>
      <p>{t('web.InventoryCustomizationManager.typesAndFieldsAvailableToThisInventory')}</p>
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
        error={operationStatus?.message ?? ''}
        fieldsHref={fieldsHref()}
        canArchiveScope={canScope}
        onClose={closeArchiveAction}
        onDismiss={onArchiveActionClose}
        onCloseAutoFocus={restoreArchiveActionFocus}
        onArchiveAssetType={archiveAssetType}
        onArchiveFieldDefinition={archiveFieldDefinition}
      />
    {/if}

    <div class="customization-grid">
      <section class="customization-column customization-surface" aria-labelledby="custom-asset-types-title">
        <div class="customization-surface-heading">
          <h3 id="custom-asset-types-title">{t('web.InventoryCustomizationManager.assetTypes')}</h3>
          <span aria-label={`${activeAssetTypes.length} custom asset types`}>{activeAssetTypes.length}</span>
        </div>
        <SegmentedControl
          label={t('web.InventoryCustomizationManager.customTypeScope')}
          value={typeScope}
          options={scopeOptions}
          onSelect={(value) => { typeScope = value as CustomDefinitionScope; }}
        />
        <div class="field-stack">
          <Label for="custom-type-key">{t('web.InventoryCustomizationManager.key')}</Label>
          <Input id="custom-type-key" bind:value={typeKey} placeholder={t('web.InventoryCustomizationManager.medicine')} />
        </div>
        <div class="field-stack">
          <Label for="custom-type-name">{t('web.InventoryCustomizationManager.displayName')}</Label>
          <Input id="custom-type-name" bind:value={typeName} placeholder={t('web.InventoryCustomizationManager.medicine2')} />
        </div>
        <div class="field-stack">
          <Label for="custom-type-description">{t('web.InventoryCustomizationManager.description')}</Label>
          <Textarea id="custom-type-description" bind:value={typeDescription} placeholder={t('web.InventoryCustomizationManager.optional')} />
        </div>
        <Button.Root disabled={busy || !typeKey.trim() || !typeName.trim() || !canScope(typeScope)} onclick={() => { void createAssetType(); }}>{t('web.InventoryCustomizationManager.createType')}</Button.Root>

        <div class="schema-list" aria-label={t('web.InventoryCustomizationManager.customAssetTypes')}>
          {#if activeAssetTypes.length === 0}
            <p class="schema-empty">{t('web.InventoryCustomizationManager.noCustomAssetTypesYet')}</p>
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
          <h3 id="custom-field-definitions-title">{t('web.InventoryCustomizationManager.fieldDefinitions')}</h3>
          <span aria-label={`${activeFieldDefinitions.length} custom fields`}>{activeFieldDefinitions.length}</span>
        </div>
        <SegmentedControl
          label={t('web.InventoryCustomizationManager.customFieldScope')}
          value={fieldScope}
          options={scopeOptions}
          onSelect={(value) => selectFieldScope(value as CustomDefinitionScope)}
        />
        <div class="field-stack">
          <Label for="custom-field-key">{t('web.InventoryCustomizationManager.key')}</Label>
          <Input id="custom-field-key" bind:value={fieldKey} placeholder={t('web.InventoryCustomizationManager.expirationDate')} />
        </div>
        <div class="field-stack">
          <Label for="custom-field-name">{t('web.InventoryCustomizationManager.displayName')}</Label>
          <Input id="custom-field-name" bind:value={fieldName} placeholder={t('web.InventoryCustomizationManager.expirationDate2')} />
        </div>
        <SegmentedControl
          label={t('web.InventoryCustomizationManager.customFieldType')}
          value={fieldType}
          options={fieldTypeOptions}
          onSelect={(value) => { fieldType = value as CustomFieldType; }}
        />
        {#if fieldType === 'enum'}
          <div class="field-stack">
            <Label for="custom-field-options">{t('web.InventoryCustomizationManager.options')}</Label>
            <Input id="custom-field-options" bind:value={enumOptions} placeholder={t('web.InventoryCustomizationManager.newOpenClosed')} />
          </div>
        {/if}
        <SegmentedControl
          label={t('web.InventoryCustomizationManager.fieldApplicability')}
          value={fieldApplicability}
          options={applicabilityOptions}
          onSelect={(value) => { fieldApplicability = value as CustomFieldApplicability; }}
        />
        {#if fieldApplicability === 'custom_asset_types'}
          <fieldset class="selection-field">
            <legend>{t('web.InventoryCustomizationManager.fieldCustomTypeTargets')}</legend>
            <p class="selection-summary">
              {selectedTargetCount === 0
                ? 'No custom types selected'
                : `${selectedTargetCount} custom ${selectedTargetCount === 1 ? 'type' : 'types'} selected`}
            </p>
            <ChoiceGrid
              label={t('web.InventoryCustomizationManager.fieldCustomTypeTargets')}
              options={targetableAssetTypeOptions}
              selectedValues={fieldTargets}
              emptyMessage={t('web.InventoryCustomizationManager.noEligibleCustomAssetTypesForThisScope')}
              onSelect={toggleTarget}
            />
          </fieldset>
        {/if}
        <Button.Root disabled={busy || !fieldKey.trim() || !fieldName.trim() || !canScope(fieldScope)} onclick={() => { void createFieldDefinition(); }}>{t('web.InventoryCustomizationManager.createField')}</Button.Root>

        <div class="schema-list" aria-label={t('web.InventoryCustomizationManager.customFieldDefinitions')}>
          {#if activeFieldDefinitions.length === 0}
            <p class="schema-empty">{t('web.InventoryCustomizationManager.noCustomFieldsYet')}</p>
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
