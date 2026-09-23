import { Text, View } from 'react-native';
import { NativeCommandButton } from './NativeCommandButton';
import { useAssetTagSelectionVisit } from '../navigation/AssetTagSelectionTask';
import type { AssetTagSelectionOption } from '../screens/AssetTagSelectionScreen';
import { useSettingsListStyles } from '../screens/SettingsList';

export function AssetTagSelectionField({ scope, tags, selectedIds, disabled, onChange }: {
  readonly scope: string;
  readonly tags: readonly AssetTagSelectionOption[];
  readonly selectedIds: readonly string[];
  readonly disabled: boolean;
  readonly onChange: (ids: readonly string[]) => void;
}) {
  const open = useAssetTagSelectionVisit({ scope, tags, selectedIds, disabled, onChange });
  const { styles } = useSettingsListStyles();
  const labels = tags.filter(tag => selectedIds.includes(tag.id)).map(tag => tag.label);
  return <View>
    <NativeCommandButton label="Choose tags" disabled={disabled} onPress={open} />
    <Text numberOfLines={2} style={styles.rowContext}>{selectedIds.length} selected{labels.length ? ` · ${labels.join(', ')}` : ''}</Text>
  </View>;
}
