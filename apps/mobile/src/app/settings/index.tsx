import { useRouter } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { SettingsScreen } from '../../ui/screens/SettingsScreen';
import type { SettingsDestination } from '../../ui/screens/SettingsScreenPresentation';
import { useEffect } from 'react';

export default function SettingsRoute() {
  const router = useRouter();
  const { customizationObservability, settingsQuery } = useAppServices();
  useEffect(() => customizationObservability.record({ name: 'settings.opened' }), [customizationObservability]);
  return (
    <SettingsScreen
      settingsQuery={settingsQuery}
      onNavigate={(destination) => navigate(router, destination)}
    />
  );
}

function navigate(router: ReturnType<typeof useRouter>, destination: SettingsDestination): void {
  switch (destination) {
    case 'account': router.push('/settings/account'); return;
    case 'appearance': router.push('/settings/appearance'); return;
    case 'tenant-settings': router.push('/settings/household'); return;
    case 'inventory-settings': router.push('/settings/inventory'); return;
    case 'connection': router.push('/settings/connection'); return;
    case 'about': router.push('/settings/about'); return;
    case 'diagnostics': router.push('/settings/diagnostics'); return;
  }
}
