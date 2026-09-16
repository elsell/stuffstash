import { NativeComposeHost as Host } from './NativeComposeHost.android';
import React from 'react';
import { Badge, BadgedBox, Icon, IconButton, Text } from '@expo/ui/jetpack-compose';
import { View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';
const icons = {
  notifications: require('./android-icons/header-notifications.xml'),
  add: require('./android-icons/header-add.xml'),
  account: require('./android-icons/header-account.xml'),
  close: require('./android-icons/header-close.xml'),
  back: require('./android-icons/header-back.xml'),
  save: require('./android-icons/header-save.xml'),
  settings: require('./android-icons/header-settings.xml'),
  compose: require('./android-icons/header-compose.xml'),
  'mark-read': require('./android-icons/header-mark-read.xml')
};
function Actions({ actions }: { readonly actions: readonly NativeHeaderAction[] }) {
  const palette = useAppearanceAwarePalette();
  return <View style={{ flexDirection: 'row' }}>{actions.map(action => {
    const icon = <Icon source={icons[action.kind]} size={24} tint={action.disabled ? palette.textMuted : palette.action} contentDescription={action.label} />;
    return (
    <Host key={action.kind} style={{ width: 48, height: 48 }}>
      <IconButton enabled={!action.disabled} onClick={() => { if (!action.disabled) action.onPress(); }}>
        {action.badgeCount !== undefined && action.badgeCount > 0 ? <BadgedBox>{icon}<BadgedBox.Badge><Badge>
          <Text>{action.badgeCount > 99 ? '99+' : String(action.badgeCount)}</Text>
        </Badge></BadgedBox.Badge></BadgedBox> : icon}
      </IconButton>
    </Host>
  ); })}</View>;
}
export function nativeHeaderActionOptions(actions: readonly NativeHeaderAction[], position: 'left' | 'right' = 'right'): HeaderOptions {
  const render = () => <Actions actions={actions} />;
  return position === 'left' ? { headerLeft: render } : { headerRight: render };
}
