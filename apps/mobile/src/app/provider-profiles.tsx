import { useAppServices } from '../ui/navigation/AppServicesContext';
import { ProviderProfilesScreen } from '../ui/screens/ProviderProfilesScreen';

export default function ProviderProfilesRoute() {
  const { providerProfileSettingsQuery, testProviderProfileCommand } = useAppServices();

  return (
    <ProviderProfilesScreen
      query={providerProfileSettingsQuery}
      testCommand={testProviderProfileCommand}
    />
  );
}
