import { afterEach, describe, expect, it } from 'vitest';
import { flushSync, mount, unmount } from 'svelte';
import { ConversationFailure } from '$lib/domain/conversation';
import type { WorkflowDefinition } from '$lib/domain/conversationWorkflow';
import WorkflowEditor from './WorkflowEditor.svelte';
let component: ReturnType<typeof mount> | undefined;
afterEach(() => { if (component) unmount(component); component = undefined; document.body.innerHTML = ''; });
const definition: WorkflowDefinition = { name: 'Home', retrieval: 'expanded', response: 'grounded',
  budget: { evidenceRounds: 3, modelCalls: 8, elapsedSeconds: 45, followUpTurns: 3 },
  steps: ['interpret', 'assess', 'respond'].map(kind => ({ kind: kind as 'interpret' | 'assess' | 'respond', attempts: 1, instructions: '', providerProfileId: null })) };
function input(name: string) { return document.querySelector<HTMLInputElement>(`[name="${name}"]`)!; }
async function submit() { document.querySelector('form')!.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); await Promise.resolve(); await Promise.resolve(); flushSync(); }
describe('workflow editor', () => {
  it('offers one conversation model and guidance without retired stage controls', () => {
    component = mount(WorkflowEditor, { target: document.body, props: { initial: definition, providers: [], onSave: async () => {} } });
    expect(document.querySelector('#provider-model')).not.toBeNull();
    expect(document.querySelector('[name="instructions"]')).not.toBeNull();
    expect(document.querySelector('[name="toolCalls"]')).not.toBeNull();
    expect(document.querySelector('#provider-respond')).toBeNull();
    expect(document.querySelector('#retrieval')).toBeNull();
    expect(document.querySelector('#response')).toBeNull();
    expect(document.querySelector('[name="attempts-interpret"]')).toBeNull();
  });
  it('focuses validation feedback while keeping the draft', async () => {
    component = mount(WorkflowEditor, { target: document.body, props: { initial: definition, providers: [],
      onSave: async () => { throw new ConversationFailure('invalid'); } } });
    await submit();
    await Promise.resolve(); flushSync();
    await expect.poll(() => document.activeElement?.getAttribute('role')).toBe('alert');
    expect(input('name').value).toBe('Home');
  });
  it('saves edited settings without mutating the loaded revision', async () => {
    const saves: WorkflowDefinition[] = [];
    component = mount(WorkflowEditor, { target: document.body, props: { initial: definition, providers: [], onSave: async value => { saves.push(value); } } });
    input('name').value = 'Baby clothes'; input('name').dispatchEvent(new Event('input', { bubbles: true })); flushSync();
    await submit();
    expect(saves[0].name).toBe('Baby clothes'); expect(definition.name).toBe('Home');
    expect(document.body.textContent).toContain('Draft saved');
    expect(document.querySelector<HTMLButtonElement>('#provider-respond')!.disabled).toBe(true);
  });
  it('preserves the draft and offers latest-revision reload after a conflict', async () => {
    let reloads = 0;
    component = mount(WorkflowEditor, { target: document.body, props: { initial: definition, providers: [],
      onSave: async () => { throw new ConversationFailure('conflict'); }, onReload: () => { reloads++; } } });
    input('name').value = 'My draft'; input('name').dispatchEvent(new Event('input', { bubbles: true })); flushSync();
    await submit();
    expect(input('name').value).toBe('My draft');
    expect(document.body.textContent).toContain('newer revision');
    Array.from(document.querySelectorAll('button')).find(button => button.textContent?.includes('Load latest'))!.click();
    expect(reloads).toBe(1);
  });
});
