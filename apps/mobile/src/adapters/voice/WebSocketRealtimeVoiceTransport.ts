import type {
  RealtimeVoiceTransport,
  RealtimeVoiceTransportInput,
  VoiceRealtimeEvent
} from '../../application/voice/RealtimeVoiceSession';

type VoiceWebSocket = {
  onopen: (() => void) | null;
  onmessage: ((event: { readonly data: string }) => void) | null;
  onerror: ((event: unknown) => void) | null;
  onclose: (() => void) | null;
  send(payload: string): void;
  close(): void;
};

type VoiceWebSocketFactory = (url: string, headers: Record<string, string>) => VoiceWebSocket;

export type WebSocketRealtimeVoiceTransportOptions = {
  readonly apiBaseUrl: string;
  readonly tokenProvider: () => string;
  readonly webSocketFactory?: VoiceWebSocketFactory;
};

export class WebSocketRealtimeVoiceTransport implements RealtimeVoiceTransport {
  private readonly apiBaseUrl: string;
  private readonly tokenProvider: () => string;
  private readonly webSocketFactory: VoiceWebSocketFactory;

  constructor(options: WebSocketRealtimeVoiceTransportOptions) {
    this.apiBaseUrl = options.apiBaseUrl.replace(/\/+$/, '');
    this.tokenProvider = options.tokenProvider;
    this.webSocketFactory = options.webSocketFactory ?? createReactNativeWebSocket;
  }

  async run(input: RealtimeVoiceTransportInput, onEvent: (event: VoiceRealtimeEvent) => Promise<void>): Promise<void> {
    const socket = this.webSocketFactory(realtimeVoiceUrl(this.apiBaseUrl), {
      Authorization: `Bearer ${this.tokenProvider()}`
    });

    await new Promise<void>((resolve, reject) => {
      let sessionId = '';
      let seq = 1;
      let lastServerSeq = 0;
      let completed = false;
      let messageChain = Promise.resolve();

      socket.onerror = (event) => {
        reject(new Error(`Voice socket failed: ${String(event)}`));
      };
      socket.onclose = () => {
        if (!completed) {
          reject(new Error('Voice socket closed before the session completed.'));
        }
      };
      socket.onopen = () => {
        socket.send(JSON.stringify({
          type: 'session.start',
          seq: seq++,
          tenantId: input.tenantId,
          inventoryId: input.inventoryId,
          source: input.source,
          requestedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech'],
          inputAudio: input.inputAudio,
          outputAudio: { mimeTypes: input.outputAudioMimeTypes }
        }));
      };
      socket.onmessage = (event) => {
        messageChain = messageChain.then(async () => {
          const message = parseServerMessage(event.data);
          validateServerMessage(message, sessionId, lastServerSeq);
          lastServerSeq = message.seq;
          await onEvent(message);
          if (message.type === 'session.started') {
            sessionId = message.sessionId;
            input.audioChunksBase64.forEach((audioBase64, index) => {
              socket.send(JSON.stringify({
                type: 'audio.chunk',
                seq: seq++,
                sessionId,
                chunkId: `mobile-${index + 1}`,
                audioBase64,
                isFinalChunk: index === input.audioChunksBase64.length - 1
              }));
            });
            socket.send(JSON.stringify({ type: 'audio.end', seq: seq++, sessionId }));
          }
          if (message.type === 'session.completed') {
            completed = true;
            socket.close();
            resolve();
          }
          if (message.type === 'session.failed') {
            completed = true;
            socket.close();
            resolve();
          }
        }).catch(reject);
      };
    });
  }
}

function realtimeVoiceUrl(apiBaseUrl: string): string {
  const url = new URL('/v1/realtime/voice', apiBaseUrl);
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
  return url.toString();
}

function createReactNativeWebSocket(url: string, headers: Record<string, string>): VoiceWebSocket {
  const WebSocketConstructor = globalThis.WebSocket as unknown as new (
    url: string,
    protocols?: string | string[],
    options?: { readonly headers?: Record<string, string> }
  ) => VoiceWebSocket;
  return new WebSocketConstructor(url, [], { headers });
}

function parseServerMessage(raw: string): VoiceRealtimeEvent {
  const message = JSON.parse(raw) as Record<string, unknown>;
  const metadata = eventMetadata(message);
  switch (message.type) {
    case 'session.started':
      return { ...metadata, type: 'session.started', sessionId: stringField(message, 'sessionId') };
    case 'session.failed':
      return {
        ...metadata,
        type: 'session.failed',
        sessionId: optionalStringField(message, 'sessionId'),
        code: stringField(message, 'code'),
        message: stringField(message, 'message')
      };
    case 'transcript.final':
      return { ...metadata, type: 'transcript.final', sessionId: stringField(message, 'sessionId'), text: stringField(message, 'text') };
    case 'agent.progress':
      return {
        ...metadata,
        type: 'agent.progress',
        sessionId: stringField(message, 'sessionId'),
        status: stringField(message, 'status'),
        message: stringField(message, 'message')
      };
		case 'tool.call.started':
		case 'tool.call.completed':
		case 'tool.call.failed':
      return {
        ...metadata,
        type: message.type,
        sessionId: stringField(message, 'sessionId'),
        toolCallId: stringField(message, 'toolCallId'),
        toolLabel: stringField(message, 'toolLabel'),
        status: optionalStringField(message, 'status'),
        code: optionalStringField(message, 'code'),
				message: optionalStringField(message, 'message')
			};
		case 'assistant.response.started':
			return { ...metadata, type: 'assistant.response.started', sessionId: stringField(message, 'sessionId'), responseId: stringField(message, 'responseId') };
		case 'assistant.response.completed': {
      const response = objectField(message, 'response');
      return {
        ...metadata,
        type: 'assistant.response.completed',
        sessionId: stringField(message, 'sessionId'),
        response: {
          kind: stringField(response, 'kind'),
          spokenResponse: stringField(response, 'spokenResponse'),
          displayResponse: stringField(response, 'displayResponse')
        }
      };
    }
    case 'tts.audio.started': {
      const format = objectField(message, 'format');
      return { ...metadata, type: 'tts.audio.started', sessionId: stringField(message, 'sessionId'), mimeType: stringField(format, 'mimeType') };
    }
    case 'tts.audio.chunk':
      return {
        ...metadata,
        type: 'tts.audio.chunk',
        sessionId: stringField(message, 'sessionId'),
        chunkId: stringField(message, 'chunkId'),
        audioBase64: stringField(message, 'audioBase64')
      };
    case 'tts.audio.completed':
      return { ...metadata, type: 'tts.audio.completed', sessionId: stringField(message, 'sessionId') };
    case 'session.completed':
      return { ...metadata, type: 'session.completed', sessionId: stringField(message, 'sessionId') };
    default:
      throw new Error(`Unsupported voice event: ${String(message.type)}`);
  }
}

function eventMetadata(message: Record<string, unknown>): { readonly seq: number; readonly sessionId?: string } {
  return {
    seq: numberField(message, 'seq'),
    sessionId: optionalStringField(message, 'sessionId')
  };
}

function validateServerMessage(message: VoiceRealtimeEvent, currentSessionId: string, lastServerSeq: number): void {
  if (message.seq <= lastServerSeq) {
    throw new Error('Voice server event sequence must be monotonic.');
  }
  if (message.type === 'session.started') {
    if (currentSessionId && message.sessionId !== currentSessionId) {
      throw new Error('Voice server event changed session.');
    }
    return;
  }
  if (!currentSessionId) {
    if (message.type !== 'session.failed') {
      throw new Error('Voice server event arrived before session start.');
    }
    return;
  }
  if (message.sessionId !== currentSessionId) {
    throw new Error('Voice server event session did not match the active session.');
  }
}

function numberField(message: Record<string, unknown>, field: string): number {
  const value = message[field];
  if (typeof value !== 'number' || !Number.isInteger(value) || value <= 0) {
    throw new Error(`Voice event field ${field} must be a positive integer.`);
  }
  return value;
}

function stringField(message: Record<string, unknown>, field: string): string {
  const value = message[field];
  if (typeof value !== 'string' || value.trim().length === 0) {
    throw new Error(`Voice event field ${field} must be a non-empty string.`);
  }
  return value;
}

function optionalStringField(message: Record<string, unknown>, field: string): string | undefined {
  const value = message[field];
  return typeof value === 'string' && value.trim().length > 0 ? value : undefined;
}

function objectField(message: Record<string, unknown>, field: string): Record<string, unknown> {
  const value = message[field];
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`Voice event field ${field} must be an object.`);
  }
  return value as Record<string, unknown>;
}
