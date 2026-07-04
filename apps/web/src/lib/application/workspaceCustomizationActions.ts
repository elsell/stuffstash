import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import { workspaceRouteHref } from './workspaceRoute';

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
