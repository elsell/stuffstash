import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { AboutSettingsScreen } from '../../ui/screens/SettingsDetailScreens';

export default function AboutSettingsRoute() {
  return <AboutSettingsScreen settingsQuery={useAppServices().settingsQuery} />;
}
