import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeRefinementButtonProps } from './NativeRefinementButton.types';

export type { NativeRefinementButtonProps } from './NativeRefinementButton.types';

/** Non-native renderer used by tests and non-mobile targets. */
export function NativeRefinementButton({
  accessibilityLabel,
  accessibilityState,
  badgeCount,
  disabled = false,
  iconOnly = false,
  label,
  onPress
}: NativeRefinementButtonProps) {
  const palette = useAppearanceAwarePalette();
  return <Pressable
    accessibilityLabel={accessibilityLabel}
    accessibilityRole="button"
    accessibilityState={{ ...accessibilityState, disabled }}
    disabled={disabled}
    onPress={disabled ? undefined : onPress}
    style={[styles.control, iconOnly ? styles.iconOnlyControl : null, { borderColor: palette.controlBorder }, disabled ? styles.disabled : null]}
  >
    <Text style={{ color: disabled ? palette.textMuted : palette.action }}>{iconOnly ? '☷' : label}</Text>
    {badgeCount && badgeCount > 0 ? <View
      accessible={false}
      importantForAccessibility="no-hide-descendants"
      pointerEvents="none"
      style={[styles.badge, { backgroundColor: palette.accent }]}
    >
      <Text style={styles.badgeText}>{badgeCount > 9 ? '9+' : badgeCount.toString()}</Text>
    </View> : null}
  </Pressable>;
}

const styles = StyleSheet.create({
  control: {
    alignItems: 'center',
    borderRadius: 18,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 44,
    minWidth: 44,
    paddingHorizontal: 12
  },
  iconOnlyControl: { height: 44, paddingHorizontal: 0, width: 44 },
  badge: {
    alignItems: 'center',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 16,
    minWidth: 16,
    paddingHorizontal: 3,
    position: 'absolute',
    right: -3,
    top: -3
  },
  badgeText: { color: '#FFFFFF', fontSize: 10, fontWeight: '700', lineHeight: 12 },
  disabled: { opacity: 0.5 }
});
