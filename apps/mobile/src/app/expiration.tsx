import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { useAppServices } from '../ui/navigation/AppServicesContext';
import { useMobileServerStateScope } from '../ui/navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../adapters/serverState/MobileQueryClient';
import { expirationRefreshDelay } from '../ui/serverState/expirationRefreshDelay';
import { isAccessFailure } from '../ui/serverState/isAccessFailure';
import { ExpirationWorkspaceScreen } from '../ui/expiration/ExpirationWorkspaceScreen';
import { parseExpirationRoute, expirationRouteParams, type ExpirationRouteParams } from '../ui/expiration/ExpirationRouteState';
import { assetDetailHref } from '../ui/screens/AssetDetailNavigation';

export default function ExpirationRoute() {
 const params = useLocalSearchParams<ExpirationRouteParams>();
 const { tenantId = '', inventoryId = '', filter } = parseExpirationRoute(params);
 const services = useAppServices(); const scope = useMobileServerStateScope(); const router = useRouter();
 const inventory = useQuery({ queryKey: mobileQueryKeys.inventoryScope(scope.scopeId), queryFn: ({ signal }) => scope.loadInventoryScope({ signal }), staleTime: Infinity });
 const matches = inventory.data?.tenantId === tenantId && inventory.data.inventoryId === inventoryId;
 const state = useInfiniteQuery({ queryKey: [...mobileQueryKeys.inventory(scope.scopeId, tenantId, inventoryId), 'expiration', 'list', filter], enabled: matches,
  initialPageParam: undefined as string | undefined,
  queryFn: ({ signal, pageParam }) => services.expirationWorkspaceQuery.list(tenantId, inventoryId, { ...filter, cursor: pageParam, limit: 30 }, signal),
  getNextPageParam: page => page.hasMore ? page.nextCursor ?? undefined : undefined,
  refetchInterval: query => expirationRefreshDelay({ data: query.state.data, expirationContext: { timezone: query.state.data?.pages[0]?.timezone } }, new Date(), new Date(query.state.dataUpdatedAt)), refetchIntervalInBackground: false });
 const items = matches && !isAccessFailure(state.error) ? [...new Map((state.data?.pages.flatMap(page => page.items) ?? []).map(item => [item.id, item])).values()] : [];
 const error = inventory.isError ? 'The inventory could not be loaded. Try again.' : inventory.isSuccess && !matches ? 'This expiration view belongs to another inventory. Return to Home and open it again.' : state.error ? 'Expiration could not be loaded. Refresh and try again.' : undefined;
 return <><Stack.Screen options={{ title: 'Expiration' }} /><ExpirationWorkspaceScreen mode={filter.mode} query={filter.query} refinementsActive={!!(filter.kind || filter.checkoutState || filter.typeId || filter.locationId || filter.tagIds?.length || filter.fromDate || filter.throughDate)} items={items} filtered={!!(filter.kind || filter.checkoutState || filter.query || filter.typeId || filter.locationId || filter.tagIds?.length || filter.fromDate || filter.throughDate)} loading={inventory.isPending || matches && state.isPending} refreshing={state.isRefetching && !state.isFetchingNextPage} appending={state.isFetchingNextPage} hasMore={state.hasNextPage} error={error}
  onMode={mode => router.setParams(expirationRouteParams(tenantId, inventoryId, { ...filter, mode }))} onSearch={query => router.setParams(expirationRouteParams(tenantId, inventoryId, { ...filter, query }))}
  onFilters={query => router.push({ pathname: '/expiration-filters', params: expirationRouteParams(tenantId, inventoryId, { ...filter, query }) })} onRefresh={() => { if (inventory.isError) void inventory.refetch(); else if (matches) void state.refetch(); }} onMore={() => { if (!state.isFetching) void state.fetchNextPage(); }} onOpenAsset={id => router.push(assetDetailHref(id))} /></>;
}
