import { t } from '$lib/presentation/localization';
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
  inventory: t('web.workspaceCustomizationPresentation.inventory'),
  tenant: t('web.workspaceCustomizationPresentation.tenant')
};

const fieldTypeLabels: Record<CustomFieldType, string> = {
  text: t('web.workspaceCustomizationPresentation.text'),
  number: t('web.workspaceCustomizationPresentation.number'),
  boolean: t('web.FieldSettingsManager.yesNo'),
  date: t('web.workspaceCustomizationPresentation.date'),
  url: t('web.workspaceCustomizationPresentation.uRL'),
  enum: t('web.workspaceCustomizationPresentation.list')
};

const applicabilityLabels: Record<CustomFieldApplicability, string> = {
  all_assets: t('web.workspaceCustomizationPresentation.allAssets'),
  custom_asset_types: t('web.workspaceCustomizationPresentation.customTypes')
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

export function customizationScopeLabel(scope: CustomDefinitionScope): string {
  return scopeLabels[scope];
}
