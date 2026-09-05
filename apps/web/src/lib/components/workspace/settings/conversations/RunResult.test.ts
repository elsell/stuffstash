import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
import type { CaseRevision } from '$lib/domain/conversationCase';
import type { RunResult as Result } from '$lib/domain/conversationRun';
import RunResult from './RunResult.svelte';
let component: ReturnType<typeof mount> | undefined; let session: ConversationSession;
afterEach(async () => { if (component) await unmount(component); component = undefined; await session?.dispose(); document.body.innerHTML = ''; });
it('lazily reads the pinned case and exposes a mismatched creation kind', async () => {
  const reads: string[] = [];
  const revision: CaseRevision = { id: 'pinned', caseId: 'case', number: 1, authorId: 'owner', createdAt: '', definition: {
    title: 'Another box', utterance: 'Add another box', assets: [], expectations: { kind: 'proposal', referencedAssets: [], locations: [], forbiddenOperations: [],
      proposals: [{ operation: 'create', newTitle: 'Box', newKind: 'container', targetId: '', destinationId: '', details: '' }] } } };
  const result: Result = { caseRevisionId: 'pinned', observation: { kind: 'proposal', referencedAssets: [], locations: [], executedOperations: [],
    proposals: [{ operation: 'create', newTitle: 'Box', newKind: 'item', targetId: '', destinationId: '', details: '' }] },
    verdict: { passed: false, failures: [{ code: 'proposal_mismatch', fixtureId: '', operation: 'create' }] }, modelCalls: 2, durationMilliseconds: 350, completedAt: '' };
  const unsupported = async (): Promise<never> => { throw new Error('Read only'); };
  const cases = { list: unsupported, create: unsupported, append: unsupported,
    get: async (_tenant: string, _case: string, version?: string) => { reads.push(version ?? 'latest'); return revision; } };
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunResult, { target: document.body, props: { session, cases, pin: { caseId: 'case', revisionId: 'pinned' }, result } });
  expect(reads).toEqual([]);
  document.querySelector('button')!.click();
  await expect.poll(() => reads).toEqual(['pinned']);
  await expect.poll(() => document.body.textContent).toContain('container');
  expect(document.body.textContent).toContain('(item)');
});
