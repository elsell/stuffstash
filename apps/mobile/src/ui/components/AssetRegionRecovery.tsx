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
  return <View style={{ gap: spacing.sm }}>
    <Text accessibilityRole="alert" style={{ color: palette.text, fontSize: 16 }}>
      {isRetrying ? `Loading ${region}…` : `Could not load ${region}.`}
    </Text>
    <NativeCommandButton label={`Retry ${region}`} disabled={isRetrying} onPress={onRetry} />
  </View>;
}
