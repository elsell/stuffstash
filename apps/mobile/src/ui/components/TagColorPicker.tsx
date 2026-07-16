import { useState } from 'react';
import { KeyboardAvoidingView, Modal, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, useWindowDimensions, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Check, Palette, X } from 'lucide-react-native';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { FullSpectrumTagColorPicker } from './FullSpectrumTagColorPicker';
import { tagColorModalLayout } from './TagColorPickerPresentation';

export const tagColorChoices = [
  '#2F80ED',
  '#2E7D32',
  '#7C3AED',
  '#D97706',
  '#DC2626',
  '#0F766E'
] as const;
const tagColorNames: Readonly<Record<string, string>> = { '#2F80ED': 'Blue', '#2E7D32': 'Green', '#7C3AED': 'Purple', '#D97706': 'Orange', '#DC2626': 'Red', '#0F766E': 'Teal' };

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
  const customSelected = hasTypedColor && !tagColorChoices.includes(normalizedValue as typeof tagColorChoices[number]);
  const [customOpen, setCustomOpen] = useState(false);
  const [customDraft, setCustomDraft] = useState('');
  const [modalHeight, setModalHeight] = useState(0);
  const { fontScale } = useWindowDimensions();
  const normalizedDraft = normalizeColor(customDraft);
  const validDraft = !customDraft.trim() || Boolean(normalizedDraft);
  const modalLayout = tagColorModalLayout({ availableHeight: modalHeight, fontScale });

  function openCustom(): void {
    if (disabled) return;
    setCustomDraft(normalizedValue ?? '');
    setCustomOpen(true);
  }

  function closeCustom(): void { setCustomOpen(false); }

  function applyCustom(): void {
    if (!validDraft) return;
    onChange(normalizedDraft ?? '');
    setCustomOpen(false);
  }

  return (
    <View accessibilityLabel="Tag color choices" style={styles.shell}>
      <View style={styles.swatches}>
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
              accessibilityLabel={`Choose ${tagColorName(color)} tag color`}
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
      </View>
      <Pressable
        accessibilityLabel="Choose a custom tag color"
        accessibilityRole="button"
        accessibilityState={{ disabled, selected: customSelected }}
        disabled={disabled}
        onPress={openCustom}
        style={[styles.customButton, customSelected ? styles.selectedSwatch : null, disabled ? styles.disabled : null]}
      >
        <View testID="custom-tag-color-indicator" style={[styles.customIndicator, normalizedValue && customSelected ? { backgroundColor: normalizedValue } : null]}>
          {customSelected && normalizedValue ? <Check color={swatchForeground(normalizedValue)} size={14} strokeWidth={2.8} /> : <Palette color={colors.textMuted} size={16} />}
        </View>
        <Text style={styles.customLabel}>Custom…</Text>
      </Pressable>
      {invalidTypedColor ? <Text accessibilityLiveRegion="polite" style={styles.invalidLabel}>Choose Custom… to correct this color.</Text> : null}
      {customOpen ? <Modal animationType="slide" onRequestClose={closeCustom} presentationStyle="pageSheet" visible>
        <SafeAreaView edges={['bottom']} style={styles.safeArea}>
        <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={styles.safeArea}>
        <View accessibilityViewIsModal onLayout={(event) => setModalHeight(event.nativeEvent.layout.height)} style={styles.modalShell}>
          <View style={styles.modalHeader}><View><Text accessibilityRole="header" style={styles.modalTitle}>Custom color</Text><Text style={styles.modalSubtitle}>{normalizedDraft ? tagColorName(normalizedDraft) : 'No color'}</Text></View></View>
          <View style={styles.pickerSurface}>
            <FullSpectrumTagColorPicker compact={modalLayout.compactSpectrum} value={normalizedDraft ?? ''} onChange={setCustomDraft} />
          </View>
          <ScrollView automaticallyAdjustKeyboardInsets contentContainerStyle={styles.supplementaryContent} keyboardShouldPersistTaps="handled" style={styles.supplementaryScroll}>
            <Text style={styles.inputLabel}>Hex color</Text>
            <TextInput accessibilityLabel="Custom tag color hex value" autoCapitalize="characters" autoCorrect={false} onChangeText={setCustomDraft} placeholder="#2F80ED" placeholderTextColor={colors.textMuted} style={styles.hexInput} value={customDraft} />
            {!validDraft ? <Text accessibilityLiveRegion="polite" style={styles.invalidLabel}>Enter a #RRGGBB color.</Text> : null}
            <Pressable accessibilityRole="button" onPress={() => setCustomDraft('')} style={styles.clearAction}><Text style={styles.clearActionText}>Clear color</Text></Pressable>
          </ScrollView>
          <View style={styles.modalActions}>
            <Pressable accessibilityRole="button" onPress={closeCustom} style={styles.cancelAction}><Text style={styles.cancelActionText}>Cancel</Text></Pressable>
            <Pressable accessibilityRole="button" accessibilityState={{ disabled: !validDraft }} disabled={!validDraft} onPress={applyCustom} style={[styles.doneAction, !validDraft ? styles.disabled : null]}><Text style={styles.doneActionText}>Done</Text></Pressable>
          </View>
        </View>
        </KeyboardAvoidingView>
        </SafeAreaView>
      </Modal> : null}
    </View>
  );
}

export function tagColorName(color: string | undefined): string {
  if (!color) return 'No color';
  const normalized = normalizeColor(color);
  return normalized ? tagColorNames[normalized] ?? `Custom color ${normalized}` : 'Invalid color';
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
  safeArea: { backgroundColor: colors.background, flex: 1 },
  swatches: {
    alignItems: 'center',
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.xs,
    paddingVertical: 2
  },
  swatch: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.lg,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    width: 44
  },
  clearSwatch: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.lg,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    width: 44
  },
  selectedSwatch: {
    borderColor: colors.action,
    borderWidth: 2
  },
  disabled: {
    opacity: 0.55
  },
  invalidLabel: {
    color: colors.warning,
    fontSize: 11,
    fontWeight: '800',
    letterSpacing: 0
  },
  customButton: { alignItems: 'center', backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, flexDirection: 'row', gap: spacing.sm, justifyContent: 'center', minHeight: 44, paddingHorizontal: spacing.sm },
  customIndicator: { alignItems: 'center', backgroundColor: colors.surface, borderColor: colors.border, borderRadius: 11, borderWidth: 1, height: 22, justifyContent: 'center', width: 22 },
  customLabel: { color: colors.textMuted, fontSize: 14, fontWeight: '700' },
  modalShell: { flex: 1, paddingHorizontal: spacing.md, paddingTop: spacing.md },
  modalHeader: { minHeight: 48, justifyContent: 'center' }, modalTitle: { color: colors.text, fontSize: 24, fontWeight: '800' }, modalSubtitle: { color: colors.textMuted, fontSize: 14, marginTop: spacing.xs },
  pickerSurface: { flexShrink: 0, marginTop: spacing.md }, supplementaryScroll: { flex: 1, marginTop: spacing.md }, supplementaryContent: { gap: spacing.md, paddingBottom: spacing.md }, inputLabel: { color: colors.text, fontSize: 14, fontWeight: '700' },
  hexInput: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, color: colors.text, fontSize: 16, minHeight: 44, paddingHorizontal: spacing.sm },
  clearAction: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, justifyContent: 'center', minHeight: 44 }, clearActionText: { color: colors.warning, fontSize: 16, fontWeight: '700' },
  modalActions: { flexDirection: 'row', gap: spacing.sm, paddingBottom: spacing.md, paddingTop: spacing.sm }, cancelAction: { alignItems: 'center', borderColor: colors.border, borderRadius: radius.md, borderWidth: 1, flex: 1, justifyContent: 'center', minHeight: 48 }, cancelActionText: { color: colors.text, fontSize: 16, fontWeight: '700' }, doneAction: { alignItems: 'center', backgroundColor: colors.action, borderRadius: radius.md, flex: 1, justifyContent: 'center', minHeight: 48 }, doneActionText: { color: colors.onAction, fontSize: 16, fontWeight: '800' }
  });
}
