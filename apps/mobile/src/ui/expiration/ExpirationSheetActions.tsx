import React from 'react';
import { Pressable, Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { ExpirationSheetActionsProps } from './ExpirationSheetActions.types';

/** Non-mobile renderer; iOS and Android resolve to real platform buttons. */
export function ExpirationSheetActions({ secondaryLabel, disabled, onApply, onBack }: ExpirationSheetActionsProps) {
  const palette = useAppearanceAwarePalette();
  const style = { minHeight: 52, alignItems: 'center', justifyContent: 'center', padding: 12 } as const;
  return <View style={{ gap: 8 }}>
    <Pressable accessibilityRole="button" accessibilityLabel="Apply expiration filters"
      accessibilityState={{ disabled }} disabled={disabled} onPress={disabled ? undefined : onApply}
      style={[style, { backgroundColor: palette.accent, opacity: disabled ? 0.5 : 1 }]}>
      <Text>Apply filters</Text>
    </Pressable>
    <Pressable accessibilityRole="button" accessibilityLabel="Cancel or return to filters" onPress={onBack} style={style}>
      <Text style={{ color: palette.action }}>{secondaryLabel}</Text>
    </Pressable>
  </View>;
}
