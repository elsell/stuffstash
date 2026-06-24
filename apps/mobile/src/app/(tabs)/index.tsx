import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { HomeScreen } from '../../ui/screens/HomeScreen';

export default function HomeRoute() {
  const { homeDashboardQuery } = useAppServices();

  return <HomeScreen dashboardQuery={homeDashboardQuery} />;
}
