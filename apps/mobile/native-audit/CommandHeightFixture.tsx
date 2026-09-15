import { useState } from 'react';
import { Button as RNButton, ScrollView, Text as RNText, View } from 'react-native';
import { Button, Host, Text } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, disabled, fixedSize, frame } from '@expo/ui/swift-ui/modifiers';
import { NativeCommandButton } from '../src/ui/components/NativeCommandButton';

/** Runner-only sizing comparison. No production state or mutation. */
export function CommandHeightFixture() {
  const [shipping, setShipping] = useState(false);
  const [pressed, setPressed] = useState(false);
  const label = 'Retry asset types';
  return <ScrollView contentContainerStyle={{ padding: 24, gap: 24 }}>
    <RNButton title={shipping ? 'Compare baseline sizing' : 'Compare shipping sizing'} onPress={() => { setShipping(!shipping); setPressed(false); }} />
    <RNText>{shipping ? 'Shipping size' : 'Baseline size'}</RNText>
    <View style={{ width: 240, maxWidth: '100%', borderWidth: 1, borderColor: 'red' }}>
      {!shipping ? <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
        <Button onPress={() => setPressed(true)} modifiers={[
          buttonStyle('borderless'), disabled(false), accessibilityLabel(label)
        ]}>
          <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 48 })]}>{label}</Text>
        </Button>
      </Host> : <NativeCommandButton label={label} onPress={() => setPressed(true)} />}
    </View>
    <RNText>Following content</RNText>
    <RNText>{pressed ? 'Retry received' : 'Retry idle'}</RNText>
  </ScrollView>;
}
