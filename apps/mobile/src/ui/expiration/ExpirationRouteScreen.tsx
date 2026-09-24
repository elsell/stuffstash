import { expirationOriginTab } from './ExpirationTabReturn';
import { usePullRefresh } from '../serverState/usePullRefresh';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { Stack, useLocalSearchParams, useRouter, useSegments } from 'expo-router';
import type { ExpirationWorkspaceQuery } from '../../application/expiration/ExpirationWorkspaceQuery';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { expirationRefreshDelay } from '../serverState/expirationRefreshDelay';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { ExpirationWorkspaceScreen } from '../expiration/ExpirationWorkspaceScreen';
import { parseExpirationRoute, expirationRouteParams, type ExpirationRouteParams } from '../expiration/ExpirationRouteState';
import { assetDetailHref } from '../screens/AssetDetailNavigation';

export function ExpirationRouteScreen({ expirationWorkspaceQuery }: { readonly expirationWorkspaceQuery: Pick<ExpirationWorkspaceQuery, 'list'> }) {
 const originTab = expirationOriginTab(useSegments());
 const params = useLocalSearchParams<ExpirationRouteParams>();
 const { tenantId = '', inventoryId = '', filter } = parseExpirationRoute(params);
 const scope = useMobileServerStateScope(); const router = useRouter();
 const inventory = useQuery({ queryKey: mobileQueryKeys.inventoryScope(scope.scopeId), queryFn: ({ signal }) => scope.loadInventoryScope({ signal }), staleTime: Infinity });
 const matches = inventory.data?.tenantId === tenantId && inventory.data.inventoryId === inventoryId;
 const state = useInfiniteQuery({ queryKey: [...mobileQueryKeys.inventory(scope.scopeId, tenantId, inventoryId), 'expiration', 'list', filter], enabled: matches,
  initialPageParam: undefined as string | undefined,
  queryFn: ({ signal, pageParam }) => expirationWorkspaceQuery.list(tenantId, inventoryId, { ...filter, cursor: pageParam, limit: 30 }, signal),
  getNextPageParam: page => page.hasMore ? page.nextCursor ?? undefined : undefined,
  refetchInterval: query => expirationRefreshDelay({ data: query.state.data, expirationContext: { timezone: query.state.data?.pages[0]?.timezone } }, new Date(), new Date(query.state.dataUpdatedAt)), refetchIntervalInBackground: false });
 const retry = async () => { if (inventory.isError) await inventory.refetch(); else if (matches) await state.refetch(); };
 const pullRefresh = usePullRefresh(retry);
 const mismatch = inventory.isSuccess && !matches;
 const retrying = inventory.isFetching || state.isFetching;
 const recovery = mismatch
  ? { label: 'Return to Home', onPress: () => router.dismissTo('/(tabs)/(home)') }
  : { label: retrying ? 'Retrying expiration' : 'Retry expiration', disabled: retrying, onPress: () => { if (!retrying) void retry(); } };
 const items = matches && !isAccessFailure(state.error) ? [...new Map((state.data?.pages.flatMap(page => page.items) ?? []).map(item => [item.id, item])).values()] : [];
 const error = inventory.isError ? 'The inventory could not be loaded. Try again.' : inventory.isSuccess && !matches ? 'This expiration view belongs to another inventory. Return to Home and open it again.' : state.error ? 'Expiration could not be loaded. Refresh and try again.' : undefined;
 return <><Stack.Screen options={{ title: 'Expiration' }} /><ExpirationWorkspaceScreen mode={filter.mode} query={filter.query} refinementsActive={!!(filter.kind || filter.checkoutState || filter.typeId || filter.locationId || filter.tagIds?.length || filter.fromDate || filter.throughDate)} items={items} filtered={!!(filter.kind || filter.checkoutState || filter.query || filter.typeId || filter.locationId || filter.tagIds?.length || filter.fromDate || filter.throughDate)} loading={inventory.isPending || matches && state.isPending} refreshing={pullRefresh.refreshing} appending={state.isFetchingNextPage} hasMore={state.hasNextPage} error={error} recovery={recovery}
  onMode={mode => router.setParams(expirationRouteParams(tenantId, inventoryId, { ...filter, mode }))} onSearch={query => router.setParams(expirationRouteParams(tenantId, inventoryId, { ...filter, query }))}
  onFilters={query => router.push({ pathname: '/expiration-filters', params: { ...expirationRouteParams(tenantId, inventoryId, { ...filter, query }), originTab } })} onRefresh={() => { void pullRefresh.refresh(); }} onMore={() => { if (!state.isFetching) void state.fetchNextPage(); }} onOpenAsset={id => router.push(assetDetailHref(id))} /></>;
}
