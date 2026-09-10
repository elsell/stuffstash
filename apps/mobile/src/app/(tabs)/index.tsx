import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { NotificationHomeEntry } from '../../ui/navigation/NotificationHomeEntry';
import { HomeScreen } from '../../ui/screens/HomeScreen';

export default function HomeRoute() {
  const { assetCheckoutCommand, homeDashboardQuery } = useAppServices();

  return <HomeScreen notificationAction={<NotificationHomeEntry />} assetCheckoutCommand={assetCheckoutCommand} dashboardQuery={homeDashboardQuery} />;
}
