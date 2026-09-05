import { Pressable, Text, View } from 'react-native';
import { useSettingsListStyles } from './SettingsList';

export function SettingsRefreshNotice({ visible, onRetry }: { readonly visible: boolean; readonly onRetry: () => Promise<void> }) {
  const { styles } = useSettingsListStyles();
  if (!visible) return null;
  return <View><Text accessibilityRole="alert" style={styles.errorMessage}>Some settings could not be refreshed. Previously loaded values are shown.</Text>
    <Pressable accessibilityRole="button" onPress={() => void onRetry()} style={styles.retryButton}><Text style={styles.retryText}>Retry refresh</Text></Pressable>
  </View>;
}
