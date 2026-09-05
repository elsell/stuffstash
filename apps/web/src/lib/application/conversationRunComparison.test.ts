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
it('compares a fully evaluated failing baseline with an improved passing revision', () => {
  const baseline = run('baseline'); baseline.state = 'failed'; baseline.passedCases = 0;
  baseline.results[0].verdict = { passed: false, failures: [{ code: 'missing_reference', fixtureId: 'clothes', operation: '' }] };
  const comparison = compareConversationRuns(baseline, run('candidate'));
  expect(comparison.compatible).toBe(true);
  if (!comparison.compatible) throw new Error('Expected assertion failures to remain comparable');
  expect(comparison.baseline.passedCases).toBe(0); expect(comparison.candidate.passedCases).toBe(1);
  expect(comparison.cases[0].baseline.passed).toBe(false);
});
it('does not treat an operational failure as a completed comparison', () => {
  const baseline = run('baseline'); baseline.state = 'failed'; baseline.failureCode = 'provider_unavailable';
  expect(compareConversationRuns(baseline, run('candidate'))).toEqual({ compatible: false, reason: 'incomplete' });
});
