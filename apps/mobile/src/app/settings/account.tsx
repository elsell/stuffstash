import { useAppConnectionActions, useAppServices } from '../../ui/navigation/AppServicesContext';
import { AccountSettingsScreen } from '../../ui/screens/SettingsDetailScreens';

export default function AccountSettingsRoute() {
  const { settingsQuery } = useAppServices();
  const { signOut } = useAppConnectionActions();
  return <AccountSettingsScreen onSignOut={signOut} settingsQuery={settingsQuery} />;
}
