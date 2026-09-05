import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import CaseFixtures from './CaseFixtures.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const value: CaseDefinition = { title: 'Clothes', utterance: 'Find clothes', assets: [
  { id: 'clothes', title: '3–6 months', kind: 'item', description: '', parentId: '', tagNames: ['baby'] }
], expectations: { kind: 'answer', referencedAssets: ['clothes'], locations: [], proposals: [], forbiddenOperations: [] } };
describe('case fixtures', () => {
  it('preserves fixture identity and commas in edited tags', () => {
    const changes: CaseDefinition[] = [];
    component = mount(CaseFixtures, { target: document.body, props: { value, onChange: changed => changes.push(changed) } });
    const tags = document.querySelector<HTMLTextAreaElement>('textarea[name="fixture-tags-clothes"]')!;
    tags.value = 'baby\nclothes, winter'; tags.dispatchEvent(new Event('input', { bubbles: true }));
    expect(changes[0].assets[0]).toMatchObject({ id: 'clothes', tagNames: ['baby', 'clothes, winter'] });
    expect(value.assets[0].tagNames).toEqual(['baby']);
    const remove = Array.from(document.querySelectorAll('button')).find(button => button.textContent?.includes('Remove fixture'))!;
    expect(remove.disabled).toBe(true);
  });
});
