import { NativeComposeHost as Host } from './NativeComposeHost.android';
import { Icon, IconButton } from '@expo/ui/jetpack-compose';
import { useAppearancePalette } from '../theme/AppearanceContext';
import type { NativeConversationButtonProps } from './NativeConversationButton.types';
const icons = {
  record: require('./android-icons/conversation-record.xml'),
  send: require('./android-icons/conversation-send.xml'),
  cancel: require('./android-icons/conversation-cancel.xml')
};

export function NativeConversationButton({ kind, label, disabled = false, onPress }: NativeConversationButtonProps) {
  const palette = useAppearancePalette();
  return <Host style={{ width: 48, height: 48 }}>
    <IconButton enabled={!disabled} onClick={() => { if (!disabled) onPress(); }}>
      <Icon source={icons[kind]} size={24} tint={disabled ? palette.textMuted : palette.action} contentDescription={label} />
    </IconButton>
  </Host>;
}
