import { StyleSheet, View } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';

export function VoicePlanNameEditor({ value, onChange, onSave, onCancel }: {
  readonly value: string;
  readonly onChange: (value: string) => void;
  readonly onSave: (value: string) => void;
  readonly onCancel: () => void;
}) {
  const palette = useAppearancePalette();
  const normalized = value.replace(/\s+/g, ' ').trim();
  const save = () => { if (normalized) onSave(normalized); };
  return <View style={styles.editor}>
    <AppTextInput accessibilityLabel="Proposed item name" autoFocus maxLength={200}
      value={value} onChangeText={onChange} onSubmitEditing={save}
      returnKeyType="done" selectTextOnFocus
      style={[styles.input, { color: palette.text, backgroundColor: palette.surface, borderColor: palette.accent }]} />
    <View style={styles.commands}>
      <View style={styles.command}><NativeCommandButton label="Cancel" onPress={onCancel} /></View>
      <View style={styles.command}><NativeCommandButton label="Save" disabled={!normalized} onPress={save} /></View>
    </View>
  </View>;
}

const styles = StyleSheet.create({
  editor: { alignItems: 'stretch', gap: spacing.xs },
  commands: { flexDirection: 'row', gap: spacing.sm },
  command: { flex: 1, minWidth: 0 },
  input: { borderRadius: radius.sm, borderWidth: 2, fontSize: 16, fontWeight: '700', minHeight: 44, paddingHorizontal: spacing.sm }
});
