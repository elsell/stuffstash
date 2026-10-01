import { t } from '$lib/presentation/localization';
import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import { workspaceRouteHref } from './workspaceRoute';

export type CustomizationArchiveKind = 'asset_type' | 'field_definition' | 'unavailable';
export type CustomizationManagerStatusKind = 'missing-context' | 'denied' | 'error';

export interface CustomizationArchiveConfirmation {
  kind: CustomizationArchiveKind;
  title: string;
  targetLabel: string;
  description: string;
  buttonLabel: string;
  unavailable: boolean;
  disabled: boolean;
}

export interface CustomizationManagerStatus {
  kind: CustomizationManagerStatusKind;
  message: string;
  alert: boolean;
}

export function customizationFieldsHref(tenantId: string | null, inventoryId: string | null): string {
  return workspaceRouteHref({ mode: 'settings', settingsSection: 'fields' }, tenantId, inventoryId);
}

export function customizationArchiveAssetTypeHref(
  tenantId: string | null,
  inventoryId: string | null,
  assetType: CustomAssetType
): string {
  return workspaceRouteHref(
    {
      mode: 'settings',
      settingsSection: 'fields',
      customizationAction: 'archive_asset_type',
      customAssetTypeId: assetType.id
    },
    tenantId,
    inventoryId
  );
}

export function customizationArchiveFieldDefinitionHref(
  tenantId: string | null,
  inventoryId: string | null,
  definition: CustomFieldDefinition
): string {
  return workspaceRouteHref(
    {
      mode: 'settings',
      settingsSection: 'fields',
      customizationAction: 'archive_field_definition',
      customFieldDefinitionId: definition.id
    },
    tenantId,
    inventoryId
  );
}

export function customizationArchiveConfirmation(input: {
  assetType: CustomAssetType | null;
  fieldDefinition: CustomFieldDefinition | null;
  busy: boolean;
  canArchiveScope: (scope: CustomAssetType['scope']) => boolean;
}): CustomizationArchiveConfirmation {
  if (input.assetType) {
    return {
      kind: 'asset_type',
      title: t('web.workspaceCustomizationActions.archiveAssetType'),
      targetLabel: input.assetType.displayName,
      description: t('web.workspaceCustomizationActions.existingAssetsKeepTheirDataThisTypeWillStop'),
      buttonLabel: t('web.workspaceCustomizationActions.archive'),
      unavailable: false,
      disabled: input.busy || !input.canArchiveScope(input.assetType.scope)
    };
  }
  if (input.fieldDefinition) {
    return {
      kind: 'field_definition',
      title: t('web.workspaceCustomizationActions.archiveFieldDefinition'),
      targetLabel: input.fieldDefinition.displayName,
      description: t('web.workspaceCustomizationActions.existingAssetsKeepTheirFieldValuesThisFieldWill'),
      buttonLabel: t('web.workspaceCustomizationActions.archive'),
      unavailable: false,
      disabled: input.busy || !input.canArchiveScope(input.fieldDefinition.scope)
    };
  }
  return {
    kind: 'unavailable',
    title: t('web.workspaceCustomizationActions.archiveTargetUnavailable'),
    targetLabel: t('web.workspaceCustomizationActions.thisSchemaItemIsNotAvailableInTheCurrent'),
    description: '',
    buttonLabel: t('web.workspaceCustomizationActions.backToFields'),
    unavailable: true,
    disabled: false
  };
}

export function customizationManagerAccessStatus(input: {
  hasTenant: boolean;
  hasInventory: boolean;
  canManage: boolean;
}): CustomizationManagerStatus | null {
  if (!input.hasTenant || !input.hasInventory) {
    return {
      kind: 'missing-context',
      message: t('web.workspaceCustomizationActions.selectAnInventoryBeforeManagingFields'),
      alert: false
    };
  }
  if (!input.canManage) {
    return {
      kind: 'denied',
      message: t('web.workspaceCustomizationActions.customFieldsRequireTenantOrInventoryConfigurationAccess'),
      alert: true
    };
  }
  return null;
}

export function customizationManagerOperationStatus(error: string): CustomizationManagerStatus | null {
  if (!error) {
    return null;
  }
  return {
    kind: 'error',
    message: error,
    alert: true
  };
}
