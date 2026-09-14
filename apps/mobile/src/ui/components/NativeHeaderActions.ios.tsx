import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';
const symbols = { notifications: 'bell', add: 'plus', account: 'person.crop.circle', close: 'xmark', save: 'checkmark', settings: 'gearshape', 'mark-read': 'checkmark.message' } as const;
export function nativeHeaderActionOptions(actions: readonly NativeHeaderAction[], position: 'left' | 'right' = 'right'): HeaderOptions {
  const items: NonNullable<HeaderOptions['unstable_headerRightItems']> = () => actions.map(action => ({
    type: 'button', width: 44, label: action.label, accessibilityLabel: action.label,
    icon: { type: 'sfSymbol', name: symbols[action.kind] },
    sharesBackground: true, disabled: action.disabled ?? false,
    ...(action.badgeCount && action.badgeCount > 0 ? { badge: { value: action.badgeCount > 99 ? '99+' : action.badgeCount } } : {}),
    onPress: () => { if (!action.disabled) action.onPress(); }
  }));
  return position === 'left' ? { unstable_headerLeftItems: items } : { unstable_headerRightItems: items };
}
