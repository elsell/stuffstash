import { useLocalSearchParams, useRouter } from 'expo-router';
import { useAppServices } from '../../../../../ui/navigation/AppServicesContext';
import { VoiceAdminGuard } from '../../../../../ui/navigation/VoiceAdminGuard';
import { ProviderProfileDetailScreen } from '../../../../../ui/screens/VoiceSettingsScreens';

export default function ProviderProfileDetailRoute() {
  const router = useRouter();
  const params = useLocalSearchParams<{ providerProfileId?: string | string[] }>();
  const profileId = Array.isArray(params.providerProfileId) ? params.providerProfileId[0] : params.providerProfileId;
  const { manageProviderProfileCommand, providerProfileSettingsQuery, settingsQuery, testProviderProfileCommand } = useAppServices();
  return (
    <VoiceAdminGuard settingsQuery={settingsQuery}>
      <ProviderProfileDetailScreen
        manageCommand={manageProviderProfileCommand}
        profileId={profileId ?? ''}
        query={providerProfileSettingsQuery}
        testCommand={testProviderProfileCommand}
        onEditCredential={() => router.push(`/settings/voice/profiles/${encodeURIComponent(profileId ?? '')}/credential`)}
        onEditPrompt={() => router.push(`/settings/voice/profiles/${encodeURIComponent(profileId ?? '')}/prompt`)}
      />
    </VoiceAdminGuard>
  );
}
