import { describe, expect, it } from 'vitest';
import { RunAPIRepository } from './runRepository';
const run = {
  id: 'run', state: 'succeeded', version: 4, workflowId: 'workflow', revisionId: 'revision', totalCases: 1, completedCases: 1, passedCases: 1,
  createdAt: '2026-09-05T12:00:00Z', updatedAt: '2026-09-05T12:01:00Z', authorId: 'owner', coverage: 'text_only',
  cases: [{ caseId: 'case', revisionId: 'case-revision', title: 'Find baby clothes' }],
  providers: [{ step: 'interpret', profileId: 'local', configurationId: 'fingerprint' }],
  startedAt: '2026-09-05T12:00:01Z', finishedAt: '2026-09-05T12:01:00Z',
  results: [{ caseRevisionId: 'case-revision', modelCalls: 2, durationMilliseconds: 1532.5, completedAt: '2026-09-05T12:01:00Z',
    observation: { kind: 'answer', referencedAssets: ['clothes'], locations: [{ assetId: 'clothes', ancestorId: 'loft' }], proposals: null, executedOperations: null },
    verdict: { passed: true, failures: null } }]
};
function repository(data: unknown, requests: Request[] = []) {
  return new RunAPIRepository('https://api.example.test', () => 'session', async (input, init) => {
    requests.push(new Request(input, init)); return Response.json({ data, meta: { tenantId: 'home' } });
  });
}
describe('evaluation run API repository', () => {
  it('preserves observed facts, pinned configuration and latency without private worker fields', async () => {
    const result = await repository({ ...run, leaseToken: 'private' }).get('home', 'run');
    expect(result.coverage).toBe('text_only');
    expect(result.providers).toEqual(run.providers);
    expect(result.cases).toEqual(run.cases);
    expect(result.results[0]).toEqual({ ...run.results[0], observation: { ...run.results[0].observation, proposals: [], executedOperations: [] }, verdict: { passed: true, failures: [] } });
    expect(result).not.toHaveProperty('leaseToken');
  });
  it('sends pinned inputs and cancellation version, preserving a completed response', async () => {
    const requests: Request[] = [];
    const repo = repository(run, requests);
    const input = { workflowId: 'workflow', revisionId: 'revision', cases: [{ caseId: 'case', revisionId: 'case-revision' }] };
    await repo.queue('home', input);
    const result = await repo.cancel('home', 'run', 3);
    expect(await requests[0].json()).toEqual(input);
    expect(await requests[1].json()).toEqual({ expectedVersion: 3 });
    expect(requests[1].url).toBe('https://api.example.test/tenants/home/conversation-evaluation-runs/run/cancellation');
    expect(result.state).toBe('succeeded');
  });
  it('rejects mismatched identities and unknown states', async () => {
    await expect(repository({ ...run, id: 'other' }).get('home', 'run')).rejects.toMatchObject({ kind: 'invalid' });
    await expect(repository({ ...run, state: 'unknown' }).get('home', 'run')).rejects.toMatchObject({ kind: 'invalid' });
  });
});
