import React from 'react';
import { Bell, Plus, UserCircle, X, Check, CheckCheck, Settings, ChevronLeft } from 'lucide-react-native';
import { Pressable, View } from 'react-native';
import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';
const icons = { notifications: Bell, add: Plus, account: UserCircle, close: X, back: ChevronLeft, save: Check, settings: Settings, 'mark-read': CheckCheck };
/** Non-mobile preview renderer. Native platforms supply their own adapter. */
export function nativeHeaderActionOptions(actions: readonly NativeHeaderAction[], position: 'left' | 'right' = 'right'): HeaderOptions {
  const render = () => <View style={{ flexDirection: 'row' }}>{actions.map(action => {
    const Icon = icons[action.kind];
    return <Pressable key={action.kind} accessibilityRole="button" accessibilityLabel={action.label} disabled={action.disabled} accessibilityState={{ disabled: action.disabled ?? false }} onPress={() => { if (!action.disabled) action.onPress(); }}
      style={{ minWidth: 48, minHeight: 48, alignItems: 'center', justifyContent: 'center' }}><Icon size={24} /></Pressable>;
  })}</View>;
  return position === 'left' ? { headerLeft: render } : { headerRight: render };
}
