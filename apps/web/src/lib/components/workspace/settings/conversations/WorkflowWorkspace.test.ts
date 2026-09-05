import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { ConversationFailure } from '$lib/domain/conversation';
import type { WorkflowDefinition, WorkflowRevision } from '$lib/domain/conversationWorkflow';
import type { ConversationWorkflowRepository } from '$lib/ports/conversationWorkflowRepository';
import WorkflowWorkspace from './WorkflowWorkspace.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const definition: WorkflowDefinition = { name: 'Household', retrieval: 'precise_first', response: 'grounded',
  budget: { evidenceRounds: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 }, steps: [
    { kind: 'interpret', attempts: 1, instructions: '', providerProfileId: null },
    { kind: 'assess', attempts: 1, instructions: '', providerProfileId: null },
    { kind: 'respond', attempts: 1, instructions: '', providerProfileId: null }] };
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
