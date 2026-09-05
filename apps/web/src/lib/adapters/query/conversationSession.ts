import type { ConversationScope } from '$lib/domain/conversation';
import { createConversationQueryClient, disposeConversationQueryClient } from './conversationQueryClient';

export function createConversationSession(scope: ConversationScope, onAccessLost: () => void) {
  let active = true;
  const ownedScope = Object.freeze({ ...scope });
  const client = createConversationQueryClient(() => {
    if (!active) return;
    active = false;
    onAccessLost();
  });
  return {
    scope: ownedScope,
    client,
    get active() { return active; },
    async mutate<T>(operation: () => Promise<T>, reconcile: (result: T) => void): Promise<T> {
      if (!active) throw new DOMException('Conversation session is no longer active.', 'AbortError');
      const mutation = client.getMutationCache().build<T, Error, void, unknown>(client, {
        mutationFn: async () => {
          if (!active) throw new DOMException('Conversation session is no longer active.', 'AbortError');
          return operation();
        },
        retry: false
      });
      const result = await mutation.execute(undefined);
      if (active) reconcile(result);
      return result;
    },
    async dispose(): Promise<void> {
      active = false;
      await disposeConversationQueryClient(client);
    }
  };
}
export type ConversationSession = ReturnType<typeof createConversationSession>;
