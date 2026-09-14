import React from 'react';
import { Bell, Plus, UserCircle } from 'lucide-react-native';
import { Pressable, View } from 'react-native';
import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';
const icons = { notifications: Bell, add: Plus, account: UserCircle };
/** Non-mobile preview renderer. Native platforms supply their own adapter. */
export function nativeHeaderActionOptions(actions: readonly NativeHeaderAction[]): HeaderOptions {
  return { headerRight: () => <View style={{ flexDirection: 'row' }}>{actions.map(action => {
    const Icon = icons[action.kind];
    return <Pressable key={action.kind} accessibilityRole="button" accessibilityLabel={action.label} onPress={action.onPress}
      style={{ minWidth: 48, minHeight: 48, alignItems: 'center', justifyContent: 'center' }}><Icon size={24} /></Pressable>;
  })}</View> };
}
