import { useQuery, type QueryKey, type UseQueryResult } from '@tanstack/react-query';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { useMobileServerStateScopeId } from '../navigation/MobileServerStateProvider';

type MobileServerQueryOptions<TData> = {
  readonly key: (scopeId: string) => QueryKey;
  readonly query: (signal?: AbortSignal) => Promise<TData>;
  readonly enabled?: boolean;
};

export function useMobileServerQuery<TData>({
  key,
  query,
  enabled = true
}: MobileServerQueryOptions<TData>): UseQueryResult<TData, Error> {
  const scopeId = useMobileServerStateScopeId();
  const queryKey = key(scopeId);

  return useQuery({
    queryKey,
    queryFn: ({ signal }) => query(signal),
    enabled
  });
}
