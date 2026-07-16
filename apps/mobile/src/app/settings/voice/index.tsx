import { useRouter } from 'expo-router';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { VoiceAdminGuard } from '../../../ui/navigation/VoiceAdminGuard';
import { VoiceSetupScreen } from '../../../ui/screens/VoiceSettingsScreens';

export default function VoiceSetupRoute() {
  const router = useRouter();
  const { providerProfileSettingsQuery, settingsQuery } = useAppServices();
  return (
    <VoiceAdminGuard settingsQuery={settingsQuery}>
      <VoiceSetupScreen
        query={providerProfileSettingsQuery}
        settingsQuery={settingsQuery}
        onOpenCapability={(capability) => router.push(`/settings/voice/${encodeURIComponent(capability)}`)}
        onOpenProfiles={() => router.push('/settings/voice/profiles')}
      />
    </VoiceAdminGuard>
  );
}
