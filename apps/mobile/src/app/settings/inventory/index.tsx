import { useRouter } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { InventorySettingsScreen } from '../../../ui/screens/ScopedSettingsScreens';
import { useEffect } from 'react';

export default function InventorySettingsRoute() {
  const router = useRouter();
  const services = useAppServices();
  useEffect(() => services.customizationObservability.record({ name: 'settings.level_selected', scope: 'inventory' }), [services.customizationObservability]);
  return <InventorySettingsScreen settingsQuery={services.settingsQuery} onNavigate={(destination) => {
    if (destination === 'sharing') router.push('/settings/sharing');
    else router.push(`/settings/inventory/${destination}` as never);
  }} />;
}
