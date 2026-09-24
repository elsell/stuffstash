import { Button, Host, HStack, Spacer, Text } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, disabled as nativeDisabled, fixedSize, frame } from '@expo/ui/swift-ui/modifiers';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

export function NativeCommandButton({ label, accessibilityLabel: accessibleName = label, disabled = false, onPress, prominence = 'secondary', role = 'default' }: NativeCommandButtonProps) {
  const command = <Button role={role} onPress={() => { if (!disabled) onPress(); }} modifiers={[
      buttonStyle(prominence === 'primary' ? 'borderedProminent' : prominence === 'secondary' ? 'bordered' : 'borderless'), nativeDisabled(disabled), accessibilityLabel(accessibleName),
      fixedSize({ horizontal: false, vertical: true }), frame({ minWidth: 48, minHeight: 48 })
    ]}>
      {prominence === 'primary' ? <HStack><Spacer />
        <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minHeight: 32 })]}>{label}</Text>
        <Spacer /></HStack> : <Text modifiers={[fixedSize({ horizontal: false, vertical: true }), frame({ minWidth: 48, minHeight: 32 })]}>{label}</Text>}
    </Button>;
  return <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
    {prominence === 'secondary' ? <HStack>{command}<Spacer /></HStack> : command}
  </Host>;
}
