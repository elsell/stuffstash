import { Pressable } from 'react-native';
import { Mail, MailOpen } from 'lucide-react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeReadStateButtonProps } from './NativeReadStateButton.types';

/** Preview fallback; mobile resolves the platform-native implementation. */
export function NativeReadStateButton({ read, label, disabled = false, onPress }: NativeReadStateButtonProps) {
  const palette = useAppearanceAwarePalette();
  const Symbol = read ? Mail : MailOpen;
  return <Pressable accessibilityRole="button" accessibilityLabel={label} accessibilityState={{ disabled }} disabled={disabled}
    onPress={() => { if (!disabled) onPress(); }} style={{ width: 48, height: 48, alignItems: 'center', justifyContent: 'center' }}>
    <Symbol size={20} color={disabled ? palette.textMuted : palette.action} />
  </Pressable>;
}
