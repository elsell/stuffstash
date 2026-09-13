import React from 'react';
import { Button, Host, HStack, Spacer, Text, VStack } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, controlSize, disabled as nativeDisabled, fixedSize, frame } from '@expo/ui/swift-ui/modifiers';
import type { ExpirationSheetActionsProps } from './ExpirationSheetActions.types';

function ActionLabel({ children }: { readonly children: string }) {
  return <HStack>
    <Spacer />
    <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 24 })]}>{children}</Text>
    <Spacer />
  </HStack>;
}

export function ExpirationSheetActions({ secondaryLabel, disabled, onApply, onBack }: ExpirationSheetActionsProps) {
  // React Native proposes the width; SwiftUI measures only the resulting height.
  return <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
    <VStack spacing={8}>
      <Button onPress={() => { if (!disabled) onApply(); }} modifiers={[
        buttonStyle('borderedProminent'), controlSize('large'), nativeDisabled(disabled),
        accessibilityLabel('Apply expiration filters')
      ]}><ActionLabel>Apply filters</ActionLabel></Button>
      <Button onPress={onBack} modifiers={[
        buttonStyle('bordered'), controlSize('large'), accessibilityLabel('Cancel or return to filters')
      ]}><ActionLabel>{secondaryLabel}</ActionLabel></Button>
    </VStack>
  </Host>;
}
