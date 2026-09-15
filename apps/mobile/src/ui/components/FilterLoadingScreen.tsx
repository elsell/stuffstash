import { ScrollView } from 'react-native';
import { SettingsLoadingRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { NativeCommandButton } from './NativeCommandButton';

export function FilterLoadingScreen({ onCancel }: { readonly onCancel: () => void }) {
  const { styles } = useSettingsListStyles();
  return <ScrollView style={styles.shell} contentInsetAdjustmentBehavior="automatic">
    <SettingsSection>
      <SettingsLoadingRow label="Loading filters" />
      <NativeCommandButton label="Cancel" onPress={onCancel} />
    </SettingsSection>
  </ScrollView>;
}
