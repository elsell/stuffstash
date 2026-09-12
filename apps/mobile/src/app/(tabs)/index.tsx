import { refreshExpirationHome } from '../../ui/expiration/RefreshExpirationHome';
import { useQueryClient, useIsFetching } from '@tanstack/react-query';
import { useMobileServerStateScope } from '../../ui/navigation/MobileServerStateProvider';
import { ExpirationHomeEntry } from '../../ui/expiration/ExpirationHomeEntry';
import { useAppServices } from '../../ui/navigation/AppServicesContext';
import { NotificationHomeEntry } from '../../ui/navigation/NotificationHomeEntry';
import { HomeScreen } from '../../ui/screens/HomeScreen';

export default function HomeRoute() {
  const client = useQueryClient();
  const scope = useMobileServerStateScope();
  const refreshingExpiration = useIsFetching({predicate:query => query.queryKey[1] === scope.scopeId && query.queryKey[6] === 'expiration'});
  const refreshExpiration = () => refreshExpirationHome(client,scope.scopeId);
  const { assetCheckoutCommand, homeDashboardQuery } = useAppServices();

  return <HomeScreen refreshingAdditional={refreshingExpiration > 0} onRefreshAdditional={refreshExpiration} expirationSection={<ExpirationHomeEntry />} notificationAction={<NotificationHomeEntry />} assetCheckoutCommand={assetCheckoutCommand} dashboardQuery={homeDashboardQuery} />;
}
