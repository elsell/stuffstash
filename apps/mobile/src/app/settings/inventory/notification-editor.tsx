import { useLocalSearchParams } from 'expo-router';
import { Text, View } from 'react-native';
import NotificationSettingsRoute from './notifications';
import { useSettingsListStyles } from '../../../ui/screens/SettingsList';
import { notificationSettingsEditorTarget } from '../../../ui/presentation/NotificationSettingsDestination';
export default function NotificationEditorRoute() {
  const {view,typeId,tenantId,inventoryId}=useLocalSearchParams<{view?:string;typeId?:string;tenantId?:string;inventoryId?:string}>();
  const {styles}=useSettingsListStyles();
  const target=notificationSettingsEditorTarget({view,typeId,tenantId,inventoryId});
  if(!target)return <View style={styles.shell}><Text style={styles.errorMessage}>This settings link is unavailable.</Text></View>;
  return <NotificationSettingsRoute page={target.page} expectedScope={target.scope} />;
}
