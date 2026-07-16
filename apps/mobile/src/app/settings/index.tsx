import { useRouter } from 'expo-router';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { SettingsScreen } from '../../ui/screens/SettingsScreen';
import type { SettingsDestination } from '../../ui/screens/SettingsScreenPresentation';

export default function SettingsRoute() {
  const router = useRouter();
  const { providerProfileSettingsQuery, settingsQuery } = useAppServices();
  return (
    <SettingsScreen
      providerProfileSettingsQuery={providerProfileSettingsQuery}
      settingsQuery={settingsQuery}
      onNavigate={(destination) => navigate(router, destination)}
    />
  );
}

function navigate(router: ReturnType<typeof useRouter>, destination: SettingsDestination): void {
  switch (destination) {
    case 'account': router.push('/settings/account'); return;
    case 'appearance': router.push('/settings/appearance'); return;
    case 'sharing': router.push('/settings/sharing'); return;
    case 'voice-setup': router.push('/settings/voice'); return;
    case 'connection': router.push('/settings/connection'); return;
    case 'about': router.push('/settings/about'); return;
    case 'diagnostics': router.push('/settings/diagnostics'); return;
  }
}
