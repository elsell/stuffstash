import { ChevronDown, ChevronRight } from 'lucide-react-native';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { ReactNode } from 'react';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

/** A labeled disclosure with its current value; selection content belongs to its caller. */
export function SelectionRow({ label, accessibilityLabel, value, expanded, disabled, onPress, children }: {
  readonly label: string; readonly accessibilityLabel?: string; readonly value: string; readonly expanded?: boolean; readonly disabled?: boolean;
  readonly onPress: () => void; readonly children?: ReactNode;
}) {
  const colors = useAppearancePalette();
  const Chevron = expanded ? ChevronDown : ChevronRight;
  return <View style={{ flexShrink: 0 }}>
    <Pressable accessibilityRole="button" accessibilityLabel={accessibilityLabel ?? label} accessibilityValue={{ text: value }} accessibilityState={{ disabled: !!disabled, ...(expanded === undefined ? {} : { expanded }) }} disabled={disabled} onPress={onPress} style={[styles.row, { borderColor: colors.border }]}>
      <Text style={{ color: colors.text, fontSize: 17, flexShrink: 1 }}>{label}</Text>
      <Text style={{ color: colors.textMuted, fontSize: 17, flex: 1, textAlign: 'right' }}>{value}</Text>
      <Chevron size={18} color={colors.textMuted} />
    </Pressable>
    <View style={expanded ? undefined : { display: 'none' }}>{children}</View>
  </View>;
}
const styles = StyleSheet.create({ row: { minHeight: 48, paddingVertical: spacing.sm, borderBottomWidth: StyleSheet.hairlineWidth, flexDirection: 'row', alignItems: 'center', gap: spacing.sm } });
