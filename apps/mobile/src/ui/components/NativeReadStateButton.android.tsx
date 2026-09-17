import { NativeComposeHost as Host } from './NativeComposeHost.android';
import { Icon, IconButton } from '@expo/ui/jetpack-compose';
import { size } from '@expo/ui/jetpack-compose/modifiers';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { NativeReadStateButtonProps } from './NativeReadStateButton.types';

export function NativeReadStateButton({ read, label, disabled = false, onPress }: NativeReadStateButtonProps) {
  const palette = useAppearanceAwarePalette();
  const activate = () => { if (!disabled) onPress(); };
  return <Host style={{ width: 48, height: 48 }}>
      <IconButton enabled={!disabled} onClick={activate} modifiers={[size(48, 48)]}>
        <Icon size={24} contentDescription={label} tint={disabled ? palette.textMuted : palette.action}
          source={read ? require('./android-icons/envelope.xml') : require('./android-icons/envelope-open.xml')} />
      </IconButton>
    </Host>;
}
