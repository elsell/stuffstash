import { describe, expect, it } from 'vitest';
import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import {
  customizationArchiveConfirmation,
  customizationArchiveAssetTypeHref,
  customizationArchiveFieldDefinitionHref,
  customizationFieldsHref,
  customizationManagerAccessStatus,
  customizationManagerOperationStatus
} from './workspaceCustomizationActions';

describe('workspace customization actions', () => {
  it('builds canonical fields and archive action hrefs', () => {
    expect(customizationFieldsHref('tenant-one', 'inventory-one')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/fields'
    );
    expect(customizationArchiveAssetTypeHref('tenant-one', 'inventory-one', assetType())).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/fields/asset-types/type-medicine/archive'
    );
    expect(customizationArchiveFieldDefinitionHref('tenant-one', 'inventory-one', fieldDefinition())).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/fields/field-definitions/field-expiration/archive'
    );
  });

  it('builds asset type archive confirmation presentation', () => {
    expect(
      customizationArchiveConfirmation({
        assetType: assetType(),
        fieldDefinition: null,
        busy: false,
        canArchiveScope: () => true
      })
    ).toEqual({
      kind: 'asset_type',
      title: 'Archive asset type',
      targetLabel: 'Medicine',
      description: 'Existing assets keep their data. This type will stop appearing in new asset forms.',
      buttonLabel: 'Archive',
      unavailable: false,
      disabled: false
    });
  });

  it('builds field definition archive confirmation presentation with disabled state', () => {
    expect(
      customizationArchiveConfirmation({
        assetType: null,
        fieldDefinition: fieldDefinition(),
        busy: false,
        canArchiveScope: () => false
      })
    ).toEqual({
      kind: 'field_definition',
      title: 'Archive field definition',
      targetLabel: 'Expiration date',
      description: 'Existing assets keep their field values. This field will stop appearing in edit forms.',
      buttonLabel: 'Archive',
      unavailable: false,
      disabled: true
    });
  });

  it('builds unavailable archive confirmation presentation', () => {
    expect(
      customizationArchiveConfirmation({
        assetType: null,
        fieldDefinition: null,
        busy: false,
        canArchiveScope: () => true
      })
    ).toEqual({
      kind: 'unavailable',
      title: 'Archive target unavailable',
      targetLabel: 'This schema item is not available in the current fields list.',
      description: '',
      buttonLabel: 'Back to fields',
      unavailable: true,
      disabled: false
    });
  });

  it('builds customization manager missing-context and denied statuses', () => {
    expect(customizationManagerAccessStatus({ hasTenant: true, hasInventory: false, canManage: true })).toEqual({
      kind: 'missing-context',
      message: 'Select an inventory before managing fields.',
      alert: false
    });
    expect(customizationManagerAccessStatus({ hasTenant: true, hasInventory: true, canManage: false })).toEqual({
      kind: 'denied',
      message: 'Custom fields require tenant or inventory configuration access.',
      alert: true
    });
    expect(customizationManagerAccessStatus({ hasTenant: true, hasInventory: true, canManage: true })).toBeNull();
  });

  it('builds customization manager operation error status', () => {
    expect(customizationManagerOperationStatus('Schema service unavailable.')).toEqual({
      kind: 'error',
      message: 'Schema service unavailable.',
      alert: true
    });
    expect(customizationManagerOperationStatus('')).toBeNull();
  });
});

function assetType(): CustomAssetType {
  return {
    id: 'type-medicine',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: 'medicine',
    displayName: 'Medicine',
    description: '',
    lifecycleState: 'active'
  };
}

function fieldDefinition(): CustomFieldDefinition {
  return {
    id: 'field-expiration',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: 'expiration-date',
    displayName: 'Expiration date',
    type: 'date',
    enumOptions: [],
    applicability: 'all_assets',
    customAssetTypeIds: [],
    lifecycleState: 'active'
  };
}
