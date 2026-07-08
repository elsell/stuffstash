import { describe, expect, it } from 'vitest';
import { WebSocketRealtimeVoiceTransport } from './WebSocketRealtimeVoiceTransport';

describe('WebSocketRealtimeVoiceTransport', () => {
  it('awaits an async OIDC token before opening the realtime socket', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: async () => 'refreshed-id-token',
      webSocketFactory: (url, headers) => {
        socket.url = url;
        socket.headers = headers;
        return socket;
      }
    });
    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: []
    }, async () => {});

    await waitForSocketReady(socket);
    socket.open();
    socket.receive({
      type: 'session.started',
      seq: 1,
      sessionId: 'session-1',
      acceptedInputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      acceptedOutputAudio: { mimeTypes: ['audio/mpeg'] },
      acceptedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech']
    });
    socket.receive({ type: 'session.completed', seq: 2, sessionId: 'session-1' });

    await run;
    expect(socket.headers).toEqual({ Authorization: 'Bearer refreshed-id-token' });
  });

  it('does not open a realtime socket when mobile auth is unavailable', async () => {
    let openedSocket = false;
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: async () => {
        throw new Error('Sign in to Stuff Stash.');
      },
      webSocketFactory: () => {
        openedSocket = true;
        return new FakeWebSocket();
      }
    });

    await expect(transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: []
    }, async () => {})).rejects.toThrow('Sign in to Stuff Stash.');
    expect(openedSocket).toBe(false);
  });

  it('does not open a realtime socket when mobile auth returns no token', async () => {
    let openedSocket = false;
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => '',
      webSocketFactory: () => {
        openedSocket = true;
        return new FakeWebSocket();
      }
    });

    await expect(transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: []
    }, async () => {})).rejects.toThrow('Sign in before starting a voice session.');
    expect(openedSocket).toBe(false);
  });

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
      clientCorrelationId: 'mobile-voice-1',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['YXVkaW8=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive({
      type: 'session.started',
      seq: 1,
      sessionId: 'session-1',
      acceptedInputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      acceptedOutputAudio: { mimeTypes: ['audio/mpeg'] },
      acceptedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech']
    });
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
    socket.receive({ type: 'tts.audio.chunk', seq: 7, sessionId: 'session-1', chunkId: 'tts-1', audioBase64: 'c3BlZWNo', isFinalChunk: true });
    socket.receive({ type: 'session.completed', seq: 8, sessionId: 'session-1' });
    await waitForSentMessageCount(socket, 3);
    await waitForEventType(events, 'action.plan.proposed');
    await transport.approveActionPlan('plan-1', [{
      commandId: 'cmd-water-bottle',
      photoIndex: 0,
      fileName: 'water-bottle.jpg',
      contentType: 'image/jpeg',
      sizeBytes: 123
    }]);
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
      }],
      attachmentUploadIntents: [{
        commandId: 'cmd-water-bottle',
        photoIndex: 0,
        assetId: 'asset-water-bottle',
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 123,
        directUpload: {
          uploadId: 'upload-one',
          attachmentId: 'attachment-one',
          method: 'POST',
          url: 'https://uploads.example.test/upload-one',
          headers: {},
          formFields: { key: 'object-one' },
          expiresAt: '2026-07-01T12:00:00Z'
        }
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
      source: 'mobile_voice',
      clientCorrelationId: 'mobile-voice-1',
      requestedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech']
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
      planId: 'plan-1',
      photoAttachments: [{
        commandId: 'cmd-water-bottle',
        photoIndex: 0,
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 123
      }]
    });
    expect(events).toEqual([
      {
        type: 'session.started',
        seq: 1,
        sessionId: 'session-1',
        acceptedInputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
        acceptedOutputAudio: { mimeTypes: ['audio/mpeg'] },
        acceptedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech']
      },
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
      { type: 'tts.audio.chunk', seq: 7, sessionId: 'session-1', chunkId: 'tts-1', audioBase64: 'c3BlZWNo', isFinalChunk: true },
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
        }],
        attachmentUploadIntents: [{
          commandId: 'cmd-water-bottle',
          photoIndex: 0,
          assetId: 'asset-water-bottle',
          fileName: 'water-bottle.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 123,
          directUpload: {
            uploadId: 'upload-one',
            attachmentId: 'attachment-one',
            method: 'POST',
            url: 'https://uploads.example.test/upload-one',
            headers: {},
            formFields: { key: 'object-one' },
            expiresAt: '2026-07-01T12:00:00Z'
          }
        }]
      }
    ]);
  });

  it('rejects session starts that do not accept the required voice capabilities', async () => {
    const malformedCapabilities = [
      undefined,
      ['speech_to_text', 'text_to_speech'],
      ['speech_to_text', 'language_inference', 'raw_provider'],
      ['speech_to_text', 'language_inference', 'language_inference'],
      [' speech_to_text ', 'language_inference', 'text_to_speech']
    ];

    for (const acceptedCapabilities of malformedCapabilities) {
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
      const started = sessionStarted();
      if (acceptedCapabilities === undefined) {
        delete (started as Record<string, unknown>).acceptedCapabilities;
      } else {
        (started as Record<string, unknown>).acceptedCapabilities = acceptedCapabilities;
      }
      socket.receive(started);
      socket.closeFromServer(1000);

      await expect(run).rejects.toThrow('Voice event field acceptedCapabilities must match requested voice capabilities.');
      expect(socket.sent.map((message) => message.type)).toEqual(['session.start']);
    }
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
    const followUpEvents: unknown[] = [];

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
    socket.receive(sessionStarted());
    socket.receive({ type: 'transcript.final', seq: 2, sessionId: 'session-other', text: 'Wrong session' });

    await expect(run).rejects.toThrow('session');
    expect(events).toEqual([sessionStarted()]);
  });

  it.each([
    ['missing marker', {}],
    ['wrong type marker', { isFinalChunk: 'true' }]
  ])('rejects TTS audio chunks with %s', async (_name, markerFields) => {
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
    socket.receive(sessionStarted());
    socket.receive({ type: 'tts.audio.chunk', seq: 2, sessionId: 'session-1', chunkId: 'tts-1', audioBase64: 'c3BlZWNo', ...markerFields });

    await expect(run).rejects.toThrow('isFinalChunk');
    expect(events).toEqual([sessionStarted()]);
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
    const followUpEvents: unknown[] = [];

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
    socket.receive(sessionStarted());
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

  it('keeps the socket open and sends follow-up audio after a clarification completion', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];
    const followUpEvents: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['Zmlyc3Q=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'assistant.response.completed',
      seq: 2,
      sessionId: 'session-1',
      response: { kind: 'clarification', spokenResponse: 'Which item?', displayResponse: 'Which item?' }
    });
    socket.receive({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    await waitForEventType(events, 'session.completed');

    expect(transport.canSendFollowUpAudio()).toBe(true);
    expect(socket.closedByClient).toBe(false);
    const followUp = transport.sendFollowUpAudio(['c2Vjb25k'], async (event) => {
      followUpEvents.push(event);
    });
    socket.receive({
      type: 'assistant.response.completed',
      seq: 4,
      sessionId: 'session-1',
      response: { kind: 'answer', spokenResponse: 'Found it.', displayResponse: 'Found it.' }
    });
    socket.receive({ type: 'session.completed', seq: 5, sessionId: 'session-1' });

    await followUp;
    await run;
    expect(socket.sent.map((message) => message.type)).toEqual([
      'session.start',
      'audio.chunk',
      'audio.end',
      'audio.chunk',
      'audio.end'
    ]);
    expect(socket.sent[3]).toMatchObject({ seq: 4, sessionId: 'session-1', audioBase64: 'c2Vjb25k' });
    expect(followUpEvents.map((event) => isEventType(event, 'assistant.response.completed') || isEventType(event, 'session.completed'))).toEqual([true, true]);
    expect(transport.canSendFollowUpAudio()).toBe(false);
  });

  it('cancels the existing conversation socket when follow-up audio is aborted before send', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];
    const controller = new AbortController();

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['Zmlyc3Q=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'assistant.response.completed',
      seq: 2,
      sessionId: 'session-1',
      response: { kind: 'clarification', spokenResponse: 'Which item?', displayResponse: 'Which item?' }
    });
    socket.receive({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    await waitForEventType(events, 'session.completed');
    await run;

    controller.abort();
    await expect(transport.sendFollowUpAudio(['c2Vjb25k'], undefined, { signal: controller.signal })).rejects.toMatchObject({ code: 'voice_cancelled' });

    expect(socket.sent.map((message) => message.type)).toEqual([
      'session.start',
      'audio.chunk',
      'audio.end',
      'session.cancel'
    ]);
    expect(socket.sent.at(-1)).toMatchObject({
      sessionId: 'session-1',
      reason: 'user_cancelled'
    });
    expect(socket.closedByClient).toBe(true);
    expect(transport.canSendFollowUpAudio()).toBe(false);
  });

  it('settles each repeated clarification follow-up while keeping the socket open', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];
    const firstFollowUpEvents: unknown[] = [];
    const secondFollowUpEvents: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: ['Zmlyc3Q=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'assistant.response.completed',
      seq: 2,
      sessionId: 'session-1',
      response: { kind: 'clarification', spokenResponse: 'Which item?', displayResponse: 'Which item?' }
    });
    socket.receive({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    await waitForEventType(events, 'session.completed');

    const firstFollowUp = transport.sendFollowUpAudio(['c2Vjb25k'], async (event) => {
      firstFollowUpEvents.push(event);
    });
    socket.receive({
      type: 'assistant.response.completed',
      seq: 4,
      sessionId: 'session-1',
      response: { kind: 'clarification', spokenResponse: 'Where should it go?', displayResponse: 'Where should it go?' }
    });
    socket.receive({ type: 'session.completed', seq: 5, sessionId: 'session-1' });

    await expect(promiseSettled(firstFollowUp)).resolves.toBe(true);
    expect(transport.canSendFollowUpAudio()).toBe(true);

    const secondFollowUp = transport.sendFollowUpAudio(['dGhpcmQ='], async (event) => {
      secondFollowUpEvents.push(event);
    });
    socket.receive({
      type: 'assistant.response.completed',
      seq: 6,
      sessionId: 'session-1',
      response: { kind: 'answer', spokenResponse: 'Moved it.', displayResponse: 'Moved it.' }
    });
    socket.receive({ type: 'session.completed', seq: 7, sessionId: 'session-1' });

    await secondFollowUp;
    await run;
    expect(socket.closedByClient).toBe(true);
    expect(socket.sent.map((message) => message.type)).toEqual([
      'session.start',
      'audio.chunk',
      'audio.end',
      'audio.chunk',
      'audio.end',
      'audio.chunk',
      'audio.end'
    ]);
    expect(firstFollowUpEvents.map((event) => isEventType(event, 'assistant.response.completed') || isEventType(event, 'session.completed'))).toEqual([true, true]);
    expect(secondFollowUpEvents.map((event) => isEventType(event, 'assistant.response.completed') || isEventType(event, 'session.completed'))).toEqual([true, true]);
    expect(transport.canSendFollowUpAudio()).toBe(false);
  });

  it('clears follow-up availability when an idle clarification socket closes', async () => {
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
      audioChunksBase64: ['Zmlyc3Q=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'assistant.response.completed',
      seq: 2,
      sessionId: 'session-1',
      response: { kind: 'clarification', spokenResponse: 'Which item?', displayResponse: 'Which item?' }
    });
    socket.receive({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    await run;

    expect(transport.canSendFollowUpAudio()).toBe(true);
    socket.closeFromServer(1000, 'clarification follow-up timed out');
    await waitForNoFollowUpAudio(transport);
    expect(transport.canSendFollowUpAudio()).toBe(false);
    await expect(transport.sendFollowUpAudio(['c2Vjb25k'])).rejects.toThrow('not active');
  });

  it('rejects a pending follow-up when the clarification socket closes mid-turn', async () => {
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
      audioChunksBase64: ['Zmlyc3Q=']
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'assistant.response.completed',
      seq: 2,
      sessionId: 'session-1',
      response: { kind: 'clarification', spokenResponse: 'Which item?', displayResponse: 'Which item?' }
    });
    socket.receive({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    await run;

    const followUp = transport.sendFollowUpAudio(['c2Vjb25k']);
    socket.closeFromServer(1006, 'network dropped');

    await expect(followUp).rejects.toThrow('Voice socket closed before the session completed (code 1006).');
    expect(transport.canSendFollowUpAudio()).toBe(false);
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
    socket.receive(sessionStarted());

    await expect(run).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(socket.sent.at(-1)).toMatchObject({
      type: 'session.cancel',
      sessionId: 'session-1',
      reason: 'user_cancelled'
    });
    expect(socket.closedByClient).toBe(true);

    socket.receive({ type: 'transcript.final', seq: 2, sessionId: 'session-1', text: 'late transcript' });
    await Promise.resolve();
    expect(events).toEqual([sessionStarted()]);
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
    socket.receive(sessionStarted());

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
    socket.receive(sessionStarted());
    socket.receive({ type: 'session.cancelled', seq: 2, sessionId: 'session-1' });

    await expect(run).resolves.toBeUndefined();
    expect(events).toEqual([
      sessionStarted(),
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
    socket.receive(sessionStarted());
    socket.receive({ type: 'session.completed', seq: 2, sessionId: 'session-1' });
    socket.closeFromServer(1000, 'voice session completed');

    await expect(run).resolves.toBeUndefined();
    expect(events).toEqual([
      sessionStarted(),
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

  it('rejects unsupported direct upload targets from executed action plan events', async () => {
    for (const url of [
      'http://uploads.example.test/upload-one',
      'http://192.168.1.12:3900/upload-one',
      'stuffstash-local://direct-uploads/upload-one'
    ]) {
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
        audioChunksBase64: []
      }, async (event) => {
        events.push(event);
      });

      socket.open();
      socket.receive(sessionStarted());
      socket.receive({
        type: 'action.plan.executed',
        seq: 2,
        sessionId: 'session-1',
        planId: 'plan-1',
        status: 'executed',
        message: 'The approved change was applied.',
        commandResults: [{
          commandId: 'cmd-water-bottle',
          assetId: 'asset-water-bottle',
          operation: 'create',
          assetKind: 'item'
        }],
        attachmentUploadIntents: [{
          commandId: 'cmd-water-bottle',
          photoIndex: 0,
          assetId: 'asset-water-bottle',
          fileName: 'water-bottle.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 123,
          directUpload: {
            uploadId: 'upload-one',
            attachmentId: 'attachment-one',
            method: 'POST',
            url,
            headers: {},
            formFields: {},
            expiresAt: '2026-07-01T12:00:00Z'
          }
        }]
      });

      await expect(run).rejects.toThrow('Voice event field url has unsupported direct upload URL.');
      expect(events).toEqual([
        sessionStarted()
      ]);
    }
  });

  it('rejects unsupported direct upload methods from executed action plan events', async () => {
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
      audioChunksBase64: []
    }, async () => {});

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'action.plan.executed',
      seq: 2,
      sessionId: 'session-1',
      planId: 'plan-1',
      status: 'executed',
      message: 'The approved change was applied.',
      attachmentUploadIntents: [{
        commandId: 'cmd-water-bottle',
        photoIndex: 0,
        assetId: 'asset-water-bottle',
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 123,
        directUpload: {
          uploadId: 'upload-one',
          attachmentId: 'attachment-one',
          method: 'DELETE',
          url: 'https://uploads.example.test/upload-one',
          headers: {},
          formFields: {},
          expiresAt: '2026-07-01T12:00:00Z'
        }
      }]
    });

    await expect(run).rejects.toThrow('Voice event field method has unsupported direct upload method.');
  });

  it('accepts local direct upload sentinels only when local targets are enabled', async () => {
    const socket = new FakeWebSocket();
    const transport = new WebSocketRealtimeVoiceTransport({
      apiBaseUrl: 'http://127.0.0.1:8080/',
      tokenProvider: () => 'dev:user-1',
      directUploadPolicy: { allowLocalDevelopmentTargets: true },
      webSocketFactory: () => socket
    });
    const events: unknown[] = [];

    const run = transport.run({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      source: 'mobile_voice',
      inputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
      outputAudioMimeTypes: ['audio/mpeg'],
      audioChunksBase64: []
    }, async (event) => {
      events.push(event);
    });

    socket.open();
    socket.receive(sessionStarted());
    socket.receive({
      type: 'action.plan.executed',
      seq: 2,
      sessionId: 'session-1',
      planId: 'plan-1',
      status: 'executed',
      message: 'The approved change was applied.',
      attachmentUploadIntents: [{
        commandId: 'cmd-water-bottle',
        photoIndex: 0,
        assetId: 'asset-water-bottle',
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 123,
        directUpload: {
          uploadId: 'upload-one',
          attachmentId: 'attachment-one',
          method: 'PUT',
          url: 'stuffstash-local://direct-uploads/upload-one',
          headers: {},
          formFields: {},
          expiresAt: '2026-07-01T12:00:00Z'
        }
      }]
    });

    await expect(run).resolves.toBeUndefined();
    expect(events).toEqual([
      sessionStarted(),
      {
        type: 'action.plan.executed',
        seq: 2,
        sessionId: 'session-1',
        planId: 'plan-1',
        status: 'executed',
        message: 'The approved change was applied.',
        attachmentUploadIntents: [{
          commandId: 'cmd-water-bottle',
          photoIndex: 0,
          assetId: 'asset-water-bottle',
          fileName: 'water-bottle.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 123,
          directUpload: {
            uploadId: 'upload-one',
            attachmentId: 'attachment-one',
            method: 'PUT',
            url: 'stuffstash-local://direct-uploads/upload-one',
            headers: {},
            formFields: {},
            expiresAt: '2026-07-01T12:00:00Z'
          }
        }]
      }
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

function sessionStarted() {
  return {
    type: 'session.started',
    seq: 1,
    sessionId: 'session-1',
    acceptedInputAudio: { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1 },
    acceptedOutputAudio: { mimeTypes: ['audio/mpeg'] },
    acceptedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech']
  };
}

async function waitForSentMessageCount(socket: FakeWebSocket, count: number): Promise<void> {
  for (let attempts = 0; attempts < 20 && socket.sent.length < count; attempts++) {
    await Promise.resolve();
  }
}

async function waitForSocketReady(socket: FakeWebSocket): Promise<void> {
  for (let attempts = 0; attempts < 20 && !socket.onmessage; attempts++) {
    await Promise.resolve();
  }
}

async function waitForEventType(events: readonly unknown[], type: string): Promise<void> {
  for (let attempts = 0; attempts < 20 && !events.some((event) => isEventType(event, type)); attempts++) {
    await Promise.resolve();
  }
}

async function waitForNoFollowUpAudio(transport: WebSocketRealtimeVoiceTransport): Promise<void> {
  for (let attempts = 0; attempts < 20 && transport.canSendFollowUpAudio(); attempts++) {
    await Promise.resolve();
  }
}

async function promiseSettled(promise: Promise<unknown>): Promise<boolean> {
  let settled = false;
  void promise.then(() => {
    settled = true;
  }, () => {
    settled = true;
  });
  for (let attempts = 0; attempts < 20 && !settled; attempts++) {
    await Promise.resolve();
  }
  return settled;
}

function isEventType(event: unknown, type: string): boolean {
  return typeof event === 'object' && event !== null && 'type' in event && event.type === type;
}
