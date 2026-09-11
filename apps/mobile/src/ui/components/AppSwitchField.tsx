import { StyleSheet, Switch, Text, View } from 'react-native';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

export function AppSwitchField({ label, description, value, onValueChange, disabled = false }: {
  readonly label: string;
  readonly description?: string;
  readonly value: boolean;
  readonly onValueChange: (value: boolean) => void;
  readonly disabled?: boolean;
}) {
  const colors = useAppearancePalette();
  return <View style={styles.field}>
    <View style={styles.text}>
      <Text style={[styles.label, { color: colors.text }]}>{label}</Text>
      {description ? <Text style={{ color: colors.textMuted }}>{description}</Text> : null}
    </View>
    <Switch accessibilityLabel={label} accessibilityHint={description} disabled={disabled} onValueChange={onValueChange} value={value} />
  </View>;
}

const styles = StyleSheet.create({
  field: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, paddingVertical: spacing.md },
  text: { flex: 1, gap: spacing.xs },
  label: { fontWeight: '600' },
});
