import { caseAssetKind, caseOperation, caseOutcome } from './caseMapper';
import type { OpenAPIPaths } from '@stuff-stash/api-client';
import { ConversationFailure, type ConversationRunState } from '$lib/domain/conversation';
import type { EvaluationRun, RunHead } from '$lib/domain/conversationRun';
type RunDTO = OpenAPIPaths['/tenants/{tenantId}/conversation-evaluation-runs/{runId}']['get']['responses'][200]['content']['application/json']['data'];
type HeadDTO = NonNullable<OpenAPIPaths['/tenants/{tenantId}/conversation-evaluation-runs']['get']['responses'][200]['content']['application/json']['data']>[number];
const states: ConversationRunState[] = ['queued', 'running', 'succeeded', 'failed', 'cancelled'];
export function runHead(value: HeadDTO, runId?: string): RunHead {
  const state = states.find(state => state === value?.state);
  if (!state || !value.id || (runId && value.id !== runId)) throw new ConversationFailure('invalid');
  return { id: value.id, state, version: value.version, workflowId: value.workflowId, revisionId: value.revisionId,
    totalCases: value.totalCases, completedCases: value.completedCases, passedCases: value.passedCases,
    createdAt: value.createdAt, updatedAt: value.updatedAt };
}
export function evaluationRun(value: RunDTO, runId?: string): EvaluationRun {
  const head = runHead(value, runId);
  if (value.coverage !== 'text_only') throw new ConversationFailure('invalid');
  return { ...head, authorId: value.authorId, coverage: value.coverage,
    cases: (value.cases ?? []).map(pin => ({ caseId: pin.caseId, revisionId: pin.revisionId, title: pin.title })),
    providers: (value.providers ?? []).map(provider => ({ step: provider.step, profileId: provider.profileId, configurationId: provider.configurationId })),
    startedAt: value.startedAt, finishedAt: value.finishedAt, failureCode: value.failureCode ?? '',
    results: (value.results ?? []).map(result => ({ caseRevisionId: result.caseRevisionId, modelCalls: result.modelCalls,
      durationMilliseconds: result.durationMilliseconds, completedAt: result.completedAt,
      observation: { kind: caseOutcome(result.observation.kind), referencedAssets: [...(result.observation.referencedAssets ?? [])],
        locations: (result.observation.locations ?? []).map(location => ({ assetId: location.assetId, ancestorId: location.ancestorId })),
        proposals: (result.observation.proposals ?? []).map(proposal => ({ operation: caseOperation(proposal.operation), targetId: proposal.targetId ?? '',
          destinationId: proposal.destinationId ?? '', newKind: proposal.newKind ? caseAssetKind(proposal.newKind) : '', newTitle: proposal.newTitle ?? '', details: proposal.details ?? '' })),
        executedOperations: (result.observation.executedOperations ?? []).map(caseOperation) },
      verdict: { passed: result.verdict.passed, failures: (result.verdict.failures ?? []).map(failure => ({ code: failure.code,
        fixtureId: failure.fixtureId ?? '', operation: failure.operation ? caseOperation(failure.operation) : '' })) }
    })) };
}
