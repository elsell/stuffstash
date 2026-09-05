import type { OpenAPIPaths } from '@stuff-stash/api-client';
import { ConversationFailure } from '$lib/domain/conversation';
import type { CaseAssetKind, CaseDefinition, CaseOperation, CaseOutcome, CaseRevision } from '$lib/domain/conversationCase';

type RevisionDTO = OpenAPIPaths['/tenants/{tenantId}/conversation-evaluation-cases/{caseId}']['get']['responses'][200]['content']['application/json']['data'];
const kinds: CaseAssetKind[] = ['item', 'container', 'location'];
const outcomes: CaseOutcome[] = ['answer', 'clarification', 'proposal', 'failure'];
const operations: CaseOperation[] = ['locate', 'exists', 'list_inventory', 'list_contents', 'detail', 'checkout_status',
  'asset_history', 'checkout_history', 'create', 'move', 'archive', 'restore', 'checkout', 'return', 'unsupported'];
function member<T extends string>(value: string, values: T[]): T {
  const found = values.find(candidate => candidate === value);
  if (!found) throw new ConversationFailure('invalid');
  return found;
}
export function caseRevision(value: RevisionDTO, caseId?: string, revisionId?: string): CaseRevision {
  if (!value?.id || !value.caseId || (caseId && value.caseId !== caseId) || (revisionId && value.id !== revisionId) ||
    !value.definition?.expectations) throw new ConversationFailure('invalid');
  const definition = value.definition;
  const expected = definition.expectations;
  return { id: value.id, caseId: value.caseId, number: value.number, authorId: value.authorId, createdAt: value.createdAt,
    definition: { title: definition.title, utterance: definition.utterance,
      assets: (definition.assets ?? []).map(asset => ({ id: asset.id, title: asset.title, kind: member(asset.kind, kinds),
        description: asset.description ?? '', parentId: asset.parentId ?? '', tagNames: [...(asset.tagNames ?? [])] })),
      expectations: { kind: member(expected.kind, outcomes), referencedAssets: [...(expected.referencedAssets ?? [])],
        locations: (expected.locations ?? []).map(location => ({ assetId: location.assetId, ancestorId: location.ancestorId })),
        forbiddenOperations: (expected.forbiddenOperations ?? []).map(operation => member(operation, operations)),
        proposals: (expected.proposals ?? []).map(proposal => ({ operation: member(proposal.operation, operations),
          targetId: proposal.targetId ?? '', destinationId: proposal.destinationId ?? '',
          newKind: proposal.newKind ? member(proposal.newKind, kinds) : '', newTitle: proposal.newTitle ?? '', details: proposal.details ?? '' }))
      } } };
}
export function caseDefinitionBody(value: CaseDefinition) {
  return { title: value.title, utterance: value.utterance,
    assets: value.assets.map(asset => ({ ...asset, tagNames: [...asset.tagNames] })),
    expectations: { kind: value.expectations.kind, referencedAssets: [...value.expectations.referencedAssets],
      forbiddenOperations: [...value.expectations.forbiddenOperations], locations: value.expectations.locations.map(value => ({ ...value })),
      proposals: value.expectations.proposals.map(value => ({ ...value })) } };
}
