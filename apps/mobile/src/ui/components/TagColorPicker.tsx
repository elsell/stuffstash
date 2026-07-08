import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Check, X } from 'lucide-react-native';
import { colors, radius, spacing } from '../theme/tokens';

export const tagColorChoices = [
  '#2F80ED',
  '#2E7D32',
  '#7C3AED',
  '#D97706',
  '#DC2626',
  '#0F766E'
] as const;

type TagColorPickerProps = {
  readonly value: string;
  readonly disabled?: boolean;
  readonly onChange: (value: string) => void;
};

export function TagColorPicker({ value, disabled = false, onChange }: TagColorPickerProps) {
  const normalizedValue = normalizeColor(value);
  return (
    <View accessibilityLabel="Tag color choices" style={styles.shell}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.swatches}>
        <Pressable
          accessibilityLabel="No tag color"
          accessibilityRole="button"
          accessibilityState={{ disabled, selected: normalizedValue === undefined }}
          disabled={disabled}
          onPress={() => onChange('')}
          style={[styles.clearSwatch, normalizedValue === undefined ? styles.selectedSwatch : null, disabled ? styles.disabled : null]}
        >
          <X color={colors.textMuted} size={15} strokeWidth={2.6} />
        </Pressable>
        {tagColorChoices.map((color) => {
          const selected = normalizedValue === color;
          return (
            <Pressable
              accessibilityLabel={`Choose tag color ${color}`}
              accessibilityRole="button"
              accessibilityState={{ disabled, selected }}
              disabled={disabled}
              key={color}
              onPress={() => onChange(color)}
              style={[
                styles.swatch,
                { backgroundColor: color },
                selected ? styles.selectedSwatch : null,
                disabled ? styles.disabled : null
              ]}
            >
              {selected ? <Check color={colors.onAction} size={14} strokeWidth={2.8} /> : null}
            </Pressable>
          );
        })}
      </ScrollView>
      <Text style={styles.fallbackLabel}>Or type a hex color</Text>
    </View>
  );
}

function normalizeColor(value: string): string | undefined {
  const raw = value.trim();
  if (!raw) {
    return undefined;
  }
  const color = raw.startsWith('#') ? raw : `#${raw}`;
  return /^#[0-9a-fA-F]{6}$/.test(color) ? color.toUpperCase() : undefined;
}

const styles = StyleSheet.create({
  shell: {
    gap: spacing.xs,
    minWidth: 0
  },
  swatches: {
    alignItems: 'center',
    gap: spacing.xs,
    paddingVertical: 2
  },
  swatch: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.lg,
    borderWidth: 1,
    height: 32,
    justifyContent: 'center',
    width: 32
  },
  clearSwatch: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.lg,
    borderWidth: 1,
    height: 32,
    justifyContent: 'center',
    width: 32
  },
  selectedSwatch: {
    borderColor: colors.action,
    borderWidth: 2
  },
  disabled: {
    opacity: 0.55
  },
  fallbackLabel: {
    color: colors.textMuted,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0
  }
});
