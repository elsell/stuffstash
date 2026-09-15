import React from 'react';
import { Pressable, Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeSheetActionsProps } from './NativeSheetActions.types';

/** Non-mobile renderer; iOS and Android resolve to real platform buttons. */
export function NativeSheetActions({ primaryLabel, primaryAccessibilityLabel = primaryLabel, secondaryLabel, secondaryAccessibilityLabel = secondaryLabel, disabled, secondaryDisabled = false, onApply, onBack }: NativeSheetActionsProps) {
  const palette = useAppearanceAwarePalette();
  const style = { minHeight: 52, alignItems: 'center', justifyContent: 'center', padding: 12 } as const;
  return <View style={{ gap: 8 }}>
    <Pressable accessibilityRole="button" accessibilityLabel={primaryAccessibilityLabel}
      accessibilityState={{ disabled }} disabled={disabled} onPress={() => { if (!disabled) onApply(); }}
      style={[style, { backgroundColor: palette.accent, opacity: disabled ? 0.5 : 1 }]}>
      <Text>{primaryLabel}</Text>
    </Pressable>
    <Pressable accessibilityRole="button" accessibilityLabel={secondaryAccessibilityLabel} disabled={secondaryDisabled} accessibilityState={{ disabled: secondaryDisabled }} onPress={() => { if (!secondaryDisabled) onBack(); }} style={style}>
      <Text style={{ color: palette.action }}>{secondaryLabel}</Text>
    </Pressable>
  </View>;
}
