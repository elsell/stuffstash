import { useAppConnectionActions, useAppServices } from '../../ui/navigation/AppServicesContext';
import { ConnectionSettingsScreen } from '../../ui/screens/SettingsDetailScreens';

export default function ConnectionSettingsRoute() {
  const { settingsQuery } = useAppServices();
  const { changeServer } = useAppConnectionActions();
  return <ConnectionSettingsScreen onChangeServer={changeServer} settingsQuery={settingsQuery} />;
}
