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
      let completed = false;

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
        void (async () => {
          const message = parseServerMessage(event.data);
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
            reject(new Error(message.message));
          }
        })().catch(reject);
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
  switch (message.type) {
    case 'session.started':
      return { type: 'session.started', sessionId: stringField(message, 'sessionId') };
    case 'session.failed':
      return {
        type: 'session.failed',
        code: stringField(message, 'code'),
        message: stringField(message, 'message')
      };
    case 'transcript.final':
      return { type: 'transcript.final', text: stringField(message, 'text') };
    case 'agent.progress':
      return {
        type: 'agent.progress',
        status: stringField(message, 'status'),
        message: stringField(message, 'message')
      };
    case 'tool.call.started':
    case 'tool.call.completed':
    case 'tool.call.failed':
      return {
        type: message.type,
        toolCallId: stringField(message, 'toolCallId'),
        toolLabel: stringField(message, 'toolLabel'),
        status: optionalStringField(message, 'status'),
        code: optionalStringField(message, 'code'),
        message: optionalStringField(message, 'message')
      };
    case 'assistant.response.completed': {
      const response = objectField(message, 'response');
      return {
        type: 'assistant.response.completed',
        response: {
          kind: stringField(response, 'kind'),
          spokenResponse: stringField(response, 'spokenResponse'),
          displayResponse: stringField(response, 'displayResponse')
        }
      };
    }
    case 'tts.audio.started': {
      const format = objectField(message, 'format');
      return { type: 'tts.audio.started', mimeType: stringField(format, 'mimeType') };
    }
    case 'tts.audio.chunk':
      return {
        type: 'tts.audio.chunk',
        chunkId: stringField(message, 'chunkId'),
        audioBase64: stringField(message, 'audioBase64')
      };
    case 'tts.audio.completed':
      return { type: 'tts.audio.completed' };
    case 'session.completed':
      return { type: 'session.completed' };
    default:
      throw new Error(`Unsupported voice event: ${String(message.type)}`);
  }
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
