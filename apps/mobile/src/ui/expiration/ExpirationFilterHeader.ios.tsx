import { t } from '../../presentation/localization';
import type { StackScreenProps } from 'expo-router';
import type { ExpirationFilterHeaderProps } from './ExpirationFilterHeader.types';
type NonFunction<T> = T extends (...args: any[]) => unknown ? never : T;
type Options = NonFunction<NonNullable<StackScreenProps['options']>>;
/** Real UIBarButtonItem lets UIKit own sizing, symbol rendering and glass. */
export function expirationFilterHeaderOptions({ active, onPress }: ExpirationFilterHeaderProps): Options {
 return { unstable_headerRightItems: () => [{
  type: 'button', label: t('mobile.ExpirationFilterHeaderios.filters'),
  accessibilityLabel: active ? t('mobile.ExpirationFilterHeaderios.filterExpirationItemsFiltersActive') : t('mobile.ExpirationFilterHeaderios.filterExpirationItems'),
  icon: { type: 'sfSymbol', name: active ? 'line.3.horizontal.decrease.circle.fill' : 'line.3.horizontal.decrease.circle' },
  onPress,
 }] };
}
