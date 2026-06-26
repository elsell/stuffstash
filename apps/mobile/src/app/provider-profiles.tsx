import { useAppServices } from '../ui/navigation/AppServicesContext';
import { ProviderProfilesScreen } from '../ui/screens/ProviderProfilesScreen';

export default function ProviderProfilesRoute() {
  const {
    manageProviderProfileCommand,
    providerProfileSettingsQuery,
    testProviderProfileCommand
  } = useAppServices();

  return (
    <ProviderProfilesScreen
      manageCommand={manageProviderProfileCommand}
      query={providerProfileSettingsQuery}
      testCommand={testProviderProfileCommand}
    />
  );
}
