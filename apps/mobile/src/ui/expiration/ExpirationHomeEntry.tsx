import { useRouter } from 'expo-router';
import { useAppServices } from '../navigation/AppServicesContext';
import { assetDetailHref } from '../screens/AssetDetailNavigation';
import { ExpirationHomeContent } from './ExpirationHomeContent';

export function ExpirationHomeEntry() {
 const services = useAppServices();
 const router = useRouter();
 return <ExpirationHomeContent query={services.expirationWorkspaceQuery}
  onOpen={(tenantId, inventoryId, mode) => router.push({ pathname: '/expiration', params: { tenantId, inventoryId, mode } })}
  onOpenAsset={id => router.push(assetDetailHref(id))} />;
}
