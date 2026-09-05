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
