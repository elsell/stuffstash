import { validExpirationInput } from '$lib/domain/expiration';
import type { AssetExpiration } from '$lib/domain/inventory';
import type { InventoryConversationConnection, InventoryConversationEvent, InventoryConversationScope, InventoryConversationTransport } from '$lib/ports/inventoryConversation';
export type ConversationSocket = Pick<WebSocket, 'onopen' | 'onmessage' | 'onerror' | 'onclose' | 'send' | 'close'>;
type SocketFactory = (url: string, protocol: string) => ConversationSocket;
const object = (value: unknown): value is Record<string, unknown> => value !== null && typeof value === 'object' && !Array.isArray(value);
const text = (value: unknown): string => typeof value === 'string' ? value : '';
export class BrowserConversationTransport implements InventoryConversationTransport {
  constructor(private baseUrl: string, private token: () => string | null | Promise<string | null>,
    private socketFactory: SocketFactory = (url, protocol) => new WebSocket(url, protocol)) {}
  async connect(scope: InventoryConversationScope, signal: AbortSignal, onEvent: (event: InventoryConversationEvent) => void,
    onFailure: (kind: 'authentication' | 'unavailable') => void): Promise<InventoryConversationConnection> {
    const token = await this.token(); signal.throwIfAborted();
    if (!token) { onFailure('authentication'); throw new Error('Authentication required'); }
    const url = new URL(this.baseUrl); url.pathname = `${url.pathname.replace(/\/$/, '')}/v1/realtime/voice`; url.search = ''; url.hash = '';
    if (url.protocol !== 'https:' && url.protocol !== 'http:') throw new Error('Invalid realtime endpoint');
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = this.socketFactory(url.toString(), 'stuffstash.browser.v1');
    return new Promise((resolve, reject) => {
      let retired = false; let authenticated = false; let started = false; let session = ''; let seq = 1; let last = 0; let plan = '';
      const timer = setTimeout(() => fail('unavailable'), 15_000);
      const close = () => { if (retired) return; retired = true; clearTimeout(timer); signal.removeEventListener('abort', abort); socket.onopen = null; socket.onmessage = null; socket.onerror = null; socket.onclose = null; socket.close(); if (!started) reject(new DOMException('Conversation closed', 'AbortError')); };
      const fail = (kind: 'authentication' | 'unavailable') => { if (retired) return; close(); onFailure(kind); };
      const abort = () => close(); signal.addEventListener('abort', abort, { once: true });
      const send = (value: Record<string, unknown>) => { if (retired) throw new Error('Conversation closed'); socket.send(JSON.stringify(value)); };
      const connection: InventoryConversationConnection = {
        send: (raw) => { const value = raw.trim(); if (!started || plan || !value || [...value].length > 8000) throw new Error('Invalid conversation turn'); send({ type: 'text.input', seq: seq++, sessionId: session, text: value }); },
        decide: (id, approve) => { if (!plan || id !== plan) throw new Error('Review expired'); plan = ''; send({ type: approve ? 'action.plan.approve' : 'action.plan.cancel', seq: seq++, sessionId: session, planId: id }); },
        close: () => { if (started && !retired) { try { send({ type: 'session.cancel', seq: seq++, sessionId: session, reason: 'user_cancelled' }); } catch { /* Best effort on a closing connection. */ } } close(); }
      };
      socket.onopen = () => { try { send({ type: 'session.authenticate', authorization: `Bearer ${token}` }); } catch { fail('unavailable'); } };
      socket.onerror = () => fail('unavailable');
      socket.onclose = (event) => fail(!authenticated && event.code === 1008 ? 'authentication' : 'unavailable');
      socket.onmessage = ({ data }) => {
        if (retired) return;
        try {
          if (typeof data !== 'string' || data.length > 1024 * 1024) throw new Error('Invalid frame');
          const message: unknown = JSON.parse(data); if (!object(message)) throw new Error('Invalid frame');
          if (message.type === 'session.authenticated') {
            if (authenticated) throw new Error('Repeated authentication'); authenticated = true;
            send({ type: 'session.start', seq: seq++, tenantId: scope.tenantId, inventoryId: scope.inventoryId, source: 'web_text', conversationContinuity: true,
              requestedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech'], inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 }, outputAudio: { mimeTypes: ['audio/mpeg'] } }); return;
          }
          if (!authenticated || !Number.isSafeInteger(message.seq) || (message.seq as number) <= last) throw new Error('Invalid sequence'); last = message.seq as number;
          if (message.type === 'session.failed') { fail(message.code === 'unauthenticated' ? 'authentication' : 'unavailable'); return; }
          if (message.type === 'session.started') {
            if (started || !text(message.sessionId)) throw new Error('Invalid session'); session = text(message.sessionId); started = true; clearTimeout(timer); resolve(connection); return;
          }
          if (!started || message.sessionId !== session) throw new Error('Wrong session');
          if (message.type === 'assistant.response.completed') {
            const response = message.response;
            if (!object(response) || response.tenantId !== scope.tenantId || response.inventoryId !== scope.inventoryId || response.sessionId !== session || !text(response.displayResponse)) throw new Error('Invalid response scope');
            const artifacts = Array.isArray(response.artifacts) ? response.artifacts : [];
            const assets = artifacts.filter(object).filter((asset) => asset.type === 'asset_reference' && text(asset.assetId) && text(asset.title)).map((asset) => ({ id: text(asset.assetId), title: text(asset.title) }));
            onEvent({ type: 'answer', text: text(response.displayResponse), assets });
          } else if (message.type === 'action.plan.proposed') {
            const value = message.actionPlan;
            if (plan || !object(value) || !text(value.planId) || !text(value.confirmationSummary) || !Array.isArray(value.commands) || !value.commands.length || !value.commands.every((command) => object(command) && text(command.summary) && (command.changes === undefined || validChanges(command.changes)) && validExpirationReview(command))) throw new Error('Invalid review');
            plan = text(value.planId); onEvent({ type: 'review', plan: { id: plan, summary: text(value.confirmationSummary), commands: value.commands.filter(object).map((command) => ({ summary: text(command.summary), title: text(command.title), destination: text(command.parentTitle), ...(validExpiration(command.expiration) ? { expiration: { ...command.expiration } } : {}), ...(command.expirationCleared === true ? { expirationCleared: true } : {}), ...(validChanges(command.changes) ? { changes: [...command.changes] } : {}) })), risks: Array.isArray(value.risks) ? value.risks.filter((risk): risk is string => typeof risk === 'string') : [] } });
          } else if (message.type === 'action.plan.executed') { onEvent({ type: 'changed' }); close(); onEvent({ type: 'ended' }); }
          else if (message.type === 'action.plan.cancelled' || message.type === 'session.cancelled') { onEvent({ type: 'cancelled' }); close(); onEvent({ type: 'ended' }); }
          else if (message.type === 'action.plan.failed') fail('unavailable');
          else if (message.type === 'session.completed') {
            if (plan) return;
            if (message.followUpAvailable === true) onEvent({ type: 'ready' }); else { close(); onEvent({ type: 'ended' }); }
          }
        } catch { fail('unavailable'); }
      };
      if (signal.aborted) close();
    });
  }
}

function validChanges(value: unknown): value is string[] {
 return Array.isArray(value) && value.length > 0 && value.length <= 12 && value.every(change => typeof change === 'string' && change.trim().length > 0 && change.length <= 4608);
}

function validExpiration(value: unknown): value is AssetExpiration {
 return object(value) && typeof value.date === 'string' && value.date.length > 0 && (value.precision === 'day' || value.precision === 'month') && validExpirationInput(value.date, value.precision);
}
function validExpirationReview(command: Record<string, unknown>): boolean {
 return (command.expiration === undefined || validExpiration(command.expiration)) &&
   (command.expirationCleared === undefined || command.expirationCleared === false || (command.expirationCleared === true && command.expiration === undefined && command.kind === 'update_asset'));
}
