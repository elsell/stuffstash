import React from 'react';
import { Button, Host, OutlinedButton, Text } from '@expo/ui/jetpack-compose';
import { fillMaxWidth } from '@expo/ui/jetpack-compose/modifiers';
import { View } from 'react-native';
import type { ExpirationSheetActionsProps } from './ExpirationSheetActions.types';

export function ExpirationSheetActions({ secondaryLabel, disabled, onApply, onBack }: ExpirationSheetActionsProps) {
  return <View style={{ gap: 8 }}>
    <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
      <Button enabled={!disabled} modifiers={[fillMaxWidth()]} onClick={() => { if (!disabled) onApply(); }}>
        <Text>Apply filters</Text>
      </Button>
    </Host>
    <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
      <OutlinedButton modifiers={[fillMaxWidth()]} onClick={onBack}><Text>{secondaryLabel}</Text></OutlinedButton>
    </Host>
  </View>;
}
