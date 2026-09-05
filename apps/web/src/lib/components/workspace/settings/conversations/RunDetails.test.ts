import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { createConversationSession } from '$lib/adapters/query/conversationSession';
import type { ConversationSession } from '$lib/adapters/query/conversationSession';
import type { EvaluationRun } from '$lib/domain/conversationRun';
import type { ConversationRunRepository } from '$lib/ports/conversationRunRepository';
import RunDetails from './RunDetails.svelte';
let component: ReturnType<typeof mount> | undefined; let session: ConversationSession;
afterEach(async () => { if (component) await unmount(component); component = undefined; await session?.dispose(); document.body.innerHTML = ''; });
const queued: EvaluationRun = { id: 'run', state: 'queued', version: 1, workflowId: 'workflow', revisionId: 'revision',
  totalCases: 1, completedCases: 0, passedCases: 0, createdAt: '', updatedAt: '', authorId: 'owner', coverage: 'text_only',
  cases: [{ caseId: 'case', revisionId: 'case-revision', title: 'Baby clothes' }], providers: [], results: [], startedAt: null, finishedAt: null, failureCode: '' };
class Runs implements ConversationRunRepository {
  value = structuredClone(queued); cancelledVersion = 0;
  async list() { return { items: [this.value], pagination: { limit: 20, hasMore: false, nextCursor: null } }; }
  async get() { return structuredClone(this.value); }
  async queue() { return this.value; }
  async cancel(_tenant: string, _id: string, expected: number) { this.cancelledVersion = expected; this.value = { ...this.value, state: 'failed', version: 2, failureCode: 'provider_unavailable' }; return this.value; }
}
const cases = { list: async () => ({ items: [], pagination: { limit: 20, hasMore: false, nextCursor: null } }),
  get: async (): Promise<never> => { throw new Error('A collapsed result must not load fixtures'); },
  create: async (): Promise<never> => { throw new Error('Read only'); }, append: async (): Promise<never> => { throw new Error('Read only'); } };
function button(text: string) { return Array.from(document.querySelectorAll('button')).find(value => value.textContent?.trim() === text); }
it('cancels against the observed version and displays the actual terminal result', async () => {
  const runs = new Runs(); session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunDetails, { target: document.body, props: { session, runs, cases, runId: 'run', visible: true } });
  await expect.poll(() => button('Cancel run')).toBeDefined(); button('Cancel run')!.click();
  await expect.poll(() => runs.cancelledVersion).toBe(1);
  await expect.poll(() => document.body.textContent).toContain('Failed');
  expect(button('Cancel run')).toBeUndefined(); expect(document.body.textContent).toContain('Text-only');
  expect(document.body.textContent).not.toContain('Cancelled');
});
it('shows unrun cases as pending rather than passing', async () => {
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunDetails, { target: document.body, props: { session, runs: new Runs(), cases, runId: 'run', visible: false } });
  await expect.poll(() => document.body.textContent).toContain('Not run yet');
  expect(document.body.textContent).toContain('0 of 1');
});
