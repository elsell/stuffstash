import type { StackScreenProps } from 'expo-router';
import type { ExpirationFilterHeaderProps } from './ExpirationFilterHeader.types';
type NonFunction<T> = T extends (...args: any[]) => unknown ? never : T;
type Options = NonFunction<NonNullable<StackScreenProps['options']>>;
/** Real UIBarButtonItem lets UIKit own sizing, symbol rendering and glass. */
export function expirationFilterHeaderOptions({ active, onPress }: ExpirationFilterHeaderProps): Options {
 return { unstable_headerRightItems: () => [{
  type: 'button', label: 'Filters',
  accessibilityLabel: active ? 'Filter expiration items, filters active' : 'Filter expiration items',
  icon: { type: 'sfSymbol', name: active ? 'line.3.horizontal.decrease.circle.fill' : 'line.3.horizontal.decrease.circle' },
  onPress,
 }] };
}
