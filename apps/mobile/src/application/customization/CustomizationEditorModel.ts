import type {
  CustomFieldApplicability,
  CustomFieldType,
  CustomizationKind,
  CustomizationScope
} from '../../domain/customization/Customization';
import { customizationKeyIsValid, customizationKeyValidationMessage, normalizeTagColor, suggestedCustomizationKey } from '../../domain/customization/Customization';

export type CustomizationEditorDraft = {
  readonly name: string;
  readonly key: string;
  readonly keyManuallyEdited: boolean;
  readonly description: string;
  readonly color: string;
  readonly fieldType: CustomFieldType;
  readonly applicability: CustomFieldApplicability;
  readonly enumOptions: readonly string[];
  readonly targetIds: readonly string[];
};

export function emptyCustomizationEditorDraft(): CustomizationEditorDraft {
  return {
    name: '', key: '', keyManuallyEdited: false, description: '', color: '',
    fieldType: 'text', applicability: 'all_assets', enumOptions: [], targetIds: []
  };
}

export function withEditorName(draft: CustomizationEditorDraft, name: string): CustomizationEditorDraft {
  return {
    ...draft,
    name,
    key: draft.keyManuallyEdited ? draft.key : suggestedCustomizationKey(name)
  };
}

export function withManualEditorKey(draft: CustomizationEditorDraft, key: string): CustomizationEditorDraft {
  return { ...draft, key, keyManuallyEdited: true };
}

export function customizationEditorValidation(draft: CustomizationEditorDraft, kind: CustomizationKind, mode: 'create' | 'edit') {
  const colorValid = kind !== 'tag' || !draft.color.trim() || Boolean(normalizeTagColor(draft.color));
  return {
    colorValid,
    nameValid: draft.name.trim().length > 0,
    keyValid: mode === 'edit' || customizationKeyIsValid(draft.key || suggestedCustomizationKey(draft.name)),
    keyMessage: customizationKeyValidationMessage,
    optionsValid: kind !== 'field' || draft.fieldType !== 'enum' || draft.enumOptions.length > 0,
    targetsValid: kind !== 'field' || draft.applicability !== 'custom_asset_types' || draft.targetIds.length > 0
  } as const;
}

export function customizationEditorIsValid(draft: CustomizationEditorDraft, kind: CustomizationKind, mode: 'create' | 'edit'): boolean {
  const validation = customizationEditorValidation(draft, kind, mode);
  return validation.colorValid && validation.nameValid && validation.keyValid && validation.optionsValid && validation.targetsValid;
}

export function customizationEditorSnapshot(draft: CustomizationEditorDraft): string {
  return JSON.stringify({
    name: draft.name,
    key: draft.key,
    description: draft.description,
    color: draft.color,
    fieldType: draft.fieldType,
    applicability: draft.applicability,
    enumOptions: draft.enumOptions,
    targetIds: draft.targetIds
  });
}

export function customizationEditorIsDirty(
  draft: CustomizationEditorDraft,
  initialSnapshot: string,
  mode: 'create' | 'edit',
  completed: boolean
): boolean {
  if (completed) return false;
  if (mode === 'create') {
    return customizationEditorSnapshot(draft) !== customizationEditorSnapshot(emptyCustomizationEditorDraft());
  }
  return Boolean(initialSnapshot && customizationEditorSnapshot(draft) !== initialSnapshot);
}

export function effectiveInheritedOwnership(input: {
  readonly routeHint: boolean;
  readonly recordScope?: CustomizationScope;
  readonly screenScope: CustomizationScope;
}): boolean {
  if (input.recordScope) return input.screenScope === 'inventory' && input.recordScope === 'tenant';
  return input.routeHint;
}
