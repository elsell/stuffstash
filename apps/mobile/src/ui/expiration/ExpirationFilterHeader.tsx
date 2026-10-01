import { t } from '../../presentation/localization';
import { NativeRefinementButton } from '../components/NativeRefinementButton';
import type { ExpirationFilterHeaderProps } from './ExpirationFilterHeader.types';
/** Android uses the existing platform-native refinement button in its header. */
export function expirationFilterHeaderOptions({ active, onPress }: ExpirationFilterHeaderProps) {
 return { headerRight: () => <NativeRefinementButton iconOnly label={t('mobile.ExpirationFilterHeader.filters')} systemImage="line.3.horizontal.decrease.circle" accessibilityLabel={active ? t('mobile.ExpirationFilterHeader.filterExpirationItemsFiltersActive') : t('mobile.ExpirationFilterHeader.filterExpirationItems')} badgeCount={active ? 1 : undefined} onPress={onPress} /> };
}
