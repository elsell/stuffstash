import { Text, View } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
import { useSettingsListStyles } from '../screens/SettingsList';
export type PrintCopiesControlProps = { readonly label: string; readonly value: string; readonly disabled: boolean; readonly onChange: (value: string) => void };
/** Android/default adapter retains direct numeric entry and server-owned limits. */
export function PrintCopiesControl({ label, value, disabled, onChange }: PrintCopiesControlProps) {
  const { styles, palette } = useSettingsListStyles();
  return <View style={styles.navigationRow}><Text style={styles.rowLabel}>{label}</Text>
    <AppTextInput accessibilityLabel={label} value={value} editable={!disabled} keyboardType="number-pad" inputMode="numeric"
      maxLength={String(Number.MAX_SAFE_INTEGER).length} selectTextOnFocus
      style={{ color: palette.text, fontSize: 17, minHeight: 44, borderWidth: 1, borderColor: palette.border, borderRadius: 8, paddingHorizontal: 12, marginTop: 8 }}
      onChangeText={next => { if (!disabled) onChange(next); }} />
  </View>;
}
