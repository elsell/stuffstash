import { t } from '../../presentation/localization';
import { useState, type ReactNode, type RefObject } from 'react';
import { Platform, StyleSheet, Text, View, type TextInput } from 'react-native';
import type { CustomAssetTypeDefinition, CustomFieldApplicability, CustomFieldType } from '../../domain/customization/Customization';
import { suggestedCustomizationKey } from '../../domain/customization/Customization';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { SettingsChoiceRow } from '../screens/SettingsList';
import { NativeChoicePicker } from './NativeChoicePicker';
import { NativeCommandButton } from './NativeCommandButton';
import { AppTextInput } from './AppTextInput';
import { DraftTextField } from './DraftTextField';

export function CustomizationFieldControls(props: { readonly persistedApplicability?: CustomFieldApplicability; readonly busy?: boolean; readonly applicability: CustomFieldApplicability; readonly canMutate: boolean; readonly eligibleTypes: readonly CustomAssetTypeDefinition[]; readonly enumOptions: readonly string[]; readonly fieldType: CustomFieldType; readonly mode: 'create' | 'edit'; readonly newOption: string; readonly onApplicability: (value: CustomFieldApplicability) => void; readonly onEnumOptions: (value: readonly string[]) => void; readonly onFieldType: (value: CustomFieldType) => void; readonly onNewOption: (value: string) => void; readonly onTargets: (value: readonly string[]) => void; readonly persistedEnumOptions: readonly string[]; readonly persistedTargetIds: readonly string[]; readonly targetIds: readonly string[] }) {
  const styles = createStyles(useAppearancePalette()); const types: readonly CustomFieldType[] = ['text', 'number', 'boolean', 'date', 'url', 'enum'];
  const disabled = !props.canMutate || Boolean(props.busy);
  const canChooseApplicability = props.canMutate && (props.mode === 'create' || props.persistedApplicability === 'custom_asset_types');
  const [optionError, setOptionError] = useState('');
  const [optionRevision, setOptionRevision] = useState(0);
  const unavailableTargets = props.targetIds.filter(id => !props.eligibleTypes.some(type => type.id === id));
  const unavailableSavedCount = unavailableTargets.filter(id => props.persistedTargetIds.includes(id)).length;
  const unavailableDraftTargets = unavailableTargets.filter(id => !props.persistedTargetIds.includes(id));
  return <>
    <View style={styles.formRow}>{props.mode === 'edit' ? <Text style={styles.label}>{t('mobile.CustomizationEditorFields.type')}</Text> : null}{props.mode === 'edit' ? <Text style={styles.lockedValue}>{capitalize(props.fieldType)}</Text> : <SingleChoicePicker disabled={disabled} label={t('mobile.CustomizationEditorFields.type')} onChange={props.onFieldType} options={types.map((value) => ({ label: capitalize(value), value }))} value={props.fieldType} />}</View>
    {props.fieldType === 'enum' ? <View style={styles.formRow}><Text style={styles.label}>{t('mobile.CustomizationEditorFields.options')}</Text>{props.enumOptions.map(option => props.persistedEnumOptions.includes(option) || !props.canMutate
      ? <Text key={option} style={styles.lockedValue}>{props.persistedEnumOptions.includes(option) ? t('mobile.CustomizationEditorFields.existing', { option: String(option) }) : option}</Text>
      : <NativeCommandButton key={option} label={t('mobile.CustomizationEditorFields.remove', { option: String(option) })} disabled={disabled}
          onPress={() => { if (!disabled) props.onEnumOptions(props.enumOptions.filter(value => value !== option)); }} />)}{props.enumOptions.length === 0 ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>{t('mobile.CustomizationEditorFields.addAtLeastOneOption')}</Text> : null}{props.canMutate ? <View style={styles.enumOptionInput}>
        <DraftTextField key={Platform.OS === 'ios' ? optionRevision : 'enum-option'} editable={!disabled} accessibilityLabel={t('mobile.CustomizationEditorFields.newEnumOption')} accessibilityHint={optionError || undefined} onChangeText={value => { setOptionError(''); props.onNewOption(value); }}
          placeholder={t('mobile.CustomizationEditorFields.addOption')} style={[styles.input, styles.enumDraftInput]} value={props.newOption} />
        {optionError ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>{optionError}</Text> : null}
        {props.newOption.trim() ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>{t('mobile.CustomizationEditorFields.addOrClearThisOptionBeforeSaving')}</Text> : null}
        <NativeCommandButton label={t('mobile.CustomizationEditorFields.addOption')} disabled={disabled} onPress={() => {
          if (disabled) return;
          const next = suggestedCustomizationKey(props.newOption);
          if (!next) { setOptionError('Use letters to start the option, then letters, numbers, or hyphens.'); return; }
          if (props.enumOptions.includes(next)) { setOptionError('This option already exists.'); return; }
          props.onEnumOptions([...props.enumOptions, next]);
          setOptionError('');
          props.onNewOption('');
          setOptionRevision(revision => revision + 1);
        }} />
      </View> : null}</View> : null}
    <View style={styles.formRow}>{canChooseApplicability
      ? <SingleChoicePicker disabled={disabled} label={t('mobile.CustomizationEditorFields.appliesTo')} onChange={props.onApplicability} options={[{ label: t('mobile.CustomizationEditorFields.allAssets'), value: 'all_assets' }, { label: t('mobile.CustomizationEditorFields.selectedAssetTypes'), value: 'custom_asset_types' }]} value={props.applicability} />
      : <><Text style={styles.label}>{t('mobile.CustomizationEditorFields.appliesTo')}</Text><Text style={styles.lockedValue}>{props.applicability === 'all_assets' ? t('mobile.CustomizationEditorFields.allAssets') : t('mobile.CustomizationEditorFields.selectedAssetTypes')}</Text></>}
    </View>
    {props.applicability === 'custom_asset_types' ? <View style={styles.formRow}>
      <Text style={styles.label}>{t('mobile.CustomizationEditorFields.assetTypes')}</Text>
      {props.eligibleTypes.map(type => {
        const persisted = props.persistedTargetIds.includes(type.id);
        const selected = props.targetIds.includes(type.id);
        const label = `${type.displayName}${type.scope === 'tenant' ? ' · Inherited' : ''}`;
        if (persisted) return <Text key={type.id} style={styles.lockedValue}>{t('mobile.CustomizationEditorFields.existing2', { label: String(label) })}</Text>;
        if (!props.canMutate) return selected ? <Text key={type.id} style={styles.lockedValue}>{label}</Text> : null;
        return <SettingsChoiceRow key={type.id} label={label} multiple selected={selected} disabled={disabled}
          onPress={() => {
            if (!disabled) props.onTargets(selected ? props.targetIds.filter(id => id !== type.id) : [...props.targetIds, type.id]);
          }} />;
      })}
      {unavailableSavedCount > 0 ? <Text style={styles.lockedValue}>{t('mobile.CustomizationEditorFields.existingAssetUnavailable', { unavailableSavedCount: String(unavailableSavedCount), value: String(unavailableSavedCount === 1 ? 'type is' : 'types are') })}</Text> : null}
      {unavailableDraftTargets.length > 0 ? <SettingsChoiceRow label={t('mobile.CustomizationEditorFields.unavailableSelections')} accessibilityLabel={t('mobile.CustomizationEditorFields.includeUnavailableDraftSelections')} multiple selected disabled={disabled}
        onPress={() => { if (!disabled) props.onTargets(props.targetIds.filter(id => !unavailableDraftTargets.includes(id))); }} /> : null}
      {props.targetIds.length === 0 && props.canMutate ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>{t('mobile.CustomizationEditorFields.chooseAtLeastOneAssetType')}</Text> : null}
      {props.eligibleTypes.length === 0 && props.targetIds.length === 0 ? <Text style={styles.readOnly}>{t('mobile.CustomizationEditorFields.noActiveAssetTypesAreAvailable')}</Text> : null}
    </View> : null}
  </>;
}

type LabeledInputFrameProps = { readonly label: string; readonly required?: boolean; readonly error?: string; readonly children: ReactNode };
function LabeledInputFrame({ label, required, error, children }: LabeledInputFrameProps) {
  const styles = createStyles(useAppearancePalette());
  return <View style={styles.formRow}>
    <View style={styles.labelRow}><Text style={styles.label}>{label}</Text>{required ? <Text style={styles.required}>{t('mobile.CustomizationEditorFields.required')}</Text> : null}</View>
    {children}
    {error ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>{error}</Text> : null}
  </View>;
}

export function CustomizationNameInput({ editable, error, onChangeText, value }: {
  readonly editable: boolean; readonly error?: string; readonly onChangeText: (value: string) => void; readonly value: string;
}) {
  const styles = createStyles(useAppearancePalette());
  return <LabeledInputFrame label={t('mobile.CustomizationEditorFields.name')} required error={error}>
    <DraftTextField accessibilityLabel={t('mobile.CustomizationEditorFields.name')} accessibilityHint={error ?? t('mobile.CustomizationEditorFields.required')}
      editable={editable} onChangeText={onChangeText} value={value} style={[styles.input, !editable && styles.disabled]} />
  </LabeledInputFrame>;
}

export function CustomizationLabeledInput({ editable, error, inputRef, label, multiline = false, onChangeText, required = false, value }: {
  readonly editable: boolean; readonly error?: string; readonly inputRef?: RefObject<TextInput | null>; readonly label: string;
  readonly multiline?: boolean; readonly onChangeText: (value: string) => void; readonly required?: boolean; readonly value: string;
}) {
  const styles = createStyles(useAppearancePalette());
  return <LabeledInputFrame label={label} required={required} error={error}>
    <AppTextInput accessibilityHint={error ?? (required ? t('mobile.CustomizationEditorFields.required') : undefined)} accessibilityLabel={label}
      editable={editable} multiline={multiline} onChangeText={onChangeText} ref={inputRef}
      style={[styles.input, multiline && styles.multiline, !editable && styles.disabled]} value={value} />
  </LabeledInputFrame>;
}
export function CustomizationReadOnlyValue({ label, value }: { readonly label: string; readonly value: string }) { const styles = createStyles(useAppearancePalette()); return <View style={styles.formRow}><Text style={styles.label}>{label}</Text><Text selectable style={styles.readOnlyValue}>{value}</Text></View>; }

function SingleChoicePicker<Value extends string>({ disabled, label, onChange, options, value }: { readonly disabled: boolean; readonly label: string; readonly onChange: (value: Value) => void; readonly options: readonly { readonly label: string; readonly value: Value }[]; readonly value: Value }) {
  const selected = options.find(option => option.value === value)?.label ?? value;
  return <NativeChoicePicker label={label} accessibilityLabel={t('mobile.CustomizationEditorFields.chooseCurrentValue', { label: String(label), selected: String(selected) })}
    value={value} options={options} disabled={disabled} includeEmptyOption={false}
    onChange={next => { const option = options.find(item => item.value === next); if (!disabled && option) onChange(option.value); }} />;
}

function capitalize(value: string) { return value.charAt(0).toUpperCase() + value.slice(1).replaceAll('_', ' '); }

function createStyles(colors: MobileColorPalette) { return StyleSheet.create({
  formRow: { gap: spacing.sm, padding: spacing.md }, labelRow: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between' }, label: { color: colors.text, fontSize: 15, fontWeight: '700' }, required: { color: colors.textMuted, fontSize: 13 }, input: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, color: colors.text, flex: 1, fontSize: 16, minHeight: 44, paddingHorizontal: spacing.sm, paddingVertical: spacing.sm }, multiline: { minHeight: 100, textAlignVertical: 'top' }, disabled: { opacity: 0.55 }, validationText: { color: colors.danger, fontSize: 13 }, readOnly: { color: colors.textMuted, fontSize: 14, lineHeight: 20 }, readOnlyValue: { color: colors.text, fontSize: 17, lineHeight: 23 }, lockedValue: { color: colors.text, fontSize: 15, minHeight: 30 }, enumOptionInput: { gap: spacing.sm }, enumDraftInput: { flex: 0 }
}); }
