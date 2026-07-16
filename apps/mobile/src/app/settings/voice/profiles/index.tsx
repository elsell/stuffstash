import { useRouter } from 'expo-router';
import { useAppServices } from '../../../../ui/navigation/AppServicesContext';
import { VoiceAdminGuard } from '../../../../ui/navigation/VoiceAdminGuard';
import { ProviderProfileListScreen } from '../../../../ui/screens/VoiceSettingsScreens';

export default function ProviderProfileListRoute() {
  const router = useRouter();
  const { providerProfileSettingsQuery, settingsQuery } = useAppServices();
  return (
    <VoiceAdminGuard settingsQuery={settingsQuery}>
      <ProviderProfileListScreen
        query={providerProfileSettingsQuery}
        onAdd={() => router.push('/settings/voice/profiles/add')}
        onOpenProfile={(profileId) => router.push(`/settings/voice/profiles/${encodeURIComponent(profileId)}`)}
      />
    </VoiceAdminGuard>
  );
}
