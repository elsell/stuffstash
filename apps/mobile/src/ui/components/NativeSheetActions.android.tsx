import { NativeComposeHost as Host } from './NativeComposeHost.android';
import React from 'react';
import { Button, OutlinedButton, Text } from '@expo/ui/jetpack-compose';
import { fillMaxWidth } from '@expo/ui/jetpack-compose/modifiers';
import { View } from 'react-native';
import type { NativeSheetActionsProps } from './NativeSheetActions.types';

export function NativeSheetActions({ primaryLabel, primaryAccessibilityLabel = primaryLabel, secondaryLabel, secondaryAccessibilityLabel = secondaryLabel, disabled, secondaryDisabled = false, onApply, onBack }: NativeSheetActionsProps) {
  return <View style={{ gap: 8 }}>
    <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
      <Button enabled={!disabled} modifiers={[fillMaxWidth()]} onClick={() => { if (!disabled) onApply(); }}>
        <Text>{primaryLabel}</Text>
      </Button>
    </Host>
    <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
      <OutlinedButton enabled={!secondaryDisabled} modifiers={[fillMaxWidth()]} onClick={() => { if (!secondaryDisabled) onBack(); }}><Text>{secondaryLabel}</Text></OutlinedButton>
    </Host>
  </View>;
}
