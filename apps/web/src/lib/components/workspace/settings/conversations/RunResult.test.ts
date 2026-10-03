import { afterEach, expect, it } from 'vitest';
import { localization, t } from '$lib/presentation/localization';
import { mount, unmount } from 'svelte';
import { createConversationSession, type ConversationSession } from '$lib/adapters/query/conversationSession';
import type { CaseRevision } from '$lib/domain/conversationCase';
import type { RunResult as Result } from '$lib/domain/conversationRun';
import RunResult from './RunResult.svelte';
let component: ReturnType<typeof mount> | undefined; let session: ConversationSession;
afterEach(async () => { if (component) await unmount(component); component = undefined; await session?.dispose(); document.body.innerHTML = ''; });
it.each([0, 1, 2])('presents the pinned result with %i model calls', async (modelCalls) => {
  const reads: string[] = [];
  const revision: CaseRevision = { id: 'pinned', caseId: 'case', number: 1, authorId: 'owner', createdAt: '', definition: {
    title: 'Another box', utterance: 'Add another box', assets: [{ id: 'tent', title: 'My tent', kind: 'item', description: '', parentId: '', tagNames: [] }, { id: 'bin', title: 'Summer bin', kind: 'container', description: '', parentId: '', tagNames: [] }], expectations: { kind: 'proposal', referencedAssets: [], locations: [], forbiddenOperations: [],
      proposals: [{ operation: 'create', newTitle: 'Box', newKind: 'container', targetId: '', destinationId: '', details: '' }] } } };
  const result: Result = { caseRevisionId: 'pinned', observation: { kind: 'proposal', referencedAssets: ['tent', 'bin'], locations: [], executedOperations: ['create', 'move'],
    proposals: [{ operation: 'create', newTitle: 'Box', newKind: 'item', targetId: '', destinationId: '', details: '' }] },
    verdict: { passed: false, failures: [{ code: 'missing_proposal', fixtureId: '', operation: 'create' }, { code: 'constructor', fixtureId: '', operation: '' }] }, modelCalls, durationMilliseconds: 1250, completedAt: '' };
  const unsupported = async (): Promise<never> => { throw new Error('Read only'); };
  const cases = { list: unsupported, create: unsupported, append: unsupported,
    get: async (_tenant: string, _case: string, version?: string) => { reads.push(version ?? 'latest'); return revision; } };
  session = createConversationSession({ apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, () => {});
  component = mount(RunResult, { target: document.body, props: { session, cases, pin: { caseId: 'case', revisionId: 'pinned' }, result } });
  expect(reads).toEqual([]);
  expect(document.body.textContent).toContain(t('evaluation.resultSummary', { verdict: t('evaluation.failed'), count: modelCalls, seconds: localization.number(1.25, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) }));
  document.querySelector('button')!.click();
  await expect.poll(() => reads).toEqual(['pinned']);
  await expect.poll(() => document.body.textContent).toContain(t('asset.kind.container'));
  expect(document.body.textContent).toContain('(' + t('asset.kind.item') + ')');
  expect(document.body.textContent).toContain(t('case.operation.create'));
  expect(document.body.textContent).toContain(t('case.outcome.proposal'));
  expect(document.body.textContent).toContain(t('web.RunResult.executedOperationsFull', { value: localization.list([t('case.operation.create'), t('case.operation.move')]) }));
  expect(document.body.textContent).toContain(t('evaluation.referencedItems', { items: localization.list(['My tent', 'Summer bin']) }));
  expect(document.body.textContent).toContain(t('evaluation.failure.missing_proposal'));
  expect(document.body.textContent).toContain(t('evaluation.failure.unknown'));
  expect(result.verdict.failures[0].code).toBe('missing_proposal');
  expect(result.observation.proposals[0].operation).toBe('create');
  expect(result.observation.proposals[0].newKind).toBe('item');
});
