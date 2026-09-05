import { createAuthenticatedTransport, type TokenProvider } from '@stuff-stash/api-client';
import type { RunQueue } from '$lib/domain/conversationRun';
import { ConversationFailure } from '$lib/domain/conversation';
import type { ConversationRunRepository } from '$lib/ports/conversationRunRepository';
import type { ConversationPageRequest } from '$lib/ports/conversationWorkflowRepository';
import { conversationPage, conversationResponse } from './response';
import { evaluationRun, runHead } from './runMapper';
export class RunAPIRepository implements ConversationRunRepository {
  private readonly client;
  constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) {
    this.client = createAuthenticatedTransport({ baseUrl, tokenProvider, fetch: fetchImpl });
  }
  async list(tenantId: string, page: ConversationPageRequest, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/conversation-evaluation-runs', {
      params: { path: { tenantId }, query: page }, signal
    }), tenantId, signal);
    return conversationPage(result, value => runHead(value));
  }
  async get(tenantId: string, runId: string, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.GET('/tenants/{tenantId}/conversation-evaluation-runs/{runId}', {
      params: { path: { tenantId, runId } }, signal
    }), tenantId, signal);
    return evaluationRun(result.data, runId);
  }
  async queue(tenantId: string, input: RunQueue, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-evaluation-runs', {
      params: { path: { tenantId } }, body: { workflowId: input.workflowId, revisionId: input.revisionId, cases: input.cases.map(pin => ({ ...pin })) }, signal
    }), tenantId, signal);
    const run = evaluationRun(result.data);
    if (run.workflowId !== input.workflowId || run.revisionId !== input.revisionId || run.cases.length !== input.cases.length ||
      run.cases.some((pin, index) => pin.caseId !== input.cases[index].caseId || pin.revisionId !== input.cases[index].revisionId)) {
      throw new ConversationFailure('invalid');
    }
    return run;
  }
  async cancel(tenantId: string, runId: string, expectedVersion: number, signal?: AbortSignal) {
    const result = await conversationResponse(this.client.POST('/tenants/{tenantId}/conversation-evaluation-runs/{runId}/cancellation', {
      params: { path: { tenantId, runId } }, body: { expectedVersion }, signal
    }), tenantId, signal);
    return evaluationRun(result.data, runId);
  }
}
