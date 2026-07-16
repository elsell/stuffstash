import { useRouter } from 'expo-router';
import { useAppServices } from '../../../../ui/navigation/AppServicesContext';
import { VoiceAdminGuard } from '../../../../ui/navigation/VoiceAdminGuard';
import { AddProviderProfileScreen } from '../../../../ui/screens/VoiceSettingsScreens';

export default function AddProviderProfileRoute() {
  const router = useRouter();
  const { manageProviderProfileCommand, settingsQuery } = useAppServices();
  return (
    <VoiceAdminGuard settingsQuery={settingsQuery}>
      <AddProviderProfileScreen
        manageCommand={manageProviderProfileCommand}
        onCreated={(profileId) => router.replace(`/settings/voice/profiles/${encodeURIComponent(profileId)}`)}
      />
    </VoiceAdminGuard>
  );
}
