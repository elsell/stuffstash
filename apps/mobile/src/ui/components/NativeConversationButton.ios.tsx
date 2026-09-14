import { Button, Host } from '@expo/ui/swift-ui';
import { accessibilityLabel, buttonStyle, disabled as nativeDisabled, frame, labelStyle, tint } from '@expo/ui/swift-ui/modifiers';
import { useAppearancePalette } from '../theme/AppearanceContext';
import type { NativeConversationButtonProps } from './NativeConversationButton.types';
const symbols = { record: 'mic.fill', send: 'arrow.up', cancel: 'stop.fill' } as const;

export function NativeConversationButton({ kind, label, disabled = false, onPress }: NativeConversationButtonProps) {
  const palette = useAppearancePalette();
  return <Host style={{ width: 48, height: 48 }}>
    <Button label={label} systemImage={symbols[kind]} onPress={() => { if (!disabled) onPress(); }}
      modifiers={[buttonStyle('bordered'), labelStyle('iconOnly'), frame({ width: 48, height: 48 }),
        tint(palette.action), nativeDisabled(disabled), accessibilityLabel(label)]} />
  </Host>;
}
