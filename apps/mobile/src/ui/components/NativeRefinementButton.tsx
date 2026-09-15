import React from 'react';
import { Pressable, StyleSheet, Text } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { RefinementCountBadge } from './RefinementCountBadge';
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
    <RefinementCountBadge count={badgeCount} />
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
  disabled: { opacity: 0.5 }
});
