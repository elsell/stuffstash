import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Check, X } from 'lucide-react-native';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearancePalette } from '../theme/AppearanceContext';

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
  readonly palette?: MobileColorPalette;
};

export function TagColorPicker({ value, disabled = false, onChange, palette }: TagColorPickerProps) {
  const contextPalette = useAppearancePalette();
  const colors = palette ?? contextPalette;
  const styles = createStyles(colors);
  const normalizedValue = normalizeColor(value);
  const hasTypedColor = value.trim().length > 0;
  const invalidTypedColor = hasTypedColor && normalizedValue === undefined;
  return (
    <View accessibilityLabel="Tag color choices" style={styles.shell}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.swatches}>
        <Pressable
          accessibilityLabel="No tag color"
          accessibilityRole="button"
          accessibilityState={{ disabled, selected: !hasTypedColor }}
          disabled={disabled}
          onPress={() => onChange('')}
          style={[styles.clearSwatch, !hasTypedColor ? styles.selectedSwatch : null, disabled ? styles.disabled : null]}
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
              {selected ? <Check color={swatchForeground(color)} size={14} strokeWidth={2.8} /> : null}
            </Pressable>
          );
        })}
      </ScrollView>
      <Text style={invalidTypedColor ? styles.invalidLabel : styles.fallbackLabel}>
        {invalidTypedColor ? 'Enter a #RRGGBB color' : 'Or type a hex color'}
      </Text>
    </View>
  );
}

export function swatchForeground(color: string): '#000000' | '#FFFFFF' {
  const channels = [1, 3, 5].map((index) => Number.parseInt(color.slice(index, index + 2), 16) / 255);
  const [red, green, blue] = channels.map((channel) => channel <= 0.04045
    ? channel / 12.92
    : ((channel + 0.055) / 1.055) ** 2.4);
  const luminance = (0.2126 * red) + (0.7152 * green) + (0.0722 * blue);
  return luminance >= 0.179 ? '#000000' : '#FFFFFF';
}

function normalizeColor(value: string): string | undefined {
  const raw = value.trim();
  if (!raw) {
    return undefined;
  }
  const color = raw.startsWith('#') ? raw : `#${raw}`;
  return /^#[0-9a-fA-F]{6}$/.test(color) ? color.toUpperCase() : undefined;
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
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
  },
  invalidLabel: {
    color: colors.warning,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0
  }
  });
}
