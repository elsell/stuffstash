import { mount, tick, unmount } from 'svelte';
import { expect, it, vi } from 'vitest';
import InventoryConversationPanel from './InventoryConversationPanel.svelte';
import type { InventoryConversationEvent, InventoryConversationTransport } from '$lib/ports/inventoryConversation';
it('sends from the keyboard, reviews without automatic approval, and restores opener focus', async () => {
  const sent: string[] = []; const decisions: boolean[] = []; let closed = false;
  let emit: (event: InventoryConversationEvent) => void = () => {};
  const transport: InventoryConversationTransport = { async connect(_scope, _signal, onEvent) {
    emit = onEvent;
    return { send: (text) => { sent.push(text); }, decide: (_id, approve) => { decisions.push(approve); }, close: () => { closed = true; } };
  } };
  const component = mount(InventoryConversationPanel, { target: document.body, props: { transport, tenantId: 'home', inventoryId: 'main', inventoryName: 'Main Inventory', onRefresh: async () => true, onOpenAsset() {}, onAuthenticationLost() {} } });
  try {
    await tick();
    const opener = document.querySelector<HTMLButtonElement>('[aria-label="Ask Stuff Stash"]')!; opener.focus(); opener.click();
    await vi.waitFor(() => expect(document.querySelector('textarea')).not.toBeNull());
    const input = document.querySelector<HTMLTextAreaElement>('textarea')!;
    input.value = 'Move tent'; input.dispatchEvent(new Event('input', { bubbles: true })); await tick();
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', isComposing: true, bubbles: true })); await tick(); expect(sent).toEqual([]);
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await vi.waitFor(() => expect(sent).toEqual(['Move tent']));
    emit({ type: 'review', plan: { id: 'plan', summary: 'Move tent into Garage', commands: [{ summary: 'Move tent', destination: 'Garage', changes: ['Name: Camping tent', 'Description: ' + 'x'.repeat(300)], expiration: { date: '2028-02', precision: 'month' } }], risks: [] } });
    await tick(); expect(document.body.textContent).toContain('Review changes'); expect(decisions).toEqual([]); expect(document.body.textContent).toContain('Name: Camping tent'); expect(document.body.textContent).toContain('February 2028'); expect(document.body.textContent).toContain('Description: ' + 'x'.repeat(300));
    const cancel = Array.from(document.querySelectorAll('button')).find((button) => button.textContent?.trim() === 'Cancel changes')!; cancel.click(); await tick(); expect(decisions).toEqual([false]);
    emit({ type: 'cancelled' }); await tick();
    document.querySelector<HTMLElement>('[role="dialog"]')?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await vi.waitFor(() => expect(document.activeElement).toBe(opener)); expect(closed).toBe(true);
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
