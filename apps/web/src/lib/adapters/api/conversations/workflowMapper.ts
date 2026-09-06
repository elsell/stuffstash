import type { OpenAPIPaths } from '@stuff-stash/api-client';
import { ConversationFailure } from '$lib/domain/conversation';
import type { WorkflowDefinition, WorkflowHead, WorkflowRevision, WorkflowSelection } from '$lib/domain/conversationWorkflow';

type RevisionDTO = OpenAPIPaths['/tenants/{tenantId}/conversation-workflows/{workflowId}']['get']['responses'][200]['content']['application/json']['data'];
type HeadDTO = NonNullable<OpenAPIPaths['/tenants/{tenantId}/conversation-workflows']['get']['responses'][200]['content']['application/json']['data']>[number];

export function workflowRevision(value: RevisionDTO, workflowId?: string, revisionId?: string): WorkflowRevision {
  if (!value || !value.id || !value.workflowId || (workflowId && value.workflowId !== workflowId) ||
    (revisionId && value.id !== revisionId)) throw new ConversationFailure('invalid');
  const definition = value.definition;
  if (!definition || !definition.budget || !definition.name ||
    (['modelCalls', 'toolCalls', 'elapsedSeconds', 'followUpTurns'] as const).some(key => !Number.isSafeInteger(definition.budget[key]) || definition.budget[key] < 1) ||
    (value.settingsMigration && value.settingsMigration !== 'legacy-investigation-v1')) {
    throw new ConversationFailure('invalid');
  }
  return { id: value.id, workflowId: value.workflowId, number: value.number, authorId: value.authorId, createdAt: value.createdAt,
    ...(value.settingsMigration ? { settingsMigration: 'legacy-investigation-v1' as const } : {}),
    definition: { name: definition.name, providerProfileId: definition.providerProfileId || null,
      instructions: definition.instructions ?? '', budget: { ...definition.budget } } };
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
  return { name: value.name, instructions: value.instructions, budget: { ...value.budget },
    ...(value.providerProfileId ? { providerProfileId: value.providerProfileId } : {}) };
}
