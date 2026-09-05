import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import { ConversationFailure } from '$lib/domain/conversation';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import CaseEditor from './CaseEditor.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const initial: CaseDefinition = { title: ' Find clothes ', utterance: 'Where are my baby clothes?', assets: [
  { id: 'clothes', title: '3–6 months', kind: 'item', parentId: '', description: '', tagNames: [' baby ', '', 'Baby'] }
], expectations: { kind: 'answer', referencedAssets: ['clothes'], locations: [], proposals: [], forbiddenOperations: ['create'] } };
function submit() { document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); }
describe('case editor', () => {
  it('saves a normalized case while preserving the source revision', async () => {
    const saved: CaseDefinition[] = [];
    component = mount(CaseEditor, { target: document.body, props: { initial, onSave: async value => { saved.push(value); } } });
    submit(); await expect.poll(() => saved.length).toBe(1);
    expect(saved[0].assets[0].tagNames).toEqual(['baby']); expect(saved[0].title).toBe('Find clothes');
    expect(initial.title).toBe(' Find clothes ');
  });
  it('focuses a linked validation summary without dispatching invalid input', async () => {
    let saved = false;
    component = mount(CaseEditor, { target: document.body, props: { initial: { ...initial, title: '' }, onSave: async () => { saved = true; } } });
    submit(); await expect.poll(() => document.activeElement?.getAttribute('role')).toBe('alert');
    expect(saved).toBe(false);
    const link = document.querySelector<HTMLAnchorElement>('[role="alert"] a')!; link.click();
    expect(document.activeElement?.getAttribute('name')).toBe('case-title');
  });
  it('preserves the case when a newer revision prevents saving', async () => {
    component = mount(CaseEditor, { target: document.body, props: { initial, onSave: async () => { throw new ConversationFailure('conflict'); } } });
    submit(); await expect.poll(() => document.body.textContent).toContain('newer revision');
    expect(document.querySelector<HTMLInputElement>('input[name="case-title"]')!.value).toBe(initial.title);
  });
});
