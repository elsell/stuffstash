import type { CurrentInventoryScope } from '../../application/home/CurrentInventoryScopeQuery';
import { CancelledError, type QueryClient, type QueryKey } from '@tanstack/react-query';
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
    queryFn: ({ signal }) => readScopedMobileResource(input.client, input.serverState.scopeId, inventoryScope, signal, input.query),
    ...(input.force ? { staleTime: 0 } : {})
  });
}

/** Prevent implicit-selected-scope ports from publishing under a departed query key. */
export async function readScopedMobileResource<T>(client: QueryClient, scopeId: string, expected: CurrentInventoryScope, signal: AbortSignal, read: (signal?: AbortSignal) => Promise<T>): Promise<T> {
  const assertCurrent = () => {
    signal.throwIfAborted();
    const current = client.getQueryData<CurrentInventoryScope>(mobileQueryKeys.inventoryScope(scopeId));
    if (!current || current.tenantId !== expected.tenantId || current.inventoryId !== expected.inventoryId) throw new CancelledError();
  };
  assertCurrent();
  const value = await read(signal);
  assertCurrent();
  return value;
}
