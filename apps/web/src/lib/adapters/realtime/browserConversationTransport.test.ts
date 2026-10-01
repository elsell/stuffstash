import { expect, it } from 'vitest';
import { BrowserConversationTransport, type ConversationSocket } from './browserConversationTransport';
class Socket implements ConversationSocket {
  onopen: (() => void) | null = null; onmessage: ((event: { data: unknown }) => void) | null = null;
  onerror: (() => void) | null = null; onclose: (() => void) | null = null;
  sent: Record<string, unknown>[] = []; closed = false;
  send(data: string) { this.sent.push(JSON.parse(data)); }
  close() { this.closed = true; }
  message(value: unknown) { this.onmessage?.({ data: JSON.stringify(value) }); }
}
it('authenticates in a frame then preserves scope and numbered text/approval messages', async () => {
  const socket = new Socket(); let url = ''; let protocol = '';
  const transport = new BrowserConversationTransport('https://api.example.test', () => 'secret', (target, subprotocol) => { url = target; protocol = subprotocol; return socket; });
  const events: unknown[] = []; const controller = new AbortController();
  const pending = transport.connect({ tenantId: 'home', inventoryId: 'main' }, controller.signal, (event) => events.push(event), () => {});
  await Promise.resolve(); socket.onopen?.();
  expect(url).toBe('wss://api.example.test/v1/realtime/voice'); expect(protocol).toBe('stuffstash.browser.v1');
  expect(socket.sent).toEqual([{ type: 'session.authenticate', authorization: 'Bearer secret' }]);
  socket.message({ type: 'session.authenticated' }); expect(socket.sent[1].inventoryId).toBe('main');
  socket.message({ type: 'session.started', seq: 1, sessionId: 'session' });
  const connection = await pending; connection.send('Find tent');
  expect(socket.sent[2]).toMatchObject({ type: 'text.input', seq: 2, sessionId: 'session', text: 'Find tent' });
  socket.message({ type: 'action.plan.proposed', seq: 2, sessionId: 'session', actionPlan: { planId: 'plan', confirmationSummary: 'Move tent', commands: [{ summary: 'Move tent' }], risks: [] } });
  connection.decide('plan', false); expect(socket.sent[3]).toMatchObject({ type: 'action.plan.cancel', seq: 3, planId: 'plan' });
  controller.abort(); expect(socket.closed).toBe(true);
  const count = events.length; socket.message({ type: 'session.completed', seq: 3, sessionId: 'session' }); expect(events).toHaveLength(count);
});
it('rejects response scope drift and retires the socket', async () => {
  const socket = new Socket(); let failed = false;
  const transport = new BrowserConversationTransport('https://api.example.test', () => 'secret', () => socket);
  const pending = transport.connect({ tenantId: 'home', inventoryId: 'main' }, new AbortController().signal, () => {}, () => { failed = true; });
  await Promise.resolve(); socket.onopen?.(); socket.message({ type: 'session.authenticated' }); socket.message({ type: 'session.started', seq: 1, sessionId: 'session' }); await pending;
  socket.message({ type: 'assistant.response.completed', seq: 2, sessionId: 'session', response: { tenantId: 'other', inventoryId: 'main', displayResponse: 'Private' } });
  expect(failed).toBe(true); expect(socket.closed).toBe(true);
});
it('does not sign the user out when the network fails before authentication', async () => {
  const socket = new Socket(); const failures: string[] = [];
  const transport = new BrowserConversationTransport('https://api.example.test', () => 'secret', () => socket);
  const pending = transport.connect({ tenantId: 'home', inventoryId: 'main' }, new AbortController().signal, () => {}, kind => failures.push(kind));
  const rejected = expect(pending).rejects.toThrow();
  await Promise.resolve(); socket.onerror?.(); await rejected;
  expect(failures).toEqual(['unavailable']);
});
