import type {
  CustomAssetType,
  CustomDefinitionScope,
  CustomFieldApplicability,
  CustomFieldType
} from '$lib/domain/inventory';
import { customDefinitionScopes, customFieldApplicabilities, customFieldTypes } from '$lib/domain/inventory';

export interface CustomizationOption<TValue extends string = string> {
  value: TValue;
  label: string;
  description?: string;
  disabled?: boolean;
}

const scopeLabels: Record<CustomDefinitionScope, string> = {
  inventory: 'Inventory',
  tenant: 'Tenant'
};

const fieldTypeLabels: Record<CustomFieldType, string> = {
  text: 'Text',
  number: 'Number',
  boolean: 'Yes/no',
  date: 'Date',
  url: 'URL',
  enum: 'List'
};

const applicabilityLabels: Record<CustomFieldApplicability, string> = {
  all_assets: 'All assets',
  custom_asset_types: 'Custom types'
};

export function customizationScopeOptions(input: {
  canConfigureInventory: boolean;
  canConfigureTenant: boolean;
}): CustomizationOption<CustomDefinitionScope>[] {
  return customDefinitionScopes.map((scope) => ({
    value: scope,
    label: scopeLabels[scope],
    disabled: scope === 'inventory' ? !input.canConfigureInventory : !input.canConfigureTenant
  }));
}

export function customizationFieldTypeOptions(): CustomizationOption<CustomFieldType>[] {
  return customFieldTypes.map((type) => ({
    value: type,
    label: fieldTypeLabels[type]
  }));
}

export function customizationApplicabilityOptions(): CustomizationOption<CustomFieldApplicability>[] {
  return customFieldApplicabilities.map((applicability) => ({
    value: applicability,
    label: applicabilityLabels[applicability]
  }));
}

export function customizationTargetAssetTypeOptions(input: {
  assetTypes: CustomAssetType[];
  fieldScope: CustomDefinitionScope;
}): CustomizationOption[] {
  return input.assetTypes
    .filter((assetType) => assetType.lifecycleState === 'active')
    .filter((assetType) => input.fieldScope === 'inventory' || assetType.scope === 'tenant')
    .map((assetType) => ({
      value: assetType.id,
      label: assetType.displayName,
      description: scopeLabels[assetType.scope]
    }));
}
