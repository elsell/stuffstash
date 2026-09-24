import { Pressable, Text } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeCommandButtonProps } from './NativeCommandButton.types';

/** Preview renderer; mobile platforms use native command controls. */
export function NativeCommandButton({ label, accessibilityLabel: accessibleName = label, disabled = false, onPress, prominence = 'secondary', role = 'default' }: NativeCommandButtonProps) {
  const palette = useAppearanceAwarePalette();
  return <Pressable accessibilityRole="button" accessibilityLabel={accessibleName}
    accessibilityState={{ disabled }} disabled={disabled} onPress={() => { if (!disabled) onPress(); }}
    style={{ minHeight: 48, justifyContent: 'center', padding: 8, alignSelf: prominence === 'secondary' ? 'flex-start' : undefined, borderWidth: prominence === 'secondary' ? 1 : 0, borderColor: palette.action, borderRadius: 8, backgroundColor: prominence === 'primary' ? role === 'destructive' ? palette.danger : palette.action : undefined }}>
    <Text style={{ color: disabled ? palette.textMuted : prominence === 'primary' ? palette.onAction : role === 'destructive' ? palette.danger : palette.action }}>{label}</Text>
  </Pressable>;
}
