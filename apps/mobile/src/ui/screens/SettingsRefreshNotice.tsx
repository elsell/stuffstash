import { t } from '../../presentation/localization';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { Text, View } from 'react-native';
import { useSettingsListStyles } from './SettingsList';

export function SettingsRefreshNotice({ visible, onRetry, message = t('mobile.SettingsRefreshNotice.someSettingsCouldNotBeRefreshedPreviouslyLoadedValues') }: { readonly visible: boolean; readonly onRetry: () => Promise<void>; readonly message?: string }) {
  const { styles } = useSettingsListStyles();
  if (!visible) return null;
  return <View><Text accessibilityRole="alert" style={styles.errorMessage}>{message}</Text>
    <NativeCommandButton label={t('mobile.SettingsRefreshNotice.retryRefresh')} onPress={() => void onRetry()} />
  </View>;
}
