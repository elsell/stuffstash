import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { DiagnosticsSettingsScreen } from '../../ui/screens/SettingsDetailScreens';

export default function DiagnosticsSettingsRoute() {
  return <DiagnosticsSettingsScreen settingsQuery={useAppServices().settingsQuery} />;
}
