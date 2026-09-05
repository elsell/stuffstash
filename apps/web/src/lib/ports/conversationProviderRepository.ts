import type { ConversationModelChoice } from '$lib/domain/conversationProvider';
export interface ConversationProviderRepository {
  list(tenantId: string, signal?: AbortSignal): Promise<ConversationModelChoice[]>;
}
