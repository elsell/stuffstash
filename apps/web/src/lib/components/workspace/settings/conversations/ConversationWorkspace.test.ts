import { afterEach, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { ConversationFailure } from '$lib/domain/conversation';
import type { WorkflowRevision } from '$lib/domain/conversationWorkflow';
import type { ConversationWorkspaceRepositories } from '$lib/ports/conversationWorkspace';
import ConversationWorkspace from './ConversationWorkspace.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const page = { items: [], pagination: { limit: 20, hasMore: false, nextCursor: null } };
const unsupported = async (): Promise<never> => { throw new Error('Unsupported fake operation'); };
function repositories(): ConversationWorkspaceRepositories {
  return { apiIdentity: 'api', providers: { list: async () => [] },
    workflows: { list: async () => ({ ...page, items: [{ id: 'workflow', name: 'Household', latestRevision: 1, latestRevisionId: 'revision', activeRevisionId: null, createdAt: '', updatedAt: '' }] }), history: async () => page, selection: async () => null, get: unsupported, create: unsupported, append: unsupported, activate: unsupported },
    cases: { list: async () => page, get: unsupported, create: unsupported, append: unsupported },
    runs: { list: async () => page, get: unsupported, queue: unsupported, cancel: unsupported } };
}
function button(text: string) { return Array.from(document.querySelectorAll('button')).find(value => value.textContent?.trim() === text); }
const scope = { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' };
it('keeps a pending workflow load in its visible section', async () => {
  const ports = repositories(); let reject!: (reason: Error) => void;
  ports.workflows.get = () => new Promise<WorkflowRevision>((_resolve, fail) => { reject = fail; });
  component = mount(ConversationWorkspace, { target: document.body, props: { scope, repositories: ports } });
  await expect.poll(() => Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('Household'))).toBeDefined();
  Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('Household'))!.click();
  await expect.poll(() => button('Test cases')?.disabled).toBe(true);
  button('Test cases')!.click(); expect(button('Workflows')?.getAttribute('aria-pressed')).toBe('true');
  reject(new ConversationFailure('invalid'));
  await expect.poll(() => button('Test cases')?.disabled).toBe(false);
});
it('removes every section when case access is revoked', async () => {
  const ports = repositories(); ports.cases.list = async () => { throw new ConversationFailure('forbidden'); };
  component = mount(ConversationWorkspace, { target: document.body, props: { scope, repositories: ports } });
  await expect.poll(() => button('Test cases')).toBeDefined(); button('Test cases')!.click();
  await expect.poll(() => document.body.textContent).toContain('Your account no longer has access');
  expect(button('Workflows')).toBeUndefined(); expect(document.body.textContent).not.toContain('Household');
});
it('refreshes run setup after saving a revision in the workflow section', async () => {
  const ports = repositories();
  let revision: WorkflowRevision = { id: 'revision-1', workflowId: 'workflow', number: 1, authorId: 'owner', createdAt: '', definition: {
    name: 'Household', retrieval: 'precise_first', response: 'grounded', budget: { evidenceRounds: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 }, steps: [
      { kind: 'interpret', attempts: 1, instructions: '', providerProfileId: null }, { kind: 'assess', attempts: 1, instructions: '', providerProfileId: null }, { kind: 'respond', attempts: 1, instructions: '', providerProfileId: null }] } };
  ports.workflows.list = async () => ({ ...page, items: [{ id: 'workflow', name: 'Household', latestRevision: revision.number, latestRevisionId: revision.id, activeRevisionId: null, createdAt: '', updatedAt: '' }] });
  ports.workflows.get = async () => structuredClone(revision);
  ports.workflows.append = async (_tenant, _workflow, _expected, definition) => { revision = { ...revision, number: 2, id: 'revision-2', definition }; return revision; };
  component = mount(ConversationWorkspace, { target: document.body, props: { scope, repositories: ports } });
  await expect.poll(() => button('Runs')).toBeDefined(); button('Runs')!.click();
  await expect.poll(() => button('New run')).toBeDefined(); button('New run')!.click();
  await expect.poll(() => document.querySelector('[aria-label="Choose workflow"]')?.textContent).toContain('Revision 1');
  button('Discard run setup')!.click();
  await expect.poll(() => button('Workflows')?.disabled).toBe(false); button('Workflows')!.click();
  const select = () => Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('Household'));
  await expect.poll(select).toBeDefined(); select()!.click();
  await expect.poll(() => document.querySelector('form')).not.toBeNull();
  document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
  await expect.poll(() => revision.number).toBe(2);
  const close = () => Array.from(document.querySelectorAll('button')).find(value => value.textContent?.includes('Close editor'));
  await expect.poll(() => close()?.disabled).toBe(false); close()!.click();
  await expect.poll(() => button('Runs')?.disabled).toBe(false); button('Runs')!.click(); button('New run')!.click();
  await expect.poll(() => document.querySelector('[aria-label="Choose workflow"]')?.textContent).toContain('Revision 2');
});
