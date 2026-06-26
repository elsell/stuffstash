import { useAppConnectionActions, useAppServices } from '../ui/navigation/AppServicesContext';
import { SettingsScreen } from '../ui/screens/SettingsScreen';

export default function SettingsRoute() {
  const { settingsQuery } = useAppServices();
  const { resetConnectionProfile } = useAppConnectionActions();

  return <SettingsScreen settingsQuery={settingsQuery} onResetConnection={resetConnectionProfile} />;
}
