import { useLocalSearchParams, useRouter } from 'expo-router';
import { useAppServices } from '../../../../../ui/navigation/AppServicesContext';
import { VoiceAdminGuard } from '../../../../../ui/navigation/VoiceAdminGuard';
import { ProviderCredentialScreen } from '../../../../../ui/screens/VoiceSettingsScreens';

export default function ProviderCredentialRoute() {
  const router = useRouter();
  const params = useLocalSearchParams<{ providerProfileId?: string | string[] }>();
  const profileId = Array.isArray(params.providerProfileId) ? params.providerProfileId[0] : params.providerProfileId;
  const { manageProviderProfileCommand, providerProfileSettingsQuery, settingsQuery } = useAppServices();
  return (
    <VoiceAdminGuard settingsQuery={settingsQuery}>
      <ProviderCredentialScreen manageCommand={manageProviderProfileCommand} profileId={profileId ?? ''} query={providerProfileSettingsQuery} onCancel={() => router.back()} onSaved={() => router.back()} />
    </VoiceAdminGuard>
  );
}
