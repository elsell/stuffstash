import { View } from 'react-native';
import { NativeChoicePicker } from './NativeChoicePicker';
import { useSettingsListStyles } from '../screens/SettingsList';

/** A value choice in the current form, not a navigation destination. */
export function SettingsPickerRow<Value extends string>({ label, accessibilityLabel, value, options, disabled, onChange }: {
  readonly label: string; readonly accessibilityLabel: string; readonly value: Value;
  readonly options: readonly { readonly value: Value; readonly label: string }[];
  readonly disabled?: boolean; readonly onChange: (value: Value) => void;
}) {
  const { styles } = useSettingsListStyles();
  return <View style={styles.navigationRow}><NativeChoicePicker label={label} accessibilityLabel={accessibilityLabel}
    value={value} options={options} disabled={disabled} includeEmptyOption={false}
    onChange={next => { const option = options.find(item => item.value === next); if (!disabled && option) onChange(option.value); }} />
  </View>;
}
