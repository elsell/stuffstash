import type { QueryClient, QueryKey } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import type { MobileServerStateScope } from '../navigation/MobileServerStateProvider';

export async function fetchMobileInventoryServerQuery<TData>(input: {
  readonly client: QueryClient;
  readonly serverState: MobileServerStateScope;
  readonly key: (
    scopeId: string,
    tenantId: string,
    inventoryId: string
  ) => QueryKey;
  readonly query: (signal?: AbortSignal) => Promise<TData>;
  readonly force?: boolean;
}): Promise<TData> {
  const inventoryScope = await input.client.ensureQueryData({
    queryKey: mobileQueryKeys.inventoryScope(input.serverState.scopeId),
    queryFn: ({ signal }) => input.serverState.loadInventoryScope({ signal }),
    staleTime: Infinity
  });
  const queryKey = input.key(
    input.serverState.scopeId,
    inventoryScope.tenantId,
    inventoryScope.inventoryId
  );
  if (input.force) {
    await input.client.invalidateQueries({ queryKey, exact: true, refetchType: 'none' });
  }
  return input.client.fetchQuery({
    queryKey,
    queryFn: ({ signal }) => input.query(signal),
    ...(input.force ? { staleTime: 0 } : {})
  });
}
