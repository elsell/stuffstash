import { useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { useMobileServerStateScopeId } from '../ui/navigation/MobileServerStateProvider';
import { useSettingsModel } from '../ui/screens/SettingsScreenState';
import { useSettingsListStyles } from '../ui/screens/SettingsList';
import { NotificationInboxScreen } from '../ui/screens/NotificationInboxScreen';
import { mobileQueryKeys } from '../adapters/serverState/MobileQueryClient';
import { assetDetailHref } from '../ui/screens/AssetDetailNavigation';

export default function NotificationInboxRoute() {
  const services = useAppServices();
  const scopeId = useMobileServerStateScopeId();
  const model = useSettingsModel(services.settingsQuery);
  const client = useQueryClient();
  const router = useRouter();
  const { styles, palette } = useSettingsListStyles();
  if (model.state.status === 'loading') return <View style={styles.shell}><ActivityIndicator accessibilityLabel="Loading inventory" color={palette.action} /></View>;
  if (model.state.status === 'error') return <View style={[styles.shell, styles.errorContainer]}><Text accessibilityRole="alert" style={styles.errorMessage}>{model.state.message}</Text><Pressable accessibilityRole="button" onPress={() => void model.load()} style={styles.retryButton}><Text style={styles.retryText}>Retry</Text></Pressable></View>;
  const { selectedTenant, selectedInventory } = model.state.settings;
  return <NotificationInboxScreen key={JSON.stringify([scopeId, selectedTenant.id, selectedInventory.id])} tenantId={selectedTenant.id} inventoryId={selectedInventory.id} queries={services.notificationInboxQueries}
    onOpenAsset={(assetId) => router.push(assetDetailHref(assetId))}
    onSettings={() => router.push('/settings/inventory/notifications')}
    onChanged={() => { void client.invalidateQueries({ queryKey: mobileQueryKeys.notificationCount(scopeId, selectedTenant.id, selectedInventory.id) }); }} />;
}
