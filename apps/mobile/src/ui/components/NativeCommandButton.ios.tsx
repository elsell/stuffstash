import { Button, Host, Text } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, disabled as nativeDisabled, fixedSize, frame } from '@expo/ui/swift-ui/modifiers';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

export function NativeCommandButton({ label, disabled = false, onPress }: NativeCommandButtonProps) {
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
    <Button onPress={() => { if (!disabled) onPress(); }} modifiers={[
      buttonStyle('borderless'), nativeDisabled(disabled), accessibilityLabel(label)
    ]}>
      <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 48 })]}>{label}</Text>
    </Button>
  </Host>;
}
