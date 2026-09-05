import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { ConversationFailure } from '$lib/domain/conversation';
import type { CaseDefinition, CaseRevision } from '$lib/domain/conversationCase';
import type { ConversationCaseRepository } from '$lib/ports/conversationCaseRepository';
import CaseWorkspace from './CaseWorkspace.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
class Cases implements ConversationCaseRepository {
  revision: CaseRevision = { id: 'revision', caseId: 'case', number: 1, authorId: 'owner', createdAt: '2026-09-05T12:00:00Z',
    definition: { title: 'Find baby clothes', utterance: 'Where are my baby clothes?', assets: [], expectations: { kind: 'clarification', referencedAssets: [], locations: [], proposals: [], forbiddenOperations: ['create'] } } };
  async list() { return { items: [{ id: 'case', title: this.revision.definition.title, latestRevision: this.revision.number, latestRevisionId: this.revision.id, createdAt: this.revision.createdAt, updatedAt: this.revision.createdAt }], pagination: { limit: 20, hasMore: false, nextCursor: null } }; }
  async get() { return structuredClone(this.revision); }
  async create(_tenant: string, definition: CaseDefinition) { this.revision = { ...this.revision, definition }; return this.revision; }
  async append(_tenant: string, _id: string, expected: number, definition: CaseDefinition) {
    if (expected !== this.revision.number) throw new ConversationFailure('conflict');
    this.revision = { ...this.revision, id: 'revision-2', number: expected + 1, definition }; return this.revision;
  }
}
function button(text: string) { return Array.from(document.querySelectorAll('button')).find(button => button.textContent?.includes(text)); }
describe('case workspace', () => {
  it('opens a saved case and appends its edited revision', async () => {
    const cases = new Cases();
    component = mount(CaseWorkspace, { target: document.body, props: { scope: { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, cases } });
    await expect.poll(() => button('Find baby clothes')).toBeDefined(); button('Find baby clothes')!.click();
    await expect.poll(() => document.querySelector('input[name="case-title"]')).not.toBeNull();
    const title = document.querySelector<HTMLInputElement>('input[name="case-title"]')!;
    title.value = 'Locate winter clothes'; title.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
    await expect.poll(() => cases.revision.number).toBe(2); expect(cases.revision.definition.title).toBe('Locate winter clothes');
  });
  it('removes case content when configure access is denied', async () => {
    const cases = new Cases(); cases.list = async () => { throw new ConversationFailure('forbidden'); };
    component = mount(CaseWorkspace, { target: document.body, props: { scope: { apiIdentity: 'api', principalId: 'owner', tenantId: 'home' }, cases } });
    await expect.poll(() => document.body.textContent).toContain('Test cases unavailable'); expect(button('New test case')).toBeUndefined();
  });
});
