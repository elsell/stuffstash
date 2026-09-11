import { useRouter } from 'expo-router';
import { useAppServices } from './AppServicesContext';
import { useSettingsModel } from '../screens/SettingsScreenState';
import { NotificationBell } from '../components/NotificationBell';

export function NotificationHomeEntry() {
  const services = useAppServices();
  const router = useRouter();
  const model = useSettingsModel(services.settingsQuery);
  if (model.state.status !== 'ready') return null;
  const { selectedTenant, selectedInventory } = model.state.settings;
  return <NotificationBell key={JSON.stringify([services.serviceScopeId, selectedTenant.id, selectedInventory.id])} tenantId={selectedTenant.id} inventoryId={selectedInventory.id}
    initialize={async (signal) => { await services.createNotificationPreferencesSession(selectedTenant.id, selectedInventory.id).initialize(Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC', { signal }); }}
    count={(signal) => services.notificationInboxQueries.count(selectedTenant.id, selectedInventory.id, { signal })}
    onOpen={() => router.push('/notifications')} />;
}
