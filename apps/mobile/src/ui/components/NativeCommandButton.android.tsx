import { nativeContentDescription } from './NativeComposeAccessibility.android';
import { NativeComposeHost as Host } from './NativeComposeHost.android';
import { fillMaxWidth } from '@expo/ui/jetpack-compose/modifiers';
import { Button, OutlinedButton, Text, TextButton } from '@expo/ui/jetpack-compose';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';
import { useAppearanceAwarePalette } from '../theme/appearance';

export function NativeCommandButton({ label, accessibilityLabel: accessibleName = label, disabled = false, onPress, prominence = 'secondary', role = 'default' }: NativeCommandButtonProps) {
  const palette = useAppearanceAwarePalette();
  const Control = prominence === 'primary' ? Button : prominence === 'secondary' ? OutlinedButton : TextButton;
  return <Host matchContents={prominence === 'secondary' ? true : { vertical: true }} style={prominence === 'secondary' ? { alignSelf: 'flex-start', maxWidth: '100%', minHeight: 48 } : { width: '100%', minHeight: 48 }}>
    <Control colors={role === 'destructive' ? (prominence === 'primary' ? { containerColor: palette.danger, contentColor: palette.onAction } : { contentColor: palette.danger }) : undefined} modifiers={[nativeContentDescription(accessibleName), ...(prominence === 'primary' ? [fillMaxWidth()] : [])]} enabled={!disabled} onClick={() => { if (!disabled) onPress(); }}><Text>{label}</Text></Control>
  </Host>;
}
