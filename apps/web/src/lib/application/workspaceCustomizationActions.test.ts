import { describe, expect, it } from 'vitest';
import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';
import {
  customizationArchiveAssetTypeHref,
  customizationArchiveFieldDefinitionHref,
  customizationFieldsHref
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
