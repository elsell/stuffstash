import type { CaseDefinition, CaseHead, CaseRevision } from '$lib/domain/conversationCase';
import type { ConversationPage, ConversationPageRequest } from './conversationWorkflowRepository';
export interface ConversationCaseRepository {
  list(tenantId: string, page: ConversationPageRequest, signal?: AbortSignal): Promise<ConversationPage<CaseHead>>;
  get(tenantId: string, caseId: string, revisionId?: string, signal?: AbortSignal): Promise<CaseRevision>;
  create(tenantId: string, definition: CaseDefinition, signal?: AbortSignal): Promise<CaseRevision>;
  append(tenantId: string, caseId: string, expectedRevision: number, definition: CaseDefinition, signal?: AbortSignal): Promise<CaseRevision>;
}
