import { Pressable, Text } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

/** Preview renderer; mobile platforms use native text buttons. */
export function NativeCommandButton({ label, disabled = false, onPress, prominence = 'standard' }: NativeCommandButtonProps) {
  const palette = useAppearanceAwarePalette();
  return <Pressable accessibilityRole="button" accessibilityLabel={label}
    accessibilityState={{ disabled }} disabled={disabled} onPress={() => { if (!disabled) onPress(); }}
    style={{ minHeight: 48, justifyContent: 'center', padding: 8, backgroundColor: prominence === 'primary' ? palette.action : undefined }}>
    <Text style={{ color: disabled ? palette.textMuted : prominence === 'primary' ? palette.onAction : palette.action }}>{label}</Text>
  </Pressable>;
}
