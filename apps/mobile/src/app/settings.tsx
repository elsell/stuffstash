import { useRouter } from 'expo-router';
import { useAppConnectionActions, useAppServices } from '../ui/navigation/AppServicesContext';
import { SettingsScreen } from '../ui/screens/SettingsScreen';

export default function SettingsRoute() {
  const router = useRouter();
  const { settingsQuery } = useAppServices();
  const { resetConnectionProfile } = useAppConnectionActions();

  return (
    <SettingsScreen
      settingsQuery={settingsQuery}
      onOpenProviderProfiles={() => router.push('/provider-profiles')}
      onResetConnection={resetConnectionProfile}
    />
  );
}
