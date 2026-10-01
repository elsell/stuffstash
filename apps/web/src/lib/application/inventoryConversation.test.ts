import { describe, expect, it } from 'vitest';
import { InventoryConversation } from './inventoryConversation';
import type { InventoryConversationConnection, InventoryConversationEvent, InventoryConversationTransport } from '$lib/ports/inventoryConversation';
class ConversationServer implements InventoryConversationTransport {
  event: (event: InventoryConversationEvent) => void = () => {};
  failure: (kind: 'authentication' | 'unavailable') => void = () => {};
  decisions: { id: string; approve: boolean }[] = []; texts: string[] = []; closed = false;
  async connect(_scope: unknown, signal: AbortSignal, event: typeof this.event, failure: typeof this.failure): Promise<InventoryConversationConnection> {
    this.event = event; this.failure = failure;
    return { send: (text) => { this.texts.push(text); }, decide: (id, approve) => { this.decisions.push({ id, approve }); }, close: () => { this.closed = true; } };
  }
}
describe('inventory conversation', () => {
  it('requires an explicit plan decision and never retries an uncertain approval', async () => {
    const server = new ConversationServer(); let refreshed = 0;
    const conversation = new InventoryConversation(server, { tenantId: 'home', inventoryId: 'main' }, () => {}, () => { refreshed++; }, () => {});
    await conversation.send('Move the tent');
    server.event({ type: 'review', plan: { id: 'plan', summary: 'Move tent', commands: [{ summary: 'Move tent into Garage' }], risks: [] } });
    expect(server.decisions).toEqual([]);
    conversation.decide(true); conversation.decide(true);
    expect(server.decisions).toEqual([{ id: 'plan', approve: true }]);
    server.failure('unavailable');
    expect(conversation.state.uncertain).toBe(true);
    await conversation.send('Try again');
    expect(server.texts).toEqual(['Move the tent']);
    expect(refreshed).toBe(1);
    await conversation.recover();
    expect(conversation.state.uncertain).toBe(false);
  });
  it('retires callbacks on scope teardown and cancels a plan without approving', async () => {
    const server = new ConversationServer();
    const conversation = new InventoryConversation(server, { tenantId: 'home', inventoryId: 'main' }, () => {}, () => {}, () => {});
    await conversation.send('Create tent');
    server.event({ type: 'review', plan: { id: 'plan', summary: 'Create tent', commands: [{ summary: 'Create tent' }], risks: [] } });
    conversation.decide(false);
    expect(server.decisions).toEqual([{ id: 'plan', approve: false }]);
    conversation.dispose(); server.event({ type: 'answer', text: 'Private late answer', assets: [] });
    expect(conversation.state.messages).toEqual([]);
    expect(server.closed).toBe(true);
  });
});
