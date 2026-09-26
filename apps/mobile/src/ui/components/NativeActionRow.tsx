import { useCommittedCommand } from './useCommittedCommand';
import { Pressable, StyleSheet, Text } from 'react-native';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { spacing } from '../theme/tokens';

export function NativeActionRow({ label, accessibilityLabel = label, disabled, role, onPress }: NativeCommandButtonProps) {
  const press = useCommittedCommand(onPress, disabled);
  const palette = useAppearanceAwarePalette();
  return <Pressable accessibilityRole="button" accessibilityLabel={accessibilityLabel}
    accessibilityState={{ disabled }} disabled={disabled} onPress={press}
    style={styles.row}>
    <Text style={[styles.label, { color: role === 'destructive' ? palette.danger : palette.action, opacity: disabled ? 0.5 : 1 }]}>{label}</Text>
  </Pressable>;
}

const styles = StyleSheet.create({
  row: { justifyContent: 'center', minHeight: 52, minWidth: 44, paddingHorizontal: spacing.md, paddingVertical: spacing.sm },
  label: { fontSize: 17, fontWeight: '500' }
});
