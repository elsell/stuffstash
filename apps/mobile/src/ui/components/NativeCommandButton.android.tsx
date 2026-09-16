import { NativeComposeHost as Host } from './NativeComposeHost.android';
import { fillMaxWidth } from '@expo/ui/jetpack-compose/modifiers';
import { Button, Text, TextButton } from '@expo/ui/jetpack-compose';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';
import { useAppearanceAwarePalette } from '../theme/appearance';

export function NativeCommandButton({ label, disabled = false, onPress, prominence = 'standard', role = 'default' }: NativeCommandButtonProps) {
  const palette = useAppearanceAwarePalette();
  const Control = prominence === 'primary' ? Button : TextButton;
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 48 }}>
    <Control colors={role === 'destructive' ? (prominence === 'primary' ? { containerColor: palette.danger, contentColor: palette.onAction } : { contentColor: palette.danger }) : undefined} modifiers={prominence === 'primary' ? [fillMaxWidth()] : undefined} enabled={!disabled} onClick={() => { if (!disabled) onPress(); }}><Text>{label}</Text></Control>
  </Host>;
}
