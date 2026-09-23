import { Button, Host, HStack, Spacer, Text } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, disabled as nativeDisabled, fixedSize, frame } from '@expo/ui/swift-ui/modifiers';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

export function NativeCommandButton({ label, disabled = false, onPress, prominence = 'standard', role = 'default' }: NativeCommandButtonProps) {
  return <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
    <Button role={role} onPress={() => { if (!disabled) onPress(); }} modifiers={[
      buttonStyle(prominence === 'primary' ? 'borderedProminent' : 'borderless'), nativeDisabled(disabled), accessibilityLabel(label),
      fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 48 })
    ]}>
      {prominence === 'primary' ? <HStack><Spacer />
        <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 32 })]}>{label}</Text>
        <Spacer /></HStack> : <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 48 })]}>{label}</Text>}
    </Button>
  </Host>;
}
