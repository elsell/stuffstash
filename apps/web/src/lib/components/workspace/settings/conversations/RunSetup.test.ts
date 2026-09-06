import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
import type { WorkflowRevision } from '$lib/domain/conversationWorkflow';
import type { RunQueue, EvaluationRun } from '$lib/domain/conversationRun';
import RunSetup from './RunSetup.svelte';
let component: ReturnType<typeof mount> | undefined; let session: ConversationSession;
afterEach(async () => { if (component) await unmount(component); component = undefined; await session?.dispose(); document.body.innerHTML = ''; });
it('queues exactly the selected saved workflow and case revisions', async () => {
  const requests: RunQueue[] = [];
  const revision: WorkflowRevision = { id: 'workflow-revision', workflowId: 'workflow', number: 1, authorId: 'owner', createdAt: '', definition: {
    name: 'Household', providerProfileId: null, instructions: '', budget: { toolCalls: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 } } };
  const pagination = { limit: 20, hasMore: false, nextCursor: null };
  const unsupported = async (): Promise<never> => { throw new Error('Unsupported fake operation'); };
  const ports: ConversationWorkspaceRepositories = { apiIdentity: 'api', providers: { list: async () => [] },
    workflows: { list: async () => ({ items: [{ id: 'workflow', name: 'Household', latestRevision: 1, latestRevisionId: revision.id, activeRevisionId: null, createdAt: '', updatedAt: '' }], pagination }),
      get: async () => revision, history: async () => ({ items: [revision], pagination }), selection: async () => null, create: unsupported, append: unsupported, activate: unsupported },
    cases: { list: async () => ({ items: [{ id: 'case', title: 'Baby clothes', latestRevision: 2, latestRevisionId: 'case-revision', createdAt: '', updatedAt: '' }], pagination }), get: unsupported, create: unsupported, append: unsupported },
    runs: { list: unsupported, get: unsupported, cancel: unsupported, queue: async (_tenant, input) => { requests.push(structuredClone(input)); return new Promise<EvaluationRun>(() => {}); } } };
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunSetup, { target: document.body, props: { session, repositories: ports, onQueued: () => {} } });
  const button = (text: string) => Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes(text));
  await expect.poll(() => button('Household')).toBeDefined(); button('Household')!.click();
  await expect.poll(() => button('Baby clothes')).toBeDefined(); button('Baby clothes')!.click();
  await expect.poll(() => button('Run selected cases')?.disabled).toBe(false);
  expect(document.body.textContent).toContain('8 model calls'); button('Run selected cases')!.click();
  await expect.poll(() => requests.length).toBe(1);
  expect(requests[0]).toEqual({ workflowId: 'workflow', revisionId: 'workflow-revision', cases: [{ caseId: 'case', revisionId: 'case-revision' }] });
  expect(button('Queueing')?.disabled).toBe(true);
});
