import { useLocalSearchParams } from 'expo-router';
import { ScrollView, Text, View } from 'react-native';
import NotificationSettingsRoute from './notifications';
import { useSettingsListStyles } from '../../../ui/screens/SettingsList';
import { notificationSettingsEditorTarget } from '../../../ui/presentation/NotificationSettingsDestination';
export default function NotificationEditorRoute() {
  const {view,typeId,tenantId,inventoryId}=useLocalSearchParams<{view?:string;typeId?:string;tenantId?:string;inventoryId?:string}>();
  const {styles}=useSettingsListStyles();
  const target=notificationSettingsEditorTarget({view,typeId,tenantId,inventoryId});
  if(!target)return <ScrollView style={styles.shell} contentContainerStyle={{ flexGrow: 1 }} contentInsetAdjustmentBehavior="automatic"><Text style={styles.errorMessage}>This settings link is unavailable.</Text></ScrollView>;
  return <NotificationSettingsRoute page={target.page} expectedScope={target.scope} />;
}
