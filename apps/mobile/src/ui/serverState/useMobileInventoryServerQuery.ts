import { isAccessFailure } from './isAccessFailure';
import { useQuery, useQueryClient, type QueryKey, type RefetchOptions, type UseQueryResult } from '@tanstack/react-query';
import { fetchMobileInventoryServerQuery, readScopedMobileResource } from './fetchMobileInventoryServerQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';

type MobileInventoryServerQueryOptions<TData> = {
  readonly key: (scopeId: string, tenantId: string, inventoryId: string) => QueryKey;
  readonly query: (signal?: AbortSignal) => Promise<TData>;
  readonly enabled?: boolean;
};

type MobileInventoryServerQueryResult<TData> = Omit<UseQueryResult<TData, Error>, 'refetch'> & {
  reconcile: () => Promise<TData>;
  resourceKey: QueryKey;
  refetch: (options?: RefetchOptions) => Promise<{ data: TData | undefined; error: Error | null }>;
};

export function useMobileInventoryServerQuery<TData>({
  key,
  query,
  enabled = true
}: MobileInventoryServerQueryOptions<TData>): MobileInventoryServerQueryResult<TData> {
  const client = useQueryClient();
  const serverState = useMobileServerStateScope();
  const inventoryScope = useQuery({
    queryKey: mobileQueryKeys.inventoryScope(serverState.scopeId),
    queryFn: ({ signal }) => serverState.loadInventoryScope({ signal }),
    staleTime: Infinity
  });
  const queryKey = inventoryScope.data
    ? key(serverState.scopeId, inventoryScope.data.tenantId, inventoryScope.data.inventoryId)
    : key(serverState.scopeId, 'inventory-pending', 'inventory-pending');

  const resourceQuery = useQuery({
    queryKey,
    queryFn: ({ signal }) => readScopedMobileResource(client, serverState.scopeId, inventoryScope.data!, signal, query),
    enabled: enabled && inventoryScope.isSuccess,
    subscribed: enabled && inventoryScope.isSuccess
  });
  const reconcile = () => fetchMobileInventoryServerQuery({ client, serverState, key, query });
  const retryScopeAndResource = async (options?: RefetchOptions) => {
    try {
      await inventoryScope.refetch({ ...options, throwOnError: true });
      return { data: await reconcile(), error: null };
    } catch (error) {
      if (options?.throwOnError) throw error;
      return { data: undefined, error: error instanceof Error ? error : new Error('Inventory could not be loaded.') };
    }
  };
  if (!inventoryScope.isError) {
    return { ...resourceQuery, data: isAccessFailure(resourceQuery.error) ? undefined : resourceQuery.data, reconcile, resourceKey: queryKey };
  }
  return {
    ...resourceQuery,
    reconcile,
    resourceKey: queryKey,
    data: undefined,
    refetch: retryScopeAndResource,
    error: inventoryScope.error,
    isError: true,
    isLoading: false,
    isPending: false,
    status: 'error'
  };
}
