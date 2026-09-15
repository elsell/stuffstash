import { Mic, ArrowUp, Square } from 'lucide-react-native';
import { Pressable } from 'react-native';
import { useAppearancePalette } from '../theme/AppearanceContext';
import type { NativeConversationButtonProps } from './NativeConversationButton.types';
const icons = { record: Mic, send: ArrowUp, cancel: Square };

export function NativeConversationButton({ kind, label, disabled = false, onPress }: NativeConversationButtonProps) {
  const palette = useAppearancePalette(); const Icon = icons[kind];
  return <Pressable accessibilityRole="button" accessibilityLabel={label} accessibilityState={{ disabled }}
    disabled={disabled} onPress={() => { if (!disabled) onPress(); }}
    style={{ width: 48, height: 48, alignItems: 'center', justifyContent: 'center' }}>
    <Icon size={24} color={disabled ? palette.textMuted : palette.action} />
  </Pressable>;
}
