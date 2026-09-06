import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
import type { EvaluationRun } from '$lib/domain/conversationRun';
import RunComparison from './RunComparison.svelte';
let component: ReturnType<typeof mount> | undefined; let session: ConversationSession;
afterEach(async () => { if (component) await unmount(component); component = undefined; await session?.dispose(); document.body.innerHTML = ''; });
it('loads a baseline on demand and refuses changed provider configuration', async () => {
  const current: EvaluationRun = { id: 'current', workflowId: 'workflow', revisionId: 'revision', state: 'succeeded', version: 3, totalCases: 1, completedCases: 1, passedCases: 1,
    createdAt: '2026-09-05T12:00:00Z', updatedAt: '', authorId: 'owner', coverage: 'text_only', startedAt: '', finishedAt: '', failureCode: '',
    providers: [{ profileId: 'model', configurationId: 'current-config' }], cases: [{ caseId: 'case', revisionId: 'case-revision', title: 'Clothes' }],
    results: [{ caseRevisionId: 'case-revision', verdict: { passed: true, failures: [] }, modelCalls: 1, durationMilliseconds: 100, completedAt: '', observation: { kind: 'answer', referencedAssets: [], locations: [], proposals: [], executedOperations: [] } }] };
  const baseline = { ...current, id: 'baseline', providers: [{ ...current.providers[0], configurationId: 'previous-config' }] };
  let reads = 0; const unsupported = async (): Promise<never> => { throw new Error('Read only'); };
  const runs = { list: async () => ({ items: [baseline], pagination: { limit: 20, hasMore: false, nextCursor: null } }), get: async () => { reads++; return baseline; }, queue: unsupported, cancel: unsupported };
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunComparison, { target: document.body, props: { session, runs, current } });
  expect(reads).toBe(0); document.querySelector('button')!.click();
  const select = () => Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('1/1 passed'));
  await expect.poll(select).toBeDefined(); select()!.click();
  await expect.poll(() => document.body.textContent).toContain('Provider configuration differs'); expect(reads).toBe(1);
  expect(document.querySelector('table')).toBeNull();
});
