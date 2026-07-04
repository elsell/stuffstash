import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import { workspaceRouteHref } from './workspaceRoute';

export type CustomizationArchiveKind = 'asset_type' | 'field_definition' | 'unavailable';

export interface CustomizationArchiveConfirmation {
  kind: CustomizationArchiveKind;
  title: string;
  targetLabel: string;
  description: string;
  buttonLabel: string;
  unavailable: boolean;
  disabled: boolean;
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
      title: 'Archive asset type',
      targetLabel: input.assetType.displayName,
      description: 'Existing assets keep their data. This type will stop appearing in new asset forms.',
      buttonLabel: 'Archive',
      unavailable: false,
      disabled: input.busy || !input.canArchiveScope(input.assetType.scope)
    };
  }
  if (input.fieldDefinition) {
    return {
      kind: 'field_definition',
      title: 'Archive field definition',
      targetLabel: input.fieldDefinition.displayName,
      description: 'Existing assets keep their field values. This field will stop appearing in edit forms.',
      buttonLabel: 'Archive',
      unavailable: false,
      disabled: input.busy || !input.canArchiveScope(input.fieldDefinition.scope)
    };
  }
  return {
    kind: 'unavailable',
    title: 'Archive target unavailable',
    targetLabel: 'This schema item is not available in the current fields list.',
    description: '',
    buttonLabel: 'Back to fields',
    unavailable: true,
    disabled: false
  };
}
