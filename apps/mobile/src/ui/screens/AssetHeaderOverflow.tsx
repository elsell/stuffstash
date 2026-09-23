import { View } from 'react-native';
import { nativeHeaderActionOptions } from '../components/NativeHeaderActions';
import { AssetOverflowMenu } from './AssetOverflowMenu';
import type { AssetHeaderOverflowProps } from './AssetHeaderOverflow.types';

/** Android and non-native test renderer: keep the platform menu in headerRight. */
export function assetHeaderOverflowScreenOptions(props: AssetHeaderOverflowProps) {
  if (!props.onEdit) return { headerRight: () => <AssetOverflowMenu {...props} /> };
  const edit = nativeHeaderActionOptions([{ kind: 'compose', label: 'Edit',
    disabled: props.disabled, onPress: props.onEdit }]);
  return { headerRight: () => <View style={{ flexDirection: 'row', alignItems: 'center' }}>
    {edit.headerRight?.({ canGoBack: true })}<AssetOverflowMenu {...props} />
  </View> };
}
