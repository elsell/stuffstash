import { useAppServices } from '../ui/navigation/AppServicesContext';
import { SettingsScreen } from '../ui/screens/SettingsScreen';

export default function SettingsRoute() {
  const { settingsQuery } = useAppServices();

  return <SettingsScreen settingsQuery={settingsQuery} />;
}
