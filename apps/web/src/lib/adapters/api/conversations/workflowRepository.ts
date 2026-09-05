import { createAuthenticatedTransport, type TokenProvider } from '@stuff-stash/api-client';
import type { WorkflowActivation, WorkflowDefinition } from '$lib/domain/conversationWorkflow';
import type { ConversationPageRequest, ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
import { conversationPage, conversationResponse } from './response';
import { workflowDefinitionBody, workflowHead, workflowRevision, workflowSelection } from './workflowMapper';

export class WorkflowAPIRepository implements ConversationWorkflowRepository {
  private readonly client;
  constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
    this.client = createAuthenticatedTransport({ baseUrl, tokenProvider, fetch: fetchImpl });
  }
  async list(tenantId: string, page: ConversationPageRequest, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/conversation-workflows', {
      params: { path: { tenantId }, query: page }, signal
    }), tenantId, signal);
    return conversationPage(result, workflowHead);
  }
  async history(tenantId: string, workflowId: string, page: ConversationPageRequest, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions', {
      params: { path: { tenantId, workflowId }, query: page }, signal
    }), tenantId, signal);
    return conversationPage(result, value => workflowRevision(value, workflowId));
  }
  async selection(tenantId: string, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/conversation-workflow-selection', {
      params: { path: { tenantId } }, signal
    }), tenantId, signal);
    return workflowSelection(result.data);
  }
  async get(tenantId: string, workflowId: string, revisionId?: string, signal?: AbortSignal) {
    const pending = revisionId
      ? this.client.GET('/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions/{revisionId}', {
        params: { path: { tenantId, workflowId, revisionId } }, signal })
      : this.client.GET('/tenants/{tenantId}/conversation-workflows/{workflowId}', {
        params: { path: { tenantId, workflowId } }, signal });
    return workflowRevision((await conversationResponse(pending, tenantId, signal)).data, workflowId, revisionId);
  }
  async create(tenantId: string, definition: WorkflowDefinition, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-workflows', {
      params: { path: { tenantId } }, body: { definition: workflowDefinitionBody(definition) }, signal
    }), tenantId, signal);
    return workflowRevision(result.data);
  }
  async append(tenantId: string, workflowId: string, expectedRevision: number, definition: WorkflowDefinition, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-workflows/{workflowId}/revisions', {
      params: { path: { tenantId, workflowId } }, body: { expectedRevision, definition: workflowDefinitionBody(definition) }, signal
    }), tenantId, signal);
    return workflowRevision(result.data, workflowId);
  }
  async activate(tenantId: string, workflowId: string, input: WorkflowActivation, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-workflows/{workflowId}/activation', {
      params: { path: { tenantId, workflowId } }, body: { revisionId: input.revisionId, runId: input.runId,
        cases: input.cases.map(pin => ({ ...pin })), ...(input.expected ? { expected: { ...input.expected } } : {}) }, signal
    }), tenantId, signal);
    return workflowRevision(result.data, workflowId, input.revisionId);
  }
}
