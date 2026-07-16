import SegmentedControl from '@expo/ui/community/segmented-control';
import { Platform, Pressable, StyleSheet, Text, View } from 'react-native';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { expoSegmentedControlAvailable } from './SettingsSegmentedControlPresentation';

export type SettingsSegment = { readonly label: string; readonly value: string };

export function SettingsSegmentedControl({
  disabled = false,
  onChange,
  segments,
  value
}: {
  readonly disabled?: boolean;
  readonly onChange: (value: string) => void;
  readonly segments: readonly SettingsSegment[];
  readonly value: string;
}) {
  const colors = useAppearancePalette();
  const styles = createStyles(colors);
  const selectedIndex = Math.max(0, segments.findIndex((segment) => segment.value === value));

  if (expoSegmentedControlAvailable(Platform.OS)) {
    return <SegmentedControl
      enabled={!disabled}
      onValueChange={(label) => {
        const segment = segments.find((candidate) => candidate.label === label);
        if (segment) onChange(segment.value);
      }}
      selectedIndex={selectedIndex}
      style={styles.native}
      tintColor={colors.selected}
      values={segments.map((segment) => segment.label)}
    />;
  }

  return <View accessibilityRole="tablist" style={styles.fallback}>
    {segments.map((segment) => {
      const selected = segment.value === value;
      return <Pressable
        accessibilityRole="tab"
        accessibilityState={{ disabled, selected }}
        disabled={disabled}
        key={segment.value}
        onPress={() => onChange(segment.value)}
        style={[styles.segment, selected && styles.selected, disabled && styles.disabled]}
      >
        <Text style={[styles.label, selected && styles.selectedLabel]}>{segment.label}</Text>
      </Pressable>;
    })}
  </View>;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
    native: { minHeight: 44 },
    fallback: { backgroundColor: colors.surfaceMuted, borderRadius: radius.md, flexDirection: 'row', padding: 3 },
    segment: { alignItems: 'center', borderRadius: radius.sm, flex: 1, justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.sm },
    selected: { backgroundColor: colors.surface },
    disabled: { opacity: 0.55 },
    label: { color: colors.textMuted, fontSize: 15, fontWeight: '600' },
    selectedLabel: { color: colors.text }
  });
}
