import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { ConversationFailure } from '$lib/domain/conversation';
import type { WorkflowDefinition, WorkflowRevision } from '$lib/domain/conversationWorkflow';
import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
import WorkflowWorkspace from './WorkflowWorkspace.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const definition: WorkflowDefinition = { name: 'Household', providerProfileId: null, instructions: '',
  budget: { toolCalls: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 } };
class Workflows implements ConversationWorkflowRepository {
  revision: WorkflowRevision = { id: 'revision', workflowId: 'workflow', number: 1, authorId: 'owner', createdAt: '2026-09-05T12:00:00Z', definition };
  async list() { return { items: [{ id: 'workflow', name: 'Household', latestRevision: this.revision.number, latestRevisionId: this.revision.id, activeRevisionId: null, createdAt: this.revision.createdAt, updatedAt: this.revision.createdAt }], pagination: { limit: 20, hasMore: false, nextCursor: null } }; }
  async history() { return { items: [this.revision], pagination: { limit: 20, hasMore: false, nextCursor: null } }; }
  async get() { return structuredClone(this.revision); }
  async selection() { return null; }
  async create(_tenant: string, value: WorkflowDefinition) { this.revision = { ...this.revision, definition: value }; return this.revision; }
  async append(_tenant: string, _workflow: string, expected: number, value: WorkflowDefinition) {
    if (expected !== this.revision.number) throw new ConversationFailure('conflict');
    this.revision = { ...this.revision, id: 'revision-2', number: expected + 1, definition: value }; return this.revision;
  }
  async activate() { return this.revision; }
}
function button(text: string) { return Array.from(document.querySelectorAll('button')).find(button => button.textContent?.includes(text)); }
describe('workflow workspace', () => {
  it('disables saving while loading a conflicting revision for comparison', async () => {
    const workflows = new Workflows(); let writes = 0;
    workflows.append = async () => { writes++; throw new ConversationFailure('conflict'); };
    component = mount(WorkflowWorkspace, { target: document.body, props: { scope: { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, workflows, providers: { list: async () => [] } } });
    await expect.poll(() => button('Household')).toBeDefined(); button('Household')!.click();
    await expect.poll(() => document.querySelector('form')).not.toBeNull();
    document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await expect.poll(() => document.body.textContent).toContain('newer revision');
    let finish!: (value: WorkflowRevision) => void;
    workflows.get = () => new Promise(done => { finish = done; });
    const compare = Array.from(document.querySelectorAll('button')).find(value => /latest/i.test(value.textContent ?? ''))!;
    compare.click();
    await expect.poll(() => button('Save draft')?.disabled).toBe(true);
    document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    expect(writes).toBe(1); finish(workflows.revision);
    await expect.poll(() => button('Save draft')?.disabled).toBe(false);
  });

  it('prevents a new draft from racing a pending revision load', async () => {
    const workflows = new Workflows();
    let resolve!: (value: WorkflowRevision) => void;
    workflows.get = () => new Promise(done => { resolve = done; });
    component = mount(WorkflowWorkspace, { target: document.body, props: { scope: { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, workflows, providers: { list: async () => [] } } });
    await expect.poll(() => button('Household')).toBeDefined(); button('Household')!.click();
    await expect.poll(() => button('New workflow')?.disabled).toBe(true);
    button('New workflow')!.click();
    resolve(workflows.revision);
    await expect.poll(() => document.querySelector<HTMLInputElement>('input[name="name"]')?.value).toBe('Household');
  });
  it('loads a saved workflow and appends a draft through its repository', async () => {
    const workflows = new Workflows();
    component = mount(WorkflowWorkspace, { target: document.body, props: { scope: { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, workflows, providers: { list: async () => [] } } });
    await expect.poll(() => button('Household')).toBeDefined(); button('Household')!.click();
    await expect.poll(() => document.querySelector('input[name="name"]')).not.toBeNull();
    const name = document.querySelector<HTMLInputElement>('input[name="name"]')!;
    name.value = 'Baby clothes'; name.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await expect.poll(() => workflows.revision.number).toBe(2);
    expect(workflows.revision.definition.name).toBe('Baby clothes');
  });
  it('clears the workspace when configure access is denied', async () => {
    const workflows = new Workflows();
    workflows.list = async () => { throw new ConversationFailure('forbidden'); };
    component = mount(WorkflowWorkspace, { target: document.body, props: { scope: { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, workflows, providers: { list: async () => [] } } });
    await expect.poll(() => document.body.textContent).toContain('Conversation settings unavailable');
    expect(button('New workflow')).toBeUndefined();
  });
});
