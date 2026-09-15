import { type RefObject } from 'react';
import { StyleSheet, Text, View, type TextInput } from 'react-native';
import type { CustomAssetTypeDefinition, CustomFieldApplicability, CustomFieldType } from '../../domain/customization/Customization';
import { suggestedCustomizationKey } from '../../domain/customization/Customization';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { SettingsChoiceRow } from '../screens/SettingsList';
import { NativeChoicePicker } from './NativeChoicePicker';
import { NativeCommandButton } from './NativeCommandButton';
import { AppTextInput } from './AppTextInput';

export function CustomizationFieldControls(props: { readonly busy?: boolean; readonly applicability: CustomFieldApplicability; readonly canMutate: boolean; readonly eligibleTypes: readonly CustomAssetTypeDefinition[]; readonly enumOptions: readonly string[]; readonly fieldType: CustomFieldType; readonly mode: 'create' | 'edit'; readonly newOption: string; readonly onApplicability: (value: CustomFieldApplicability) => void; readonly onEnumOptions: (value: readonly string[]) => void; readonly onFieldType: (value: CustomFieldType) => void; readonly onNewOption: (value: string) => void; readonly onTargets: (value: readonly string[]) => void; readonly persistedEnumOptions: readonly string[]; readonly persistedTargetIds: readonly string[]; readonly targetIds: readonly string[] }) {
  const styles = createStyles(useAppearancePalette()); const types: readonly CustomFieldType[] = ['text', 'number', 'boolean', 'date', 'url', 'enum'];
  const disabled = !props.canMutate || Boolean(props.busy);
  const unavailableTargets = props.targetIds.filter(id => !props.eligibleTypes.some(type => type.id === id));
  const unavailableSavedCount = unavailableTargets.filter(id => props.persistedTargetIds.includes(id)).length;
  const unavailableDraftTargets = unavailableTargets.filter(id => !props.persistedTargetIds.includes(id));
  return <>
    <View style={styles.formRow}>{props.mode === 'edit' ? <Text style={styles.label}>Type</Text> : null}{props.mode === 'edit' ? <Text style={styles.lockedValue}>{capitalize(props.fieldType)}</Text> : <SingleChoicePicker disabled={disabled} label="Type" onChange={props.onFieldType} options={types.map((value) => ({ label: capitalize(value), value }))} value={props.fieldType} />}</View>
    {props.fieldType === 'enum' ? <View style={styles.formRow}><Text style={styles.label}>Options</Text>{props.enumOptions.map(option => props.persistedEnumOptions.includes(option) || !props.canMutate
      ? <Text key={option} style={styles.lockedValue}>{props.persistedEnumOptions.includes(option) ? `${option} · Existing` : option}</Text>
      : <NativeCommandButton key={option} label={`Remove ${option}`} disabled={disabled}
          onPress={() => { if (!disabled) props.onEnumOptions(props.enumOptions.filter(value => value !== option)); }} />)}{props.enumOptions.length === 0 ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>Add at least one option.</Text> : null}{props.canMutate ? <View style={styles.enumOptionInput}>
        <AppTextInput editable={!disabled} accessibilityLabel="New enum option" onChangeText={props.onNewOption}
          placeholder="Add option" style={[styles.input, styles.enumDraftInput]} value={props.newOption} />
        <NativeCommandButton label="Add option" disabled={disabled} onPress={() => {
          if (disabled) return;
          const next = suggestedCustomizationKey(props.newOption);
          if (next && !props.enumOptions.includes(next)) props.onEnumOptions([...props.enumOptions, next]);
          props.onNewOption('');
        }} />
      </View> : null}</View> : null}
    <View style={styles.formRow}>{props.mode === 'edit' ? <Text style={styles.label}>Applies to</Text> : null}{props.mode === 'edit' ? props.applicability === 'all_assets' ? <Text style={styles.lockedValue}>All assets</Text> : <><Text style={styles.lockedValue}>Selected asset types</Text>{props.canMutate ? <NativeCommandButton disabled={disabled} label="Expand to all assets" onPress={() => { if (!disabled) props.onApplicability('all_assets'); }} /> : null}</> : <SingleChoicePicker disabled={disabled} label="Applies to" onChange={props.onApplicability} options={[{ label: 'All assets', value: 'all_assets' }, { label: 'Selected asset types', value: 'custom_asset_types' }]} value={props.applicability} />}</View>
    {props.applicability === 'custom_asset_types' ? <View style={styles.formRow}>
      <Text style={styles.label}>Asset types</Text>
      {props.eligibleTypes.map(type => {
        const persisted = props.persistedTargetIds.includes(type.id);
        const selected = props.targetIds.includes(type.id);
        const label = `${type.displayName}${type.scope === 'tenant' ? ' · Inherited' : ''}`;
        if (persisted) return <Text key={type.id} style={styles.lockedValue}>{`${label} · Existing`}</Text>;
        if (!props.canMutate) return selected ? <Text key={type.id} style={styles.lockedValue}>{label}</Text> : null;
        return <SettingsChoiceRow key={type.id} label={label} multiple selected={selected} disabled={disabled}
          onPress={() => {
            if (!disabled) props.onTargets(selected ? props.targetIds.filter(id => id !== type.id) : [...props.targetIds, type.id]);
          }} />;
      })}
      {unavailableSavedCount > 0 ? <Text style={styles.lockedValue}>{`${unavailableSavedCount} existing asset ${unavailableSavedCount === 1 ? 'type is' : 'types are'} unavailable`}</Text> : null}
      {unavailableDraftTargets.length > 0 ? <SettingsChoiceRow label="Unavailable selections" accessibilityLabel="Include unavailable draft selections" multiple selected disabled={disabled}
        onPress={() => { if (!disabled) props.onTargets(props.targetIds.filter(id => !unavailableDraftTargets.includes(id))); }} /> : null}
      {props.targetIds.length === 0 && props.canMutate ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>Choose at least one asset type.</Text> : null}
      {props.eligibleTypes.length === 0 && props.targetIds.length === 0 ? <Text style={styles.readOnly}>No active asset types are available.</Text> : null}
    </View> : null}
  </>;
}

export function CustomizationLabeledInput({ editable, error, inputRef, label, multiline = false, onChangeText, required = false, value }: { readonly editable: boolean; readonly error?: string; readonly inputRef?: RefObject<TextInput | null>; readonly label: string; readonly multiline?: boolean; readonly onChangeText: (value: string) => void; readonly required?: boolean; readonly value: string }) { const styles = createStyles(useAppearancePalette()); return <View style={styles.formRow}><View style={styles.labelRow}><Text style={styles.label}>{label}</Text>{required ? <Text style={styles.required}>Required</Text> : null}</View><AppTextInput accessibilityHint={error ?? (required ? 'Required' : undefined)} accessibilityLabel={label} editable={editable} multiline={multiline} onChangeText={onChangeText} ref={inputRef} style={[styles.input, multiline && styles.multiline, !editable && styles.disabled]} value={value} />{error ? <Text accessibilityLiveRegion="polite" style={styles.validationText}>{error}</Text> : null}</View>; }
export function CustomizationReadOnlyValue({ label, value }: { readonly label: string; readonly value: string }) { const styles = createStyles(useAppearancePalette()); return <View style={styles.formRow}><Text style={styles.label}>{label}</Text><Text selectable style={styles.readOnlyValue}>{value}</Text></View>; }

function SingleChoicePicker<Value extends string>({ disabled, label, onChange, options, value }: { readonly disabled: boolean; readonly label: string; readonly onChange: (value: Value) => void; readonly options: readonly { readonly label: string; readonly value: Value }[]; readonly value: Value }) {
  const selected = options.find(option => option.value === value)?.label ?? value;
  return <NativeChoicePicker label={label} accessibilityLabel={`Choose ${label}. Current value ${selected}`}
    value={value} options={options} disabled={disabled} includeEmptyOption={false}
    onChange={next => { const option = options.find(item => item.value === next); if (!disabled && option) onChange(option.value); }} />;
}

function capitalize(value: string) { return value.charAt(0).toUpperCase() + value.slice(1).replaceAll('_', ' '); }

function createStyles(colors: MobileColorPalette) { return StyleSheet.create({
  formRow: { gap: spacing.sm, padding: spacing.md }, labelRow: { alignItems: 'center', flexDirection: 'row', justifyContent: 'space-between' }, label: { color: colors.text, fontSize: 15, fontWeight: '700' }, required: { color: colors.textMuted, fontSize: 13 }, input: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, color: colors.text, flex: 1, fontSize: 16, minHeight: 44, paddingHorizontal: spacing.sm, paddingVertical: spacing.sm }, multiline: { minHeight: 100, textAlignVertical: 'top' }, disabled: { opacity: 0.55 }, validationText: { color: colors.danger, fontSize: 13 }, readOnly: { color: colors.textMuted, fontSize: 14, lineHeight: 20 }, readOnlyValue: { color: colors.text, fontSize: 17, lineHeight: 23 }, lockedValue: { color: colors.text, fontSize: 15, minHeight: 30 }, enumOptionInput: { gap: spacing.sm }, enumDraftInput: { flex: 0 }
}); }
