import { describe, expect, it } from 'vitest';
import {
  customizationApplicabilityOptions,
  customizationFieldTypeOptions,
  customizationScopeOptions,
  customizationTargetAssetTypeOptions
} from './workspaceCustomizationPresentation';
import type { CustomAssetType } from '$lib/domain/inventory';

describe('workspace customization presentation helpers', () => {
  it('builds scope options with access-aware disabled state', () => {
    expect(customizationScopeOptions({ canConfigureInventory: true, canConfigureTenant: false })).toEqual([
      { value: 'inventory', label: 'Inventory', disabled: false },
      { value: 'tenant', label: 'Tenant', disabled: true }
    ]);
  });

  it('builds field type options from canonical frontend-domain values', () => {
    expect(customizationFieldTypeOptions()).toEqual([
      { value: 'text', label: 'Text' },
      { value: 'number', label: 'Number' },
      { value: 'boolean', label: 'Yes/no' },
      { value: 'date', label: 'Date' },
      { value: 'url', label: 'URL' },
      { value: 'enum', label: 'List' }
    ]);
  });

  it('builds applicability options from canonical frontend-domain values', () => {
    expect(customizationApplicabilityOptions()).toEqual([
      { value: 'all_assets', label: 'All assets' },
      { value: 'custom_asset_types', label: 'Custom types' }
    ]);
  });

  it('filters target asset type options by lifecycle and selected field scope', () => {
    expect(
      customizationTargetAssetTypeOptions({
        assetTypes: [
          customAssetType('type-appliance', 'Appliance', 'tenant', 'active'),
          customAssetType('type-medicine', 'Medicine', 'inventory', 'active'),
          customAssetType('type-old', 'Old type', 'tenant', 'archived')
        ],
        fieldScope: 'tenant'
      })
    ).toEqual([{ value: 'type-appliance', label: 'Appliance', description: 'Tenant' }]);

    expect(
      customizationTargetAssetTypeOptions({
        assetTypes: [
          customAssetType('type-appliance', 'Appliance', 'tenant', 'active'),
          customAssetType('type-medicine', 'Medicine', 'inventory', 'active')
        ],
        fieldScope: 'inventory'
      })
    ).toEqual([
      { value: 'type-appliance', label: 'Appliance', description: 'Tenant' },
      { value: 'type-medicine', label: 'Medicine', description: 'Inventory' }
    ]);
  });
});

function customAssetType(
  id: string,
  displayName: string,
  scope: CustomAssetType['scope'],
  lifecycleState: CustomAssetType['lifecycleState']
): CustomAssetType {
  return {
    id,
    tenantId: 'tenant-one',
    inventoryId: scope === 'inventory' ? 'inventory-one' : null,
    scope,
    key: id,
    displayName,
    description: '',
    lifecycleState
  };
}
