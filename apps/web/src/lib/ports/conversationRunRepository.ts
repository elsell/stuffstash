import type { EvaluationRun, RunHead, RunQueue } from '$lib/domain/conversationRun';
import type { ConversationPage, ConversationPageRequest } from './conversationWorkflowRepository';
export interface ConversationRunRepository {
  list(tenantId: string, page: ConversationPageRequest, signal?: AbortSignal): Promise<ConversationPage<RunHead>>;
  get(tenantId: string, runId: string, signal?: AbortSignal): Promise<EvaluationRun>;
  queue(tenantId: string, input: RunQueue, signal?: AbortSignal): Promise<EvaluationRun>;
  cancel(tenantId: string, runId: string, expectedVersion: number, signal?: AbortSignal): Promise<EvaluationRun>;
}
