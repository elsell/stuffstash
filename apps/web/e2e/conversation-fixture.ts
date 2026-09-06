import type { Page } from '@playwright/test';
import type { CaseDefinition, CaseRevision } from '../src/lib/domain/conversationCase';
import type { WorkflowDefinition, WorkflowRevision, WorkflowSelection } from '../src/lib/domain/conversationWorkflow';
import type { EvaluationRun, RunQueue } from '../src/lib/domain/conversationRun';

export async function installConversationFixture(page: Page) {
  const timestamp = '2026-09-05T12:00:00Z';
  let workflow: WorkflowRevision = { id: 'workflow-revision-1', workflowId: 'workflow', number: 1, authorId: 'owner', createdAt: timestamp, definition: {
    name: 'Household voice', providerProfileId: null, instructions: '', budget: { toolCalls: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 } } };
  let savedCase: CaseRevision = { id: 'case-revision-1', caseId: 'case', number: 1, authorId: 'owner', createdAt: timestamp, definition: {
    title: 'Find baby clothes', utterance: 'Where are my baby clothes?', assets: [
      { id: 'attic', title: 'Attic', kind: 'location', parentId: '', description: '', tagNames: [] },
      { id: 'clothes', title: '3–6 months clothes', kind: 'item', parentId: 'attic', description: 'Winter clothes', tagNames: ['baby', 'clothes'] }
    ], expectations: { kind: 'answer', referencedAssets: ['clothes'], locations: [{ assetId: 'clothes', ancestorId: 'attic' }], proposals: [], forbiddenOperations: ['create'] } } };
  let selection: WorkflowSelection | null = null; let run: EvaluationRun | null = null;
  const state = { denied: false, queued: [] as RunQueue[], activations: 0 };
  const workflowRevisions = new Map([[workflow.id, workflow]]); const caseRevisions = new Map([[savedCase.id, savedCase]]);
  await page.route(/http:\/\/127\.0\.0\.1:18080\/tenants\/[^/]+\/(conversation-[^?]*|provider-profiles)(\?.*)?$/, async route => {
    const request = route.request(); const path = new URL(request.url()).pathname.split('/').slice(3); const method = request.method();
    const send = async (data: unknown, status = 200) => route.fulfill({ status, contentType: 'application/json', body: JSON.stringify({ data, meta: { tenantId: 'tenant-home', pagination: { limit: 20, hasMore: false, nextCursor: null } } }) });
    if (state.denied || !request.url().includes('/tenants/tenant-home/') || request.headers().authorization !== 'Bearer e2e-token') { await route.fulfill({ status: 403, json: { error: { code: 'forbidden', message: 'Denied' } } }); return; }
    const body = method === 'POST' ? request.postDataJSON() : undefined;
    if (path[0] === 'provider-profiles') return send([]);
    if (path[0] === 'conversation-workflow-selection') return send(selection);
    if (path[0] === 'conversation-workflows') {
      if (path.length === 1) return send([{ id: workflow.workflowId, name: workflow.definition.name, latestRevisionId: workflow.id, latestRevision: workflow.number, activeRevisionId: selection?.revisionId ?? null, createdAt: timestamp, updatedAt: timestamp }]);
      if (path[2] === 'activation' && method === 'POST') { if (!run || body.runId !== run.id || body.revisionId !== run.revisionId) return send(null, 412); selection = { workflowId: workflow.workflowId, revisionId: body.revisionId }; state.activations++; return send(workflowRevisions.get(body.revisionId)); }
      if (path[2] === 'revisions' && method === 'POST') { if (body.expectedRevision !== workflow.number) return send(null, 409); workflow = { ...workflow, id: `workflow-revision-${workflow.number + 1}`, number: workflow.number + 1, definition: body.definition as WorkflowDefinition }; workflowRevisions.set(workflow.id, workflow); return send(workflow); }
      if (path[2] === 'revisions') return send(path[3] ? workflowRevisions.get(path[3]) : [...workflowRevisions.values()]);
      return send(workflow);
    }
    if (path[0] === 'conversation-evaluation-cases') {
      if (path.length === 1) return send([{ id: savedCase.caseId, title: savedCase.definition.title, latestRevisionId: savedCase.id, latestRevision: savedCase.number, createdAt: timestamp, updatedAt: timestamp }]);
      if (method === 'POST') { if (body.expectedRevision !== savedCase.number) return send(null, 409); savedCase = { ...savedCase, id: `case-revision-${savedCase.number + 1}`, number: savedCase.number + 1, definition: body.definition as CaseDefinition }; caseRevisions.set(savedCase.id, savedCase); return send(savedCase); }
      return send(path[3] ? caseRevisions.get(path[3]) : savedCase);
    }
    if (path[0] === 'conversation-evaluation-runs') {
      if (method === 'POST' && path.length === 1) {
        state.queued.push(body as RunQueue);
        run = { id: 'run', state: 'queued', version: 1, workflowId: body.workflowId, revisionId: body.revisionId, totalCases: 1, completedCases: 0, passedCases: 0,
          createdAt: timestamp, updatedAt: timestamp, authorId: 'owner', coverage: 'text_only', cases: body.cases.map((pin: { caseId: string; revisionId: string }) => ({ ...pin, title: savedCase.definition.title })),
          providers: [], results: [], startedAt: null, finishedAt: null, failureCode: '' }; return send(run);
      }
      if (path.length === 1) return send(run ? [run] : []);
      if (!run) return send(null, 404);
      run = { ...run, state: 'succeeded', version: 3, completedCases: 1, passedCases: 1, startedAt: timestamp, finishedAt: timestamp,
        results: [{ caseRevisionId: savedCase.id, modelCalls: 2, durationMilliseconds: 450, completedAt: timestamp, verdict: { passed: true, failures: [] },
          observation: { kind: 'answer', referencedAssets: ['clothes'], locations: [{ assetId: 'clothes', ancestorId: 'attic' }], proposals: [], executedOperations: [] } }] };
      return send(run);
    }
    return send(null, 404);
  });
  return state;
}
