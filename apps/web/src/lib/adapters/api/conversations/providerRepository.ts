import { createAuthenticatedTransport, type TokenProvider } from '@stuff-stash/api-client';
import { ConversationFailure } from '$lib/domain/conversation';
import type { ConversationModelChoice } from '$lib/domain/conversationProvider';
import type { ConversationProviderRepository } from '$lib/ports/conversationProviderRepository';
import { conversationResponse } from './response';
export class ConversationProviderAPIRepository implements ConversationProviderRepository {
  private readonly client;
  constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
    this.client = createAuthenticatedTransport({ baseUrl, tokenProvider, fetch: fetchImpl });
  }
  async list(tenantId: string, signal?: AbortSignal): Promise<ConversationModelChoice[]> {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/provider-profiles', {
      params: { path: { tenantId } }, signal
    }), tenantId, signal);
    const profiles = result.data ?? [];
    if (profiles.some(profile => profile.tenantId !== tenantId)) throw new ConversationFailure('invalid');
    return profiles.filter(profile => profile.capability === 'language_inference' && profile.lifecycleState === 'enabled')
      .map(profile => ({ id: profile.id, name: profile.displayName, providerKind: profile.providerKind, modelName: profile.modelName }));
  }
}
