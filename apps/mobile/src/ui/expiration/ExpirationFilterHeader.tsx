import { NativeRefinementButton } from '../components/NativeRefinementButton';
import type { ExpirationFilterHeaderProps } from './ExpirationFilterHeader.types';
/** Android uses the existing platform-native refinement button in its header. */
export function expirationFilterHeaderOptions({ active, onPress }: ExpirationFilterHeaderProps) {
 return { headerRight: () => <NativeRefinementButton iconOnly label="Filters" systemImage="line.3.horizontal.decrease.circle" accessibilityLabel={active ? 'Filter expiration items, filters active' : 'Filter expiration items'} badgeCount={active ? 1 : undefined} onPress={onPress} /> };
}
