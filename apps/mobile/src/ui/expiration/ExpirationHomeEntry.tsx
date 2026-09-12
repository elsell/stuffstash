import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useAppServices } from '../navigation/AppServicesContext';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { expirationRefreshDelay } from '../serverState/expirationRefreshDelay';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { ExpirationHomeSection } from './ExpirationHomeSection';
import { assetDetailHref } from '../screens/AssetDetailNavigation';

export function ExpirationHomeEntry() {
 const services = useAppServices();
 const scope = useMobileServerStateScope();
 const router = useRouter();
 const inventory = useQuery({ queryKey: mobileQueryKeys.inventoryScope(scope.scopeId), queryFn: ({ signal }) => scope.loadInventoryScope({ signal }), staleTime: Infinity });
 const tenantId = inventory.data?.tenantId ?? '';
 const inventoryId = inventory.data?.inventoryId ?? '';
 const state = useQuery({ queryKey: [...mobileQueryKeys.inventory(scope.scopeId, tenantId, inventoryId), 'expiration', 'home'], enabled: !!inventory.data, queryFn: ({ signal }) => services.expirationWorkspaceQuery.home(tenantId, inventoryId, signal), refetchInterval: query => expirationRefreshDelay({ data: query.state.data, expirationContext: { timezone: query.state.data?.timezone } }, new Date(), new Date(query.state.dataUpdatedAt)), refetchIntervalInBackground: false });
 return <ExpirationHomeSection data={isAccessFailure(state.error) ? undefined : state.data} error={inventory.error || state.error ? 'Expiration could not be refreshed. Try again.' : undefined} onRetry={() => { if (inventory.isError) void inventory.refetch(); else void state.refetch(); }} onOpen={mode => router.push({ pathname: '/expiration', params: { tenantId, inventoryId, mode } })} onOpenAsset={id => router.push(assetDetailHref(id))} />;
}
