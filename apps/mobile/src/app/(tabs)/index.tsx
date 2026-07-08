import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { HomeScreen } from '../../ui/screens/HomeScreen';

export default function HomeRoute() {
  const { assetCheckoutCommand, homeDashboardQuery } = useAppServices();

  return <HomeScreen assetCheckoutCommand={assetCheckoutCommand} dashboardQuery={homeDashboardQuery} />;
}
