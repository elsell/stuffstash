import { Button, Host } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, disabled as nativeDisabled, frame, labelStyle } from '@expo/ui/swift-ui/modifiers';
import type { NativeReadStateButtonProps } from './NativeReadStateButton.types';

export function NativeReadStateButton({ read, label, disabled = false, onPress }: NativeReadStateButtonProps) {
  return <Host style={{ width: 48, height: 48 }}>
    <Button label={label} systemImage={read ? 'envelope' : 'envelope.open'}
      modifiers={[buttonStyle('borderless'), labelStyle('iconOnly'), frame({ width: 48, height: 48 }), accessibilityLabel(label), nativeDisabled(disabled)]}
      onPress={() => { if (!disabled) onPress(); }} />
  </Host>;
}
