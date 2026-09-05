import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
import type { EvaluationRun } from '$lib/domain/conversationRun';
import { ConversationFailure } from '$lib/domain/conversation';
import RunWorkspace from './RunWorkspace.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const run: EvaluationRun = { id: 'run', state: 'queued', version: 1, workflowId: 'workflow', revisionId: 'revision', totalCases: 1, completedCases: 0, passedCases: 0,
  createdAt: '2026-09-05T12:00:00Z', updatedAt: '', authorId: 'owner', coverage: 'text_only', cases: [{ caseId: 'case', revisionId: 'case-revision', title: 'Clothes' }],
  providers: [], results: [], startedAt: null, finishedAt: null, failureCode: '' };
const page = { items: [], pagination: { limit: 20, hasMore: false, nextCursor: null } };
const unsupported = async (): Promise<never> => { throw new Error('Read only'); };
function repositories(): ConversationWorkspaceRepositories {
  return { apiIdentity: 'api', providers: { list: async () => [] },
    workflows: { list: async () => page, history: async () => page, selection: async () => null, get: unsupported, create: unsupported, append: unsupported, activate: unsupported },
    cases: { list: async () => page, get: unsupported, create: unsupported, append: unsupported },
    runs: { list: async () => ({ ...page, items: [run] }), get: async () => run, queue: unsupported, cancel: unsupported } };
}
const scope = { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' };
it('opens a saved run through the tenant repository', async () => {
  const ports = repositories(); const reads: string[] = [];
  ports.runs.get = async tenant => { reads.push(tenant); return run; };
  component = mount(RunWorkspace, { target: document.body, props: { scope, repositories: ports, visible: true } });
  const button = () => Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('Queued'));
  await expect.poll(button).toBeDefined(); button()!.click();
  await expect.poll(() => reads).toEqual(['home']);
  await expect.poll(() => document.body.textContent).toContain('Text-only evaluation');
});
it('notifies the owner and removes run content on configure denial', async () => {
  const ports = repositories(); let denied = false; ports.runs.list = async () => { throw new ConversationFailure('forbidden'); };
  component = mount(RunWorkspace, { target: document.body, props: { scope, repositories: ports, visible: true, onAccessLost: () => { denied = true; } } });
  await expect.poll(() => denied).toBe(true);
  expect(document.body.textContent).not.toContain('New run');
});
