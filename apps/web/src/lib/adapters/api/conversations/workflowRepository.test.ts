import { describe, expect, it } from 'vitest';
import { WorkflowAPIRepository } from './workflowRepository';
import type { WorkflowDefinition } from '$lib/domain/conversationWorkflow';

const definition: WorkflowDefinition = {
 name: 'Household', providerProfileId: 'local', instructions: 'Recognize baby clothing',
 budget: { toolCalls: 3, modelCalls: 8, elapsedSeconds: 40, followUpTurns: 3 }
};
const revision = { id: 'revision', workflowId: 'workflow', number: 2, authorId: 'owner', createdAt: '2026-09-05T12:00:00Z', definition };
function repository(reply: unknown, status = 200, requests: Request[] = []) {
  return new WorkflowAPIRepository('https://api.example.test', () => 'session', async (input, init) => {
    requests.push(new Request(input, init));
    return Response.json(reply, { status });
  });
}

describe('workflow API repository', () => {
  it('discloses converted settings and clears provenance on a new revision', async () => {
    const converted = await repository({ data: { ...revision, settingsMigration: 'legacy-investigation-v1' }, meta: {} }).get('home', 'workflow');
    expect(converted.settingsMigration).toBe('legacy-investigation-v1');
    const requests: Request[] = [];
    const saved = await repository({ data: { ...revision, id: 'new-revision' }, meta: {} }, 201, requests).append('home', 'workflow', 2, converted.definition);
    expect(saved.settingsMigration).toBeUndefined();
    expect(await requests[0].json()).not.toHaveProperty('settingsMigration');
  });
  it('sends immutable draft revisions and activation evidence through their scoped commands', async () => {
    const requests: Request[] = [];
    const repo = repository({ data: revision, meta: { tenantId: 'home' } }, 201, requests);
    const settings: WorkflowDefinition = { ...definition };
    await repo.create('home', settings);
    await repo.append('home', 'workflow', 1, settings);
    const evidence = { revisionId: 'revision', runId: 'run', cases: [{ caseId: 'case', revisionId: 'case-revision' }],
      expected: { workflowId: 'old-workflow', revisionId: 'old-revision' } };
    await repo.activate('home', 'workflow', evidence);
    expect(requests.map(request => new URL(request.url).pathname)).toEqual([
      '/tenants/home/conversation-workflows', '/tenants/home/conversation-workflows/workflow/revisions',
      '/tenants/home/conversation-workflows/workflow/activation'
    ]);
    expect(await requests[1].json()).toEqual({ expectedRevision: 1, definition: settings });
    expect(await requests[2].json()).toEqual(evidence);
  });

  it('propagates cancellation to the transport without retryable failure translation', async () => {
    const controller = new AbortController();
    const reason = new Error('workspace replaced');
    controller.abort(reason);
    await expect(repository({}).get('home', 'workflow', undefined, controller.signal)).rejects.toBe(reason);
  });

  it('reads a pinned revision and maps model guidance without losing settings', async () => {
    const requests: Request[] = [];
    const result = await repository({ data: revision, meta: { tenantId: 'home' } }, 200, requests)
      .get('home', 'workflow', 'revision');
    expect(result.definition).toEqual(definition);
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

  it('rejects incomplete conversation budgets', async () => {
    const { toolCalls: _retired, ...incomplete } = definition.budget;
    await expect(repository({ data: { ...revision, definition: { ...definition, budget: incomplete } }, meta: {} })
      .get('home', 'workflow')).rejects.toMatchObject({ kind: 'invalid' });
  });

  it('rejects invalid conversation budgets', async () => {
    await expect(repository({ data: { ...revision, definition: { ...definition, budget: { ...definition.budget, toolCalls: 0 } } }, meta: {} })
      .get('home', 'workflow')).rejects.toMatchObject({ kind: 'invalid' });
  });
});
