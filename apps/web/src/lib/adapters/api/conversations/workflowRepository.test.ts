import { describe, expect, it } from 'vitest';
import { WorkflowAPIRepository } from './workflowRepository';

const definition = {
  name: 'Household', retrieval: 'expanded', response: 'grounded',
  budget: { evidenceRounds: 3, modelCalls: 8, elapsedSeconds: 40, followUpTurns: 3 },
  steps: [
    { kind: 'interpret', attempts: 2, instructions: 'Recognize baby clothing', providerProfileId: 'local' },
    { kind: 'assess', attempts: 1 }, { kind: 'respond', attempts: 1 }
  ]
};
const revision = { id: 'revision', workflowId: 'workflow', number: 2, authorId: 'owner', createdAt: '2026-09-05T12:00:00Z', definition };
function repository(reply: unknown, status = 200, requests: Request[] = []) {
  return new WorkflowAPIRepository('https://api.example.test', () => 'session', async (input, init) => {
    requests.push(new Request(input, init));
    return Response.json(reply, { status });
  });
}

describe('workflow API repository', () => {
  it('reads a pinned revision and maps optional step fields without losing settings', async () => {
    const requests: Request[] = [];
    const result = await repository({ data: revision, meta: { tenantId: 'home' } }, 200, requests)
      .get('home', 'workflow', 'revision');
    expect(result.definition).toEqual({ ...definition, steps: [definition.steps[0],
      { kind: 'assess', attempts: 1, instructions: '', providerProfileId: null },
      { kind: 'respond', attempts: 1, instructions: '', providerProfileId: null }] });
    expect(requests[0].url).toBe('https://api.example.test/tenants/home/conversation-workflows/workflow/revisions/revision');
    expect(requests[0].headers.get('Authorization')).toBe('Bearer session');
  });

  it('returns the default selection and bounded list cursor', async () => {
    expect(await repository({ data: null, meta: {} }).selection('home')).toBeNull();
    const requests: Request[] = [];
    const result = await repository({ data: [], meta: { pagination: { limit: 10, hasMore: true, nextCursor: 'next' } } }, 200, requests)
      .list('home', { limit: 10, cursor: 'previous' });
    expect(result).toEqual({ items: [], pagination: { limit: 10, hasMore: true, nextCursor: 'next' } });
    expect(new URL(requests[0].url).searchParams.get('cursor')).toBe('previous');
  });

  it.each([401, 403, 409, 412, 422, 503])('maps status %s without surfacing raw error text', async status => {
    const kinds: Record<number, string> = { 401: 'unauthenticated', 403: 'forbidden', 409: 'conflict', 412: 'precondition', 422: 'invalid', 503: 'unavailable' };
    await expect(repository({ error: { message: 'private upstream details' } }, status).get('home', 'workflow'))
      .rejects.toMatchObject({ kind: kinds[status], message: kinds[status] });
  });

  it('rejects mismatched workflow, revision and tenant identities', async () => {
    for (const [data, meta] of [
      [{ ...revision, workflowId: 'other' }, {}],
      [{ ...revision, id: 'other' }, {}],
      [revision, { tenantId: 'other' }]
    ]) {
      await expect(repository({ data, meta }).get('home', 'workflow', 'revision')).rejects.toMatchObject({ kind: 'invalid' });
    }
  });

  it('rejects unsupported workflow policies', async () => {
    await expect(repository({ data: { ...revision, definition: { ...definition, retrieval: 'unknown' } }, meta: {} })
      .get('home', 'workflow')).rejects.toMatchObject({ kind: 'invalid' });
  });
});
