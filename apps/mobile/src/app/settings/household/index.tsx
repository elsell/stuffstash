import { useRouter } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { HouseholdSettingsScreen } from '../../../ui/screens/ScopedSettingsScreens';
import { useEffect } from 'react';

export default function HouseholdSettingsRoute() {
  const router = useRouter();
  const services = useAppServices();
  useEffect(() => services.customizationObservability.record({ name: 'settings.level_selected', scope: 'tenant' }), [services.customizationObservability]);
  return <HouseholdSettingsScreen settingsQuery={services.settingsQuery} onNavigate={(destination) => {
    if (destination === 'voice') router.push('/settings/voice');
    else router.push(`/settings/household/${destination}` as never);
  }} />;
}
