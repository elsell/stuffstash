import { NativeCommandButton } from '../components/NativeCommandButton';
import { useMemo, useState } from 'react';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import { isAccessFailure } from '../serverState/isAccessFailure';
import { useInfiniteQuery, useQuery } from '@tanstack/react-query';
import { router, Stack } from 'expo-router';
import { Text } from 'react-native';
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
  const coreIdentity = JSON.stringify(core.resourceKey);
  const [coreAccess, setCoreAccess] = useState({ identity: coreIdentity, denied: false });
  // Query retries clear error before returning data. Retain a denial until this
  // asset's core read succeeds, without carrying it into another query scope.
  const coreAccessDenied = isAccessFailure(core.error)
    || (coreAccess.identity === coreIdentity && coreAccess.denied && !core.isSuccess);
  if (coreAccess.identity !== coreIdentity || coreAccess.denied !== coreAccessDenied) {
    setCoreAccess({ identity: coreIdentity, denied: coreAccessDenied });
  }
  const assetTitle = core.data?.view.title ?? 'Asset';
  const accessDenied = isAccessFailure(history.error) || isAccessFailure(inventory.error) || coreAccessDenied;
  const first = accessDenied ? undefined : history.data?.pages[0];
  const state: AssetCheckoutHistorySheetState = first
    ? { status: 'ready', assetTitle, history: { ...first, records: history.data!.pages.flatMap((page) => page.records), hasMore: history.hasNextPage } }
    : accessDenied || history.isError || inventory.isError ? { status: 'error', assetTitle, message: 'Checkout history could not be loaded.' }
      : { status: 'loading', assetTitle };
  const retry = () => { void (inventory.isError ? inventory.refetch() : coreAccessDenied ? core.refetch() : history.refetch()); };
  const actionOptions = useNativeHeaderActionOptions([{ kind: 'close', label: 'Close', onPress: () => router.back() }]);
  const headerOptions = useMemo(() => ({ title: 'Checkout history', headerShown: true, ...actionOptions }), [actionOptions]);
  return <>
    <Stack.Screen options={headerOptions} />
    <AssetCheckoutHistorySheet state={state} footer={<>
      {!accessDenied && core.isError && !core.data ? <>
        <Text accessibilityRole="alert" style={{ color: palette.danger }}>Asset name could not be loaded.</Text>
        <NativeCommandButton label={core.isFetching ? 'Loading asset name…' : 'Try loading asset name again'}
          disabled={core.isFetching} onPress={() => { if (!core.isFetching) void core.refetch(); }} />
      </> : null}
      {state.status === 'error' ? <NativeCommandButton label="Try again" onPress={retry} /> : null}
      {history.isRefetchError && state.status === 'ready' ? <>
        <Text accessibilityRole="alert" style={{ color: palette.danger }}>Checkout history could not be refreshed. Previously loaded checkouts are shown.</Text>
        <NativeCommandButton label="Try refreshing again" onPress={retry} />
      </> : null}
      {history.isFetchNextPageError ? <Text accessibilityRole="alert" style={{ color: palette.danger }}>Older checkouts could not be loaded.</Text> : null}
      {state.status === 'ready' && history.hasNextPage ? <NativeCommandButton disabled={history.isFetching} onPress={() => { if (!history.isFetching) void history.fetchNextPage(); }}
        label={history.isFetchingNextPage ? 'Loading older checkouts…' : history.isFetchNextPageError ? 'Try older checkouts again' : 'Load older checkouts'} /> : null}
    </>} />
  </>;
}
