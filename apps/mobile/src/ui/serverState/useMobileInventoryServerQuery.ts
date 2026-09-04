import { useQuery, type QueryKey, type UseQueryResult } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScope } from '../navigation/MobileServerStateProvider';

type MobileInventoryServerQueryOptions<TData> = {
  readonly key: (scopeId: string, tenantId: string, inventoryId: string) => QueryKey;
  readonly query: (signal?: AbortSignal) => Promise<TData>;
  readonly enabled?: boolean;
};

export function useMobileInventoryServerQuery<TData>({
  key,
  query,
  enabled = true
}: MobileInventoryServerQueryOptions<TData>): UseQueryResult<TData, Error> {
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
    queryFn: ({ signal }) => query(signal),
    enabled: enabled && inventoryScope.isSuccess
  });
  if (!inventoryScope.isError) {
    return resourceQuery;
  }
  return {
    ...resourceQuery,
    error: inventoryScope.error,
    isError: true,
    isLoading: false,
    isPending: false,
    status: 'error'
  } as UseQueryResult<TData, Error>;
}
