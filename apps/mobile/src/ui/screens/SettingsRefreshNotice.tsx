import { NativeCommandButton } from '../components/NativeCommandButton';
import { Text, View } from 'react-native';
import { useSettingsListStyles } from './SettingsList';

export function SettingsRefreshNotice({ visible, onRetry }: { readonly visible: boolean; readonly onRetry: () => Promise<void> }) {
  const { styles } = useSettingsListStyles();
  if (!visible) return null;
  return <View><Text accessibilityRole="alert" style={styles.errorMessage}>Some settings could not be refreshed. Previously loaded values are shown.</Text>
    <NativeCommandButton label="Retry refresh" onPress={() => void onRetry()} />
  </View>;
}
