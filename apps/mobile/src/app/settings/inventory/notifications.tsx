import { useQueryClient } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../../adapters/serverState/MobileQueryClient';
import { useMemo } from 'react';
import { router } from 'expo-router';
import type { NotificationSettingsPage } from '../../../ui/presentation/NotificationSettingsDestination';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import type { MobileComposition } from '../../../bootstrap/mobileComposition';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { useMobileServerStateScopeId } from '../../../ui/navigation/MobileServerStateProvider';
import { useSettingsModel } from '../../../ui/screens/SettingsScreenState';
import { NotificationSettingsScreen } from '../../../ui/screens/NotificationSettingsScreen';
import { useSettingsListStyles } from '../../../ui/screens/SettingsList';

export default function NotificationSettingsRoute({ page, expectedScope }: { readonly page?: NotificationSettingsPage; readonly expectedScope?: {tenantId:string;inventoryId:string} } = {}) {
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
  if (expectedScope && (expectedScope.tenantId !== selectedTenant.id || expectedScope.inventoryId !== selectedInventory.id)) return <View style={styles.shell}><Text style={styles.errorMessage}>This inventory is no longer selected. Go back to open its settings again.</Text></View>;
  return <ScopedNotifications page={page} key={JSON.stringify([scopeId, selectedTenant.id, selectedInventory.id])} services={services} tenantId={selectedTenant.id} inventoryId={selectedInventory.id} />;
}

function ScopedNotifications({ services, tenantId, inventoryId, page }: { readonly page?: NotificationSettingsPage; readonly services: MobileComposition; readonly tenantId: string; readonly inventoryId: string }) {
  const client = useQueryClient();
  const scopeId = useMobileServerStateScopeId();
  const session = useMemo(() => services.createNotificationPreferencesSession(tenantId, inventoryId), [services, tenantId, inventoryId]);
  return <NotificationSettingsScreen page={page} onBack={() => router.back()} onNavigate={destination => router.push({pathname:'/settings/inventory/notification-editor',params:{view:destination.kind,typeId:'typeId' in destination ? destination.typeId : undefined,tenantId,inventoryId}})} tenantId={tenantId} inventoryId={inventoryId} session={session} onChanged={() => { void client.invalidateQueries({ queryKey: mobileQueryKeys.inventory(scopeId, tenantId, inventoryId) }); }} assetTypesQuery={services.inventoryAssetTypesQuery} pushSession={services.pushSession} />;
}
