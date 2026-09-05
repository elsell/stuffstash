import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
import { conversationKey } from '$lib/adapters/query/conversationQueryClient';
import type { EvaluationRun } from '$lib/domain/conversationRun';
import type { WorkflowActivation, WorkflowRevision } from '$lib/domain/conversationWorkflow';
import { ConversationFailure } from '$lib/domain/conversation';
import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
import RunActivation from './RunActivation.svelte';
let component: ReturnType<typeof mount> | undefined; let session: ConversationSession;
afterEach(async () => { if (component) await unmount(component); component = undefined; await session?.dispose(); document.body.innerHTML = ''; });
const revision: WorkflowRevision = { id: 'revision', workflowId: 'workflow', number: 2, authorId: 'owner', createdAt: '', definition: { name: 'Household', retrieval: 'precise_first', response: 'grounded', steps: [], budget: { evidenceRounds: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 } } };
const run: EvaluationRun = { id: 'run', workflowId: 'workflow', revisionId: 'revision', state: 'succeeded', version: 3, totalCases: 1, completedCases: 1, passedCases: 1,
  createdAt: '', updatedAt: '', authorId: 'owner', coverage: 'text_only', cases: [{ caseId: 'case', revisionId: 'case-revision', title: 'Baby clothes' }], providers: [],
  results: [{ caseRevisionId: 'case-revision', modelCalls: 1, durationMilliseconds: 100, completedAt: '', observation: { kind: 'answer', referencedAssets: [], locations: [], proposals: [], executedOperations: [] }, verdict: { passed: true, failures: [] } }], startedAt: '', finishedAt: '', failureCode: '' };
const unsupported = async (): Promise<never> => { throw new Error('Unsupported fake operation'); };
function repository(): ConversationWorkflowRepository { return { list: unsupported, history: unsupported, get: async () => revision, selection: async () => null, create: unsupported, append: unsupported, activate: async () => revision }; }
function button() { return Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('Activate tested revision')); }
it('sends exact passing evidence and expected selection before showing activation', async () => {
  const workflows = repository(); const requests: WorkflowActivation[] = []; let finish!: (value: WorkflowRevision) => void;
  workflows.activate = async (_tenant, _workflow, input) => { requests.push(input); return new Promise(resolve => { finish = resolve; }); };
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunActivation, { target: document.body, props: { session, workflows, run } });
  await expect.poll(() => button()?.disabled).toBe(false); button()!.click();
  await expect.poll(() => requests.length).toBe(1);
  expect(requests[0]).toEqual({ revisionId: 'revision', runId: 'run', cases: [{ caseId: 'case', revisionId: 'case-revision' }], expected: null });
  let releaseOld!: () => void;
  const oldRead = session.client.fetchQuery({ queryKey: conversationKey(session.scope, 'selection'), staleTime: 0,
    queryFn: () => new Promise<null>(resolve => { releaseOld = () => resolve(null); }) }).catch(() => undefined);
  await expect.poll(() => releaseOld).toBeDefined();
  expect(document.body.textContent).not.toContain('This revision is active'); finish(revision);
  await expect.poll(() => document.body.textContent).toContain('This revision is active');
  releaseOld(); await oldRead;
  expect(session.client.getQueryData(conversationKey(session.scope, 'selection'))).toEqual({ workflowId: 'workflow', revisionId: 'revision' });
});
it('does not activate when the server rejects stale quality evidence', async () => {
  const workflows = repository(); workflows.activate = async () => { throw new ConversationFailure('precondition'); };
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunActivation, { target: document.body, props: { session, workflows, run } });
  await expect.poll(() => button()?.disabled).toBe(false); button()!.click();
  await expect.poll(() => document.body.textContent).toContain('Run the cases again');
  expect(document.body.textContent).not.toContain('This revision is active');
});
