import { refreshExpirationHome } from '../../../ui/expiration/RefreshExpirationHome';
import { useQueryClient } from '@tanstack/react-query';
import { useMobileServerStateScope } from '../../../ui/navigation/MobileServerStateProvider';
import { ExpirationHomeEntry } from '../../../ui/expiration/ExpirationHomeEntry';
import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { NotificationHomeEntry } from '../../../ui/navigation/NotificationHomeEntry';
import { HomeScreen } from '../../../ui/screens/HomeScreen';

export default function HomeRoute() {
  const client = useQueryClient();
  const scope = useMobileServerStateScope();
  const refreshExpiration = () => refreshExpirationHome(client,scope.scopeId);
  const { assetCheckoutCommand, homeDashboardQuery } = useAppServices();

  return <NotificationHomeEntry renderAction={notificationAction => <HomeScreen onRefreshAdditional={refreshExpiration} expirationSection={<ExpirationHomeEntry />} notificationAction={notificationAction} assetCheckoutCommand={assetCheckoutCommand} dashboardQuery={homeDashboardQuery} />} />;
}
