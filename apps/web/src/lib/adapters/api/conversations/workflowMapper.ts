import type { OpenAPIPaths } from '@stuff-stash/api-client';
import { ConversationFailure } from '$lib/domain/conversation';
import type { WorkflowDefinition, WorkflowHead, WorkflowRevision, WorkflowSelection, WorkflowStepKind } from '$lib/domain/conversationWorkflow';

type RevisionDTO = OpenAPIPaths['/tenants/{tenantId}/conversation-workflows/{workflowId}']['get']['responses'][200]['content']['application/json']['data'];
type HeadDTO = NonNullable<OpenAPIPaths['/tenants/{tenantId}/conversation-workflows']['get']['responses'][200]['content']['application/json']['data']>[number];

export function workflowRevision(value: RevisionDTO, workflowId?: string, revisionId?: string): WorkflowRevision {
  if (!value || !value.id || !value.workflowId || (workflowId && value.workflowId !== workflowId) ||
    (revisionId && value.id !== revisionId)) throw new ConversationFailure('invalid');
  const definition = value.definition;
  if (!definition || !['precise_first', 'expanded'].includes(definition.retrieval) ||
    !['generated_with_grounded_fallback', 'grounded'].includes(definition.response) || definition.steps?.length !== 3) {
    throw new ConversationFailure('invalid');
  }
  const order: WorkflowStepKind[] = ['interpret', 'assess', 'respond'];
  const steps = definition.steps.map((step, index) => {
    if (step.kind !== order[index]) throw new ConversationFailure('invalid');
    return { kind: order[index], attempts: step.attempts, instructions: step.instructions ?? '', providerProfileId: step.providerProfileId || null };
  });
  return { id: value.id, workflowId: value.workflowId, number: value.number, authorId: value.authorId, createdAt: value.createdAt,
    definition: { name: definition.name, retrieval: definition.retrieval as WorkflowDefinition['retrieval'],
      response: definition.response as WorkflowDefinition['response'], budget: { ...definition.budget }, steps } };
}
export function workflowHead(value: HeadDTO): WorkflowHead {
  return { id: value.id, name: value.name, latestRevisionId: value.latestRevisionId, latestRevision: value.latestRevision,
    activeRevisionId: value.activeRevisionId || null, createdAt: value.createdAt, updatedAt: value.updatedAt };
}
export function workflowSelection(value: WorkflowSelection | null): WorkflowSelection | null {
  if (value === null) return null;
  if (!value?.workflowId || !value.revisionId) throw new ConversationFailure('invalid');
  return { workflowId: value.workflowId, revisionId: value.revisionId };
}
export function workflowDefinitionBody(value: WorkflowDefinition) {
  return { name: value.name, retrieval: value.retrieval, response: value.response, budget: { ...value.budget },
    steps: value.steps.map(step => ({ kind: step.kind, attempts: step.attempts, instructions: step.instructions,
      ...(step.providerProfileId ? { providerProfileId: step.providerProfileId } : {}) })) };
}
