import { t } from '../../presentation/localization';
import { Text, View } from 'react-native';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { spacing } from '../theme/tokens';
import { NativeCommandButton } from './NativeCommandButton';

/** Persistent query recovery inside the route or sheet that owns the data. */
export function AssetRegionRecovery({ region, isRetrying, onRetry }: {
  readonly region: 'contents' | 'photos';
  readonly isRetrying: boolean;
  readonly onRetry: () => void;
}) {
  const palette = useAppearanceAwarePalette();
  const messages = region === 'contents'
    ? { loading: 'mobile.AssetRegionRecovery.contentsLoading', failed: 'mobile.AssetRegionRecovery.contentsFailed', retry: 'mobile.AssetRegionRecovery.contentsRetry' } as const
    : { loading: 'mobile.AssetRegionRecovery.photosLoading', failed: 'mobile.AssetRegionRecovery.photosFailed', retry: 'mobile.AssetRegionRecovery.photosRetry' } as const;
  return <View style={{ gap: spacing.sm }}>
    <Text accessibilityRole="alert" style={{ color: palette.text, fontSize: 16 }}>
      {t(isRetrying ? messages.loading : messages.failed)}
    </Text>
    <NativeCommandButton label={t(messages.retry)} disabled={isRetrying} onPress={onRetry} />
  </View>;
}
