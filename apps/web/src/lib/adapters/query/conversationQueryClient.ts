import { MutationCache, QueryCache, QueryClient } from '@tanstack/svelte-query';
import { ConversationFailure, type ConversationRunState, type ConversationScope } from '$lib/domain/conversation';

export function conversationKey(scope: ConversationScope, ...resource: readonly string[]): readonly string[] {
  return ['conversations', scope.apiIdentity, scope.principalId, scope.tenantId, ...resource];
}

export function createConversationQueryClient(onAccessLost: () => void): QueryClient {
  let accessLost = false;
  const handleError = (error: unknown): void => {
    if (accessLost || !(error instanceof ConversationFailure) || (error.kind !== 'forbidden' && error.kind !== 'unauthenticated')) return;
    accessLost = true;
    void client.cancelQueries();
    client.clear();
    onAccessLost();
  };
  const client = new QueryClient({
    queryCache: new QueryCache({ onError: handleError }),
    mutationCache: new MutationCache({ onError: handleError }),
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        gcTime: 300_000,
        refetchOnWindowFocus: false,
        retry: (failureCount, error) => failureCount < 1 && (!(error instanceof ConversationFailure) || error.kind === 'unavailable')
      },
      mutations: { retry: false }
    }
  });
  return client;
}

export async function disposeConversationQueryClient(client: QueryClient): Promise<void> {
  try { await client.cancelQueries(); } finally { client.clear(); }
}

export function runPollInterval(state: ConversationRunState, failures: number, visible: boolean): number | false {
  if (!visible || (state !== 'queued' && state !== 'running')) return false;
  return Math.min(30_000, 2_000 * 2 ** Math.min(4, Math.max(0, failures)));
}
