import { fillMaxWidth } from '@expo/ui/jetpack-compose/modifiers';
import { Button, Host, Text, TextButton } from '@expo/ui/jetpack-compose';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

export function NativeCommandButton({ label, disabled = false, onPress, prominence = 'standard' }: NativeCommandButtonProps) {
  const Control = prominence === 'primary' ? Button : TextButton;
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
    <Control modifiers={prominence === 'primary' ? [fillMaxWidth()] : undefined} enabled={!disabled} onClick={() => { if (!disabled) onPress(); }}><Text>{label}</Text></Control>
  </Host>;
}
