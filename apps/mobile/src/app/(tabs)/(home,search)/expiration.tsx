import { useAppServices } from '../../../ui/navigation/AppServicesContext';
import { ExpirationRouteScreen } from '../../../ui/expiration/ExpirationRouteScreen';

export default function ExpirationRoute() {
  const services = useAppServices();
  return <ExpirationRouteScreen expirationWorkspaceQuery={services.expirationWorkspaceQuery} />;
}
