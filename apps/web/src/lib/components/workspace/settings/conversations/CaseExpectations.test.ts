import { afterEach, describe, expect, it } from 'vitest';
import { flushSync, mount, unmount } from 'svelte';
import type { CaseDefinition } from '$lib/domain/conversationCase';
import CaseExpectations from './CaseExpectations.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(async () => { if (component) await unmount(component); component = undefined; document.body.innerHTML = ''; });
const value: CaseDefinition = { title: 'Clothes', utterance: 'Find baby clothes', assets: [
  { id: 'clothes', title: '3–6 months', kind: 'item', description: '', parentId: '', tagNames: ['baby'] }
], expectations: { kind: 'answer', referencedAssets: [], locations: [], proposals: [], forbiddenOperations: [] } };
function button(text: string) { return Array.from(document.querySelectorAll('button')).find(button => button.textContent?.trim() === text)!; }
describe('case expectations', () => {
  it.each(['Create', 'Move', 'Archive', 'Restore', 'Check out', 'Return'])('edits the %s proposal through its selector', async label => {
    const changes: CaseDefinition[] = [];
    const initial = { operation: label === 'Create' ? 'move' as const : 'create' as const, targetId: '', destinationId: '', newKind: '' as const, newTitle: 'Old title', details: 'Old details' };
    component = mount(CaseExpectations, { target: document.body, props: { value: { ...value, expectations: { ...value.expectations, kind: 'proposal', proposals: [initial] } }, onChange: changed => changes.push(changed) } });
    document.querySelector<HTMLButtonElement>('#proposal-operation-0')!.click(); flushSync();
    const option = () => Array.from(document.querySelectorAll<HTMLElement>('[role="option"]')).find(item => item.textContent?.trim() === label);
    await expect.poll(option).toBeDefined(); option()!.click(); flushSync();
    await expect.poll(() => changes.length).toBeGreaterThan(0);
    expect(changes.at(-1)!.expectations.proposals[0]).toEqual({ operation: label === 'Check out' ? 'checkout' : label.toLowerCase(), targetId: '', destinationId: '', newTitle: '', newKind: label === 'Create' ? 'item' : '', details: '' });
  });
  it('lets a homeowner require an existing fixture and forbid creation', () => {
    const changes: CaseDefinition[] = [];
    component = mount(CaseExpectations, { target: document.body, props: { value, onChange: changed => changes.push(changed) } });
    button('3–6 months').click();
    expect(changes[0].expectations.referencedAssets).toEqual(['clothes']);
    button('Create').click();
    expect(changes[1].expectations.forbiddenOperations).toEqual(['create']);
    expect(value.expectations.forbiddenOperations).toEqual([]);
  });
  it('keeps incompatible proposals visible for deliberate removal', () => {
    const changes: CaseDefinition[] = [];
    const proposal = { operation: 'create' as const, targetId: '', destinationId: '', newKind: 'item' as const, newTitle: 'Clothes', details: '' };
    component = mount(CaseExpectations, { target: document.body, props: { value: { ...value, expectations: { ...value.expectations, proposals: [proposal] } }, onChange: changed => changes.push(changed) } });
    button('Remove proposed change').click();
    expect(changes[0].expectations.proposals).toEqual([]);
  });
});
