import { isAccessFailure } from '../serverState/isAccessFailure';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { router, Stack } from 'expo-router';
import { Pressable, Text } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { AssetCoreQuery } from '../../application/assets/AssetCoreQuery';
import type { AssetCheckoutHistoryQuery } from '../../application/assets/AssetCheckoutHistoryQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { AssetCheckoutHistorySheet, type AssetCheckoutHistorySheetState } from './AssetCheckoutHistorySheet';
import { useAppearancePalette } from '../theme/AppearanceContext';

export function AssetCheckoutHistorySheetRouteScreen({ assetCheckoutHistoryQuery, assetCoreQuery, assetId }: {
  readonly assetCheckoutHistoryQuery: Pick<AssetCheckoutHistoryQuery, 'execute'>;
  readonly assetCoreQuery: Pick<AssetCoreQuery, 'execute'>;
  readonly assetId: string;
}) {
  const palette = useAppearancePalette();
  const scope = useMobileServerStateScope();
  const inventory = useQuery({ queryKey: mobileQueryKeys.inventoryScope(scope.scopeId), queryFn: ({ signal }) => scope.loadInventoryScope({ signal }), staleTime: Infinity });
  const core = useMobileInventoryServerQuery({
    key: (service, tenant, selected) => mobileQueryKeys.assetCore(service, tenant, selected, assetId),
    query: (signal) => assetCoreQuery.execute(assetId, { signal })
  });
  const history = useInfiniteQuery({
    queryKey: mobileQueryKeys.assetCheckouts(scope.scopeId, inventory.data?.tenantId ?? 'pending', inventory.data?.inventoryId ?? 'pending', assetId),
    enabled: inventory.isSuccess,
    subscribed: inventory.isSuccess,
    initialPageParam: undefined as string | undefined,
    queryFn: ({ signal, pageParam }) => assetCheckoutHistoryQuery.execute({ assetId, limit: 20, cursor: pageParam, signal }),
    getNextPageParam: (page) => page.hasMore ? page.nextCursor : undefined
  });
  const assetTitle = core.data?.view.title ?? 'Asset';
  const accessDenied = isAccessFailure(history.error) || isAccessFailure(inventory.error) || (core.isError && !core.data);
  const first = accessDenied ? undefined : history.data?.pages[0];
  const state: AssetCheckoutHistorySheetState = first
    ? { status: 'ready', assetTitle, history: { ...first, records: history.data!.pages.flatMap((page) => page.records), hasMore: history.hasNextPage } }
    : accessDenied || history.isError || inventory.isError ? { status: 'error', assetTitle, message: 'Checkout history could not be loaded.' }
      : { status: 'loading', assetTitle };
  const retry = () => { void (inventory.isError ? inventory.refetch() : (core.isError && !core.data) ? core.refetch() : history.refetch()); };
  return <SafeAreaView style={{ flex: 1, backgroundColor: palette.surface }} edges={['left', 'right', 'bottom']}>
    <Stack.Screen options={{ title: 'Checkout history' }} />
    <AssetCheckoutHistorySheet state={state} onClose={() => router.back()} footer={<>
      {state.status === 'error' ? <Pressable accessibilityRole="button" onPress={retry}><Text style={{ color: palette.action, padding: 16 }}>Try again</Text></Pressable> : null}
      {history.isRefetchError && state.status === 'ready' ? <>
        <Text accessibilityRole="alert" style={{ color: palette.danger }}>Checkout history could not be refreshed. Previously loaded checkouts are shown.</Text>
        <Pressable accessibilityRole="button" onPress={retry}><Text style={{ color: palette.action, padding: 16 }}>Try refreshing again</Text></Pressable>
      </> : null}
      {history.isFetchNextPageError ? <Text accessibilityRole="alert" style={{ color: palette.danger }}>Older checkouts could not be loaded.</Text> : null}
      {state.status === 'ready' && history.hasNextPage ? <Pressable accessibilityLabel="Load older checkouts" accessibilityRole="button" disabled={history.isFetching} onPress={() => { if (!history.isFetching) void history.fetchNextPage(); }}>
        <Text style={{ color: palette.action, padding: 16 }}>{history.isFetchingNextPage ? 'Loading older checkouts…' : history.isFetchNextPageError ? 'Try older checkouts again' : 'Load older checkouts'}</Text>
      </Pressable> : null}
    </>} />
  </SafeAreaView>;
}
