import type { WorkflowActivation, WorkflowDefinition, WorkflowHead, WorkflowRevision, WorkflowSelection } from '$lib/domain/conversationWorkflow';
import type { Pagination } from './pagination';

export interface ConversationPage<T> { items: T[]; pagination: Pagination }
export interface ConversationPageRequest { limit: number; cursor?: string }
export interface ConversationWorkflowRepository {
  list(tenantId: string, page: ConversationPageRequest, signal?: AbortSignal): Promise<ConversationPage<WorkflowHead>>;
  history(tenantId: string, workflowId: string, page: ConversationPageRequest, signal?: AbortSignal): Promise<ConversationPage<WorkflowRevision>>;
  selection(tenantId: string, signal?: AbortSignal): Promise<WorkflowSelection | null>;
  get(tenantId: string, workflowId: string, revisionId?: string, signal?: AbortSignal): Promise<WorkflowRevision>;
  create(tenantId: string, definition: WorkflowDefinition, signal?: AbortSignal): Promise<WorkflowRevision>;
  append(tenantId: string, workflowId: string, expectedRevision: number, definition: WorkflowDefinition, signal?: AbortSignal): Promise<WorkflowRevision>;
  activate(tenantId: string, workflowId: string, input: WorkflowActivation, signal?: AbortSignal): Promise<WorkflowRevision>;
}
