import React from 'react';
import { Badge, BadgedBox, Host, Icon, IconButton, Text } from '@expo/ui/jetpack-compose';
import { View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';
const icons = {
  notifications: require('./android-icons/header-notifications.xml'),
  add: require('./android-icons/header-add.xml'),
  account: require('./android-icons/header-account.xml')
};
function Actions({ actions }: { readonly actions: readonly NativeHeaderAction[] }) {
  const palette = useAppearanceAwarePalette();
  return <View style={{ flexDirection: 'row' }}>{actions.map(action =>
    <Host key={action.kind} style={{ width: 48, height: 48 }}>
      <IconButton onClick={action.onPress}><BadgedBox>
        <Icon source={icons[action.kind]} size={24} tint={palette.action} contentDescription={action.label} />
        {action.badgeCount !== undefined && action.badgeCount > 0 ? <BadgedBox.Badge><Badge>
          <Text>{action.badgeCount > 99 ? '99+' : String(action.badgeCount)}</Text>
        </Badge></BadgedBox.Badge> : null}
      </BadgedBox></IconButton>
    </Host>
  )}</View>;
}
export function nativeHeaderActionOptions(actions: readonly NativeHeaderAction[]): HeaderOptions {
  return { headerRight: () => <Actions actions={actions} /> };
}
