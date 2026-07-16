import { useLocalSearchParams, useRouter } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { VoiceAdminGuard } from '../../../ui/navigation/VoiceAdminGuard';
import { VoiceCapabilityScreen } from '../../../ui/screens/VoiceSettingsScreens';

export default function VoiceCapabilityRoute() {
  const router = useRouter();
  const params = useLocalSearchParams<{ capability?: string | string[] }>();
  const capability = Array.isArray(params.capability) ? params.capability[0] : params.capability;
  const { manageProviderProfileCommand, providerProfileSettingsQuery, settingsQuery, testProviderProfileCommand } = useAppServices();
  return (
    <VoiceAdminGuard settingsQuery={settingsQuery}>
      <VoiceCapabilityScreen
        capability={capability ?? ''}
        manageCommand={manageProviderProfileCommand}
        onAddProfile={() => router.push('/settings/voice/profiles/add')}
        onEditCredential={(profileId) => router.push(`/settings/voice/profiles/${encodeURIComponent(profileId)}/credential`)}
        query={providerProfileSettingsQuery}
        testCommand={testProviderProfileCommand}
        onEditProfile={(profileId) => router.push(`/settings/voice/profiles/${encodeURIComponent(profileId)}`)}
      />
    </VoiceAdminGuard>
  );
}
