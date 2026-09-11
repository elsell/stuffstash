import { useMemo } from 'react';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import type { MobileComposition } from '../../../bootstrap/mobileComposition';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { useMobileServerStateScopeId } from '../../../ui/navigation/MobileServerStateProvider';
import { useSettingsModel } from '../../../ui/screens/SettingsScreenState';
import { NotificationSettingsScreen } from '../../../ui/screens/NotificationSettingsScreen';
import { useSettingsListStyles } from '../../../ui/screens/SettingsList';

export default function NotificationSettingsRoute() {
  const services = useAppServices();
  const scopeId = useMobileServerStateScopeId();
  const model = useSettingsModel(services.settingsQuery);
  const { styles, palette } = useSettingsListStyles();
  if (model.state.status === 'loading') return <View style={styles.shell}><ActivityIndicator accessibilityLabel="Loading inventory" color={palette.action} /></View>;
  if (model.state.status === 'error') return <View style={[styles.shell, styles.errorContainer]}>
    <Text accessibilityRole="alert" style={styles.errorMessage}>{model.state.message}</Text>
    <Pressable accessibilityRole="button" onPress={() => void model.load()} style={styles.retryButton}><Text style={styles.retryText}>Retry</Text></Pressable>
  </View>;
  const { selectedTenant, selectedInventory } = model.state.settings;
  return <ScopedNotifications key={JSON.stringify([scopeId, selectedTenant.id, selectedInventory.id])} services={services} tenantId={selectedTenant.id} inventoryId={selectedInventory.id} />;
}

function ScopedNotifications({ services, tenantId, inventoryId }: { readonly services: MobileComposition; readonly tenantId: string; readonly inventoryId: string }) {
  const session = useMemo(() => services.createNotificationPreferencesSession(tenantId, inventoryId), [services, tenantId, inventoryId]);
  return <NotificationSettingsScreen tenantId={tenantId} inventoryId={inventoryId} session={session} assetTypesQuery={services.inventoryAssetTypesQuery} pushSession={services.pushSession} />;
}
