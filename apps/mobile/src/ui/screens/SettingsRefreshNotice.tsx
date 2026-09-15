import { NativeCommandButton } from '../components/NativeCommandButton';
import { Text, View } from 'react-native';
import { useSettingsListStyles } from './SettingsList';

export function SettingsRefreshNotice({ visible, onRetry, message = 'Some settings could not be refreshed. Previously loaded values are shown.' }: { readonly visible: boolean; readonly onRetry: () => Promise<void>; readonly message?: string }) {
  const { styles } = useSettingsListStyles();
  if (!visible) return null;
  return <View><Text accessibilityRole="alert" style={styles.errorMessage}>{message}</Text>
    <NativeCommandButton label="Retry refresh" onPress={() => void onRetry()} />
  </View>;
}
