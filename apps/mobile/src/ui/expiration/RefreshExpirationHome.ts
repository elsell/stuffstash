import type { QueryClient } from '@tanstack/react-query';
export function refreshExpirationHome(client: QueryClient, scopeId: string): Promise<void> {
 return client.refetchQueries({predicate:query => query.queryKey[1] === scopeId && query.queryKey[6] === 'expiration',type:'active'},{throwOnError:true});
}
