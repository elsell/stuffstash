import { Host, Text, TextButton } from '@expo/ui/jetpack-compose';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

export function NativeCommandButton({ label, disabled = false, onPress }: NativeCommandButtonProps) {
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
    <TextButton enabled={!disabled} onClick={() => { if (!disabled) onPress(); }}><Text>{label}</Text></TextButton>
  </Host>;
}
