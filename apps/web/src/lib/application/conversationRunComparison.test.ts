import { expect, it } from 'vitest';
import type { EvaluationRun } from '$lib/domain/conversationRun';
import { compareConversationRuns } from './conversationRunComparison';
function run(id: string): EvaluationRun {
  return { id, workflowId: 'workflow', revisionId: id, state: 'succeeded', version: 3, totalCases: 1, completedCases: 1, passedCases: 1, authorId: 'owner',
    createdAt: '', updatedAt: '', startedAt: '', finishedAt: '', failureCode: '', coverage: 'text_only', cases: [{ caseId: 'case', revisionId: 'case-revision', title: 'Baby clothes' }],
    providers: [{ step: 'interpret', profileId: 'profile', configurationId: 'configuration' }], results: [{ caseRevisionId: 'case-revision', modelCalls: 2, durationMilliseconds: 300,
      completedAt: '', verdict: { passed: true, failures: [] }, observation: { kind: 'answer', referencedAssets: [], locations: [], proposals: [], executedOperations: [] } }] };
}
it('compares recorded quality, calls and execution time for identical pinned cases', () => {
  const baseline = run('baseline'); const candidate = run('candidate');
  candidate.results[0].durationMilliseconds = 100; candidate.results[0].modelCalls = 1;
  const comparison = compareConversationRuns(baseline, candidate);
  expect(comparison.compatible).toBe(true);
  if (!comparison.compatible) throw new Error('Expected compatible runs');
  expect(comparison.cases[0]).toEqual({ title: 'Baby clothes', baseline: { passed: true, modelCalls: 2, durationMilliseconds: 300 }, candidate: { passed: true, modelCalls: 1, durationMilliseconds: 100 } });
  expect(comparison.baseline).toEqual({ passedCases: 1, modelCalls: 2, durationMilliseconds: 300 });
  expect(comparison.candidate).toEqual({ passedCases: 1, modelCalls: 1, durationMilliseconds: 100 });
});
it.each(['cases', 'providers', 'incomplete', 'same'] as const)('rejects %s comparisons with a specific explanation', change => {
  const baseline = run('baseline'); const candidate = run('candidate');
  if (change === 'cases') candidate.cases[0].revisionId = 'new-case-revision';
  if (change === 'providers') candidate.providers[0].configurationId = 'new-configuration';
  if (change === 'incomplete') candidate.state = 'running';
  if (change === 'same') candidate.id = baseline.id;
  expect(compareConversationRuns(baseline, candidate)).toEqual({ compatible: false, reason: change });
});
