import { describe, expect, it } from 'vitest';
import { WebSocketRealtimeVoiceTransport } from './WebSocketRealtimeVoiceTransport';

describe('WebSocketRealtimeVoiceTransport', () => {
  it('opens the API realtime voice socket and sends start, audio chunks, and end messages', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: (url, headers) => {
        socket.url = url;
        socket.headers = headers;
        return socket;
      }
    });
    const events: unknown[] = [];
    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });
    socket.receive({ type: 'transcript.delta', seq: 2, sessionId: 'session-1', text: 'Where are' });
    socket.receive({ type: 'transcript.final', seq: 3, sessionId: 'session-1', text: 'Where are my tools?' });
    socket.receive({ type: 'assistant.response.started', seq: 4, sessionId: 'session-1', responseId: 'response-1' });
    socket.receive({
      type: 'action.plan.proposed',
      seq: 5,
      sessionId: 'session-1',
      actionPlan: {
        planId: 'plan-1',
        confirmationSummary: 'Create item water bottle?',
        commands: [{
          id: 'cmd-water-bottle',
          kind: 'create_asset',
          summary: 'Create item water bottle',
          operation: 'create',
          title: 'Water bottle',
          assetKind: 'item',
          parentTitle: 'Kitchen'
        }],
        risks: ['Adds a new item to this inventory.']
      }
    });
    socket.receive({ type: 'tts.audio.started', seq: 6, sessionId: 'session-1', format: { mimeType: 'audio/mpeg' } });
    socket.receive({ type: 'tts.audio.chunk', seq: 7, sessionId: 'session-1', chunkId: 'tts-1', audioBase64: 'c3BlZWNo' });
    socket.receive({ type: 'session.completed', seq: 8, sessionId: 'session-1' });
    await waitForSentMessageCount(socket, 3);
    await waitForEventType(events, 'action.plan.proposed');
    await transport.approveActionPlan('plan-1');
    await expect(transport.cancelActionPlan('plan-1')).rejects.toThrow('not active');
    socket.receive({ type: 'action.plan.approved', seq: 9, sessionId: 'session-1', planId: 'plan-1', status: 'approved' });
    socket.receive({
      type: 'action.plan.executed',
      seq: 10,
      sessionId: 'session-1',
      planId: 'plan-1',
      status: 'executed',
      message: 'The approved change was applied.',
      commandResults: [{
        commandId: 'cmd-water-bottle',
        assetId: 'asset-water-bottle',
        operation: 'create',
        assetKind: 'item'
      }]
    });

    await run;

    expect(socket.url).toBe('ws://127.0.0.1:8080/v1/realtime/voice');
    expect(socket.headers).toEqual({ Authorization: 'Bearer dev:user-1' });
    expect(socket.sent.map((message) => message.type)).toEqual(['session.start', 'audio.chunk', 'audio.end', 'action.plan.approve']);
    expect(socket.sent[0]).toMatchObject({
      seq: 1,
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice'
    });
    expect(socket.sent[1]).toMatchObject({
      seq: 2,
      sessionId: 'session-1',
      audioBase64: 'YXVkaW8=',
      isFinalChunk: true
    });
    expect(socket.sent[3]).toMatchObject({
      seq: 4,
      sessionId: 'session-1',
      planId: 'plan-1'
    });
    expect(events).toEqual([
      { type: 'session.started', seq: 1, sessionId: 'session-1' },
      { type: 'transcript.delta', seq: 2, sessionId: 'session-1', text: 'Where are' },
      { type: 'transcript.final', seq: 3, sessionId: 'session-1', text: 'Where are my tools?' },
      { type: 'assistant.response.started', seq: 4, sessionId: 'session-1', responseId: 'response-1' },
      {
        type: 'action.plan.proposed',
        seq: 5,
        sessionId: 'session-1',
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Create item water bottle?',
          commands: [{
            id: 'cmd-water-bottle',
            kind: 'create_asset',
            summary: 'Create item water bottle',
            operation: 'create',
            title: 'Water bottle',
            assetKind: 'item',
            parentTitle: 'Kitchen'
          }],
          risks: ['Adds a new item to this inventory.']
        }
      },
      { type: 'tts.audio.started', seq: 6, sessionId: 'session-1', mimeType: 'audio/mpeg' },
      { type: 'tts.audio.chunk', seq: 7, sessionId: 'session-1', chunkId: 'tts-1', audioBase64: 'c3BlZWNo' },
      { type: 'session.completed', seq: 8, sessionId: 'session-1' },
      { type: 'action.plan.approved', seq: 9, sessionId: 'session-1', planId: 'plan-1', status: 'approved', message: undefined },
      {
        type: 'action.plan.executed',
        seq: 10,
        sessionId: 'session-1',
        planId: 'plan-1',
        status: 'executed',
        message: 'The approved change was applied.',
        commandResults: [{
          commandId: 'cmd-water-bottle',
          assetId: 'asset-water-bottle',
          operation: 'create',
          assetKind: 'item'
        }]
      }
    ]);
  });

  it('rejects stale or cross-session server events before forwarding them', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      diagnosticsEnabled: true,
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });
    socket.receive({ type: 'transcript.final', seq: 2, sessionId: 'session-other', text: 'Wrong session' });

    await expect(run).rejects.toThrow('session');
    expect(events).toEqual([{ type: 'session.started', seq: 1, sessionId: 'session-1' }]);
  });

  it('forwards verbose agent diagnostic and tool detail events', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      diagnosticsEnabled: true,
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });
    socket.receive({
      type: 'agent.diagnostic',
      seq: 2,
      sessionId: 'session-1',
      message: 'Language prompt',
      detail: 'Transcript: move my water bottle'
    });
    socket.receive({
      type: 'tool.call.started',
      seq: 3,
      sessionId: 'session-1',
      toolCallId: 'tool-1',
      toolLabel: 'Inventory list',
      status: 'searching',
      detail: '{\n  "locationTitle": "Kitchen"\n}'
    });
    socket.receive({ type: 'session.completed', seq: 4, sessionId: 'session-1' });

    await run;
    expect(socket.sent[0]).toMatchObject({ type: 'session.start', developerDiagnostics: true });
    expect(events).toContainEqual({
      type: 'agent.diagnostic',
      seq: 2,
      sessionId: 'session-1',
      message: 'Language prompt',
      detail: 'Transcript: move my water bottle'
    });
    expect(events).toContainEqual({
      type: 'tool.call.started',
      seq: 3,
      sessionId: 'session-1',
      toolCallId: 'tool-1',
      toolLabel: 'Inventory list',
      status: 'searching',
      code: undefined,
      message: undefined,
      detail: '{\n  "locationTitle": "Kitchen"\n}'
    });
  });

  it('forwards safe session failure events as terminal states', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.failed', seq: 1, code: 'invalid_request', message: 'No provider.' });

    await run;
    expect(events).toEqual([{ type: 'session.failed', seq: 1, code: 'invalid_request', message: 'No provider.' }]);
  });

  it('sends session cancel and rejects safely when cancelled after session start', async () => {
    const socket = new FakeWebSocket();
    const controller = new AbortController();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
      if (event.type === 'session.started') {
        controller.abort();
      }
    }, { signal: controller.signal });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });

    await expect(run).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(socket.sent.at(-1)).toMatchObject({
      type: 'session.cancel',
      sessionId: 'session-1',
      reason: 'user_cancelled'
    });
    expect(socket.closedByClient).toBe(true);

    socket.receive({ type: 'transcript.final', seq: 2, sessionId: 'session-1', text: 'late transcript' });
    await Promise.resolve();
    expect(events).toEqual([{ type: 'session.started', seq: 1, sessionId: 'session-1' }]);
  });

  it('still settles cancellation when the socket rejects the best-effort cancel send', async () => {
    const socket = new FakeWebSocket();
    const controller = new AbortController();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      if (event.type === 'session.started') {
        socket.failNextSend = true;
        controller.abort();
      }
    }, { signal: controller.signal });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });

    await expect(run).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(socket.closedByClient).toBe(true);
  });


  it('forwards server cancellation as a terminal event', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });
    socket.receive({ type: 'session.cancelled', seq: 2, sessionId: 'session-1' });

    await expect(run).resolves.toBeUndefined();
    expect(events).toEqual([
      { type: 'session.started', seq: 1, sessionId: 'session-1' },
      { type: 'session.cancelled', seq: 2, sessionId: 'session-1' }
    ]);
  });

  it('allows a normal close while a terminal completion message is still queued', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.started', seq: 1, sessionId: 'session-1' });
    socket.receive({ type: 'session.completed', seq: 2, sessionId: 'session-1' });
    socket.closeFromServer(1000, 'voice session completed');

    await expect(run).resolves.toBeUndefined();
    expect(events).toEqual([
      { type: 'session.started', seq: 1, sessionId: 'session-1' },
      { type: 'session.completed', seq: 2, sessionId: 'session-1' }
    ]);
  });

  it('allows a normal close while a terminal failure message is still queued', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({ type: 'session.failed', seq: 1, code: 'invalid_request', message: 'No provider.' });
    socket.closeFromServer(1008, 'provider secret detail must not surface');

    await expect(run).resolves.toBeUndefined();
    expect(events).toEqual([
      { type: 'session.failed', seq: 1, code: 'invalid_request', message: 'No provider.' }
    ]);
  });


  it('includes safe close metadata when the socket closes before a terminal event', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async () => {});

    socket.open();
    socket.closeFromServer(1008, 'voice session rejected');

    await expect(run).rejects.toThrow(
      'Voice socket closed before the session completed (code 1008).'
    );
  });
});

class FakeWebSocket {
  onopen: (() => void) | null = null;
  onmessage: ((event: { readonly data: string }) => void) | null = null;
  onerror: ((event: unknown) => void) | null = null;
  onclose: ((event?: { readonly code?: number; readonly reason?: string }) => void) | null = null;
  readonly sent: Array<Record<string, unknown>> = [];
  closedByClient = false;
  failNextSend = false;
  url = '';
  headers: Record<string, string> = {};

  send(payload: string): void {
    if (this.failNextSend) {
      this.failNextSend = false;
      throw new Error('socket send failed');
    }
    this.sent.push(JSON.parse(payload) as Record<string, unknown>);
  }

  close(): void {
    this.closedByClient = true;
  }

  closeFromServer(code?: number, reason?: string): void {
    this.onclose?.({ code, reason });
  }

  open(): void {
    this.onopen?.();
  }

  receive(message: Record<string, unknown>): void {
    this.onmessage?.({ data: JSON.stringify(message) });
  }
}

async function waitForSentMessageCount(socket: FakeWebSocket, count: number): Promise<void> {
  for (let attempts = 0; attempts < 20 && socket.sent.length < count; attempts++) {
    await Promise.resolve();
  }
}

async function waitForEventType(events: readonly unknown[], type: string): Promise<void> {
  for (let attempts = 0; attempts < 20 && !events.some((event) => isEventType(event, type)); attempts++) {
    await Promise.resolve();
  }
}

function isEventType(event: unknown, type: string): boolean {
  return typeof event === 'object' && event !== null && 'type' in event && event.type === type;
}
