import { createAuthenticatedTransport, type TokenProvider } from '@stuff-stash/api-client';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import type { ConversationCaseRepository } from '$lib/ports/conversationCaseRepository';
import type { ConversationPageRequest } from '$lib/ports/conversationWorkflowRepository';
import { conversationPage, conversationResponse } from './response';
import { caseDefinitionBody, caseRevision } from './caseMapper';

export class CaseAPIRepository implements ConversationCaseRepository {
  private readonly client;
  constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
    this.client = createAuthenticatedTransport({ baseUrl, tokenProvider, fetch: fetchImpl });
  }
  async list(tenantId: string, page: ConversationPageRequest, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/conversation-evaluation-cases', {
      params: { path: { tenantId }, query: page }, signal
    }), tenantId, signal);
    return conversationPage(result, value => ({ id: value.id, title: value.title, latestRevision: value.latestRevision,
      latestRevisionId: value.latestRevisionId, createdAt: value.createdAt, updatedAt: value.updatedAt }));
  }
  async get(tenantId: string, caseId: string, revisionId?: string, signal?: AbortSignal) {
    const pending = revisionId
      ? this.client.GET('/tenants/{tenantId}/conversation-evaluation-cases/{caseId}/revisions/{revisionId}', {
        params: { path: { tenantId, caseId, revisionId } }, signal })
      : this.client.GET('/tenants/{tenantId}/conversation-evaluation-cases/{caseId}', {
        params: { path: { tenantId, caseId } }, signal });
    return caseRevision((await conversationResponse(pending, tenantId, signal)).data, caseId, revisionId);
  }
  async create(tenantId: string, definition: CaseDefinition, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-evaluation-cases', {
      params: { path: { tenantId } }, body: { definition: caseDefinitionBody(definition) }, signal
    }), tenantId, signal);
    return caseRevision(result.data);
  }
  async append(tenantId: string, caseId: string, expectedRevision: number, definition: CaseDefinition, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-evaluation-cases/{caseId}/revisions', {
      params: { path: { tenantId, caseId } }, body: { expectedRevision, definition: caseDefinitionBody(definition) }, signal
    }), tenantId, signal);
    return caseRevision(result.data, caseId);
  }
}
