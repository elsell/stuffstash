import { AssetOverflowMenu } from './AssetOverflowMenu';
import type { AssetHeaderOverflowProps } from './AssetHeaderOverflow.types';

/** Android and non-native test renderer: keep the platform menu in headerRight. */
export function assetHeaderOverflowScreenOptions(props: AssetHeaderOverflowProps) {
  return { headerRight: () => <AssetOverflowMenu {...props} /> };
}
