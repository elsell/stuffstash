import { t } from '../../presentation/localization';
import { ScrollView } from 'react-native';
import { SettingsLoadingRow, SettingsSection, useSettingsListStyles } from '../screens/SettingsList';
import { NativeCommandButton } from './NativeCommandButton';

export function FilterLoadingScreen({ onCancel }: { readonly onCancel: () => void }) {
  const { styles } = useSettingsListStyles();
  return <ScrollView style={styles.shell} contentInsetAdjustmentBehavior="automatic">
    <SettingsSection>
      <SettingsLoadingRow label={t('mobile.FilterLoadingScreen.loadingFilters')} />
      <NativeCommandButton label={t('mobile.FilterLoadingScreen.cancel')} onPress={onCancel} />
    </SettingsSection>
  </ScrollView>;
}
