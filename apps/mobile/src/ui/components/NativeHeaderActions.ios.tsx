import type { HeaderOptions, NativeHeaderAction } from './NativeHeaderActions.types';
const symbols = { notifications: 'bell', add: 'plus', account: 'person.crop.circle' } as const;
export function nativeHeaderActionOptions(actions: readonly NativeHeaderAction[]): HeaderOptions {
  return { unstable_headerRightItems: () => actions.map(action => ({
    type: 'button', label: action.label, accessibilityLabel: action.label,
    icon: { type: 'sfSymbol', name: symbols[action.kind] },
    sharesBackground: true,
    ...(action.badgeCount && action.badgeCount > 0 ? { badge: { value: action.badgeCount > 99 ? '99+' : action.badgeCount } } : {}),
    onPress: action.onPress
  })) };
}
