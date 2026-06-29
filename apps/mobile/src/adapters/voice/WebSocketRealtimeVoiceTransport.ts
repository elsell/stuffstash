import type {
  RealtimeVoiceTransport,
  RealtimeVoiceTransportInput,
  RealtimeVoiceTransportRunOptions,
  VoiceRealtimeEvent
} from '../../application/voice/RealtimeVoiceSession';
import { VoiceRealtimeCancelledError } from '../../application/voice/RealtimeVoiceSession';

type VoiceWebSocket = {
  onopen: (() => void) | null;
  onmessage: ((event: { readonly data: string }) => void) | null;
  onerror: ((event: unknown) => void) | null;
  onclose: ((event?: { readonly code?: number; readonly reason?: string }) => void) | null;
  send(payload: string): void;
  close(): void;
};

type VoiceWebSocketFactory = (url: string, headers: Record<string, string>) => VoiceWebSocket;

type ActiveRealtimeReviewSession = {
  readonly sendDecision: (type: 'action.plan.approve' | 'action.plan.cancel', planId: string) => void;
};

export type WebSocketRealtimeVoiceTransportOptions = {
  readonly apiBaseUrl: string;
  readonly tokenProvider: () => string;
  readonly diagnosticsEnabled?: boolean;
  readonly webSocketFactory?: VoiceWebSocketFactory;
};

export class WebSocketRealtimeVoiceTransport implements RealtimeVoiceTransport {
  private readonly apiBaseUrl: string;
  private readonly tokenProvider: () => string;
  private readonly diagnosticsEnabled: boolean;
  private readonly webSocketFactory: VoiceWebSocketFactory;
  private activeReviewSession: ActiveRealtimeReviewSession | null = null;

  constructor(options: WebSocketRealtimeVoiceTransportOptions) {
    this.apiBaseUrl = options.apiBaseUrl.replace(/\/+$/, '');
    this.tokenProvider = options.tokenProvider;
    this.diagnosticsEnabled = options.diagnosticsEnabled ?? false;
    this.webSocketFactory = options.webSocketFactory ?? createReactNativeWebSocket;
  }

  async run(
    input: RealtimeVoiceTransportInput,
    onEvent: (event: VoiceRealtimeEvent) => Promise<void>,
    options: RealtimeVoiceTransportRunOptions = {}
  ): Promise<void> {
    const socket = this.webSocketFactory(realtimeVoiceUrl(this.apiBaseUrl), {
      Authorization: `Bearer ${this.tokenProvider()}`
    });
    const thisTransport = this;

    await new Promise<void>((resolve, reject) => {
      let sessionId = '';
      let seq = 1;
      let lastServerSeq = 0;
      let completed = false;
      let hasPendingActionPlan = false;
      let settled = false;
      let decisionSent = false;
      let messageChain = Promise.resolve();
      const sendDecision = (type: 'action.plan.approve' | 'action.plan.cancel', planId: string) => {
        if (!sessionId || settled) {
          throw new Error('Voice review session is not active.');
        }
        if (decisionSent) {
          throw new Error('Voice review decision has already been sent.');
        }
        decisionSent = true;
        thisTransport.activeReviewSession = null;
        socket.send(JSON.stringify({
          type,
          seq: seq++,
          sessionId,
          planId
        }));
      };
      const abortHandler = () => {
        if (settled) {
          return;
        }
        completed = true;
        try {
          if (sessionId) {
            socket.send(JSON.stringify({
              type: 'session.cancel',
              seq: seq++,
              sessionId,
              reason: 'user_cancelled'
            }));
          }
        } catch {
          // Cancellation is best effort once the socket is already closing.
        } finally {
          socket.close();
          settleReject(new VoiceRealtimeCancelledError());
        }
      };

      function settleResolve(): void {
        if (!settled) {
          settled = true;
          thisTransport.activeReviewSession = null;
          options.signal?.removeEventListener('abort', abortHandler);
          resolve();
        }
      }

      function settleReject(error: Error): void {
        if (!settled) {
          settled = true;
          thisTransport.activeReviewSession = null;
          options.signal?.removeEventListener('abort', abortHandler);
          reject(error);
        }
      }

      if (options.signal?.aborted) {
        abortHandler();
        return;
      }
      options.signal?.addEventListener('abort', abortHandler, { once: true });
      socket.onerror = (event) => {
        settleReject(new Error(`Voice socket failed: ${String(event)}`));
      };
      socket.onclose = (event) => {
        void messageChain.then(() => {
          if (!completed) {
            settleReject(new Error(prematureCloseMessage(event)));
          }
        }).catch(settleReject);
      };
      socket.onopen = () => {
        socket.send(JSON.stringify({
          type: 'session.start',
          seq: seq++,
          tenantId: input.tenantId,
          inventoryId: input.inventoryId,
          source: input.source,
          requestedCapabilities: ['speech_to_text', 'language_inference', 'text_to_speech'],
          developerDiagnostics: this.diagnosticsEnabled,
          inputAudio: input.inputAudio,
          outputAudio: { mimeTypes: input.outputAudioMimeTypes }
        }));
      };
      socket.onmessage = (event) => {
        messageChain = messageChain.then(async () => {
          if (settled) {
            return;
          }
          const message = parseServerMessage(event.data);
          validateServerMessage(message, sessionId, lastServerSeq);
          lastServerSeq = message.seq;
          if (message.type === 'session.started') {
            sessionId = message.sessionId;
          }
          await onEvent(message);
          if (message.type === 'session.started') {
            if (settled || options.signal?.aborted) {
              return;
            }
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
          if (message.type === 'action.plan.proposed') {
            hasPendingActionPlan = true;
            this.activeReviewSession = { sendDecision };
          }
          if (message.type === 'session.completed' && !hasPendingActionPlan) {
            completed = true;
            socket.close();
            settleResolve();
          }
          if (message.type === 'action.plan.cancelled' || message.type === 'action.plan.executed' || message.type === 'action.plan.failed') {
            completed = true;
            this.activeReviewSession = null;
            socket.close();
            settleResolve();
          }
          if (message.type === 'session.cancelled') {
            completed = true;
            this.activeReviewSession = null;
            socket.close();
            settleResolve();
          }
          if (message.type === 'session.failed') {
            completed = true;
            this.activeReviewSession = null;
            socket.close();
            settleResolve();
          }
        }).catch(settleReject);
      };
    });
  }

  async approveActionPlan(planId: string): Promise<void> {
    if (!this.activeReviewSession) {
      throw new Error('Voice review session is not active.');
    }
    this.activeReviewSession.sendDecision('action.plan.approve', planId);
  }

  async cancelActionPlan(planId: string): Promise<void> {
    if (!this.activeReviewSession) {
      throw new Error('Voice review session is not active.');
    }
    this.activeReviewSession.sendDecision('action.plan.cancel', planId);
  }
}

function prematureCloseMessage(event?: { readonly code?: number; readonly reason?: string }): string {
  return typeof event?.code === 'number'
    ? `Voice socket closed before the session completed (code ${event.code}).`
    : 'Voice socket closed before the session completed.';
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
    case 'transcript.delta':
    case 'transcript.final':
      return { ...metadata, type: message.type, sessionId: stringField(message, 'sessionId'), text: stringField(message, 'text') };
    case 'agent.progress':
      return {
        ...metadata,
        type: 'agent.progress',
        sessionId: stringField(message, 'sessionId'),
        status: stringField(message, 'status'),
        message: stringField(message, 'message')
      };
    case 'agent.diagnostic':
      return {
        ...metadata,
        type: 'agent.diagnostic',
        sessionId: stringField(message, 'sessionId'),
        message: stringField(message, 'message'),
        detail: optionalStringField(message, 'detail')
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
        message: optionalStringField(message, 'message'),
        detail: optionalStringField(message, 'detail')
      };
    case 'action.plan.proposed':
      return {
        ...metadata,
        type: 'action.plan.proposed',
        sessionId: stringField(message, 'sessionId'),
        actionPlan: actionPlanField(message)
      };
    case 'action.plan.approved':
    case 'action.plan.cancelled':
    case 'action.plan.executed':
    case 'action.plan.failed':
      return {
        ...metadata,
        type: message.type,
        sessionId: stringField(message, 'sessionId'),
        planId: stringField(message, 'planId'),
        status: actionPlanStatusField(message, message.type),
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
    case 'session.cancelled':
      return { ...metadata, type: 'session.cancelled', sessionId: stringField(message, 'sessionId') };
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

function actionPlanField(message: Record<string, unknown>) {
  const actionPlan = objectField(message, 'actionPlan');
  return {
    planId: stringField(actionPlan, 'planId'),
    status: 'proposed' as const,
    confirmationSummary: stringField(actionPlan, 'confirmationSummary'),
    commands: arrayField(actionPlan, 'commands').map((item) => {
      const command = objectValue(item, 'actionPlan.commands');
      return {
        kind: stringField(command, 'kind'),
        summary: stringField(command, 'summary'),
        ...optionalObjectField('id', optionalStringField(command, 'id')),
        ...optionalObjectField('operation', optionalStringField(command, 'operation')),
        ...optionalObjectField('title', optionalStringField(command, 'title')),
        ...optionalObjectField('assetKind', optionalStringField(command, 'assetKind')),
        ...optionalObjectField('parentAssetId', optionalStringField(command, 'parentAssetId')),
        ...optionalObjectField('parentTitle', optionalStringField(command, 'parentTitle')),
        ...optionalObjectField('parentKind', optionalStringField(command, 'parentKind')),
        ...optionalObjectField('parentCommandId', optionalStringField(command, 'parentCommandId'))
      };
    }),
    risks: arrayField(actionPlan, 'risks').map((item) => {
      if (typeof item !== 'string' || item.trim().length === 0) {
        throw new Error('Voice action plan risk must be a non-empty string.');
      }
      return item;
    })
  };
}

function optionalObjectField<T extends string>(field: T, value: string | undefined): { readonly [key in T]?: string } {
  return value ? { [field]: value } as { readonly [key in T]?: string } : {};
}

function actionPlanStatusField(
  message: Record<string, unknown>,
  type: 'action.plan.approved' | 'action.plan.cancelled' | 'action.plan.executed' | 'action.plan.failed'
): 'approved' | 'cancelled' | 'executed' | 'failed' {
  const status = stringField(message, 'status');
  switch (type) {
    case 'action.plan.approved':
      if (status === 'approved') {
        return status;
      }
      break;
    case 'action.plan.cancelled':
      if (status === 'cancelled') {
        return status;
      }
      break;
    case 'action.plan.executed':
      if (status === 'executed') {
        return status;
      }
      break;
    case 'action.plan.failed':
      if (status === 'failed') {
        return status;
      }
      break;
  }
  throw new Error('Voice action plan event status did not match the event type.');
}

function arrayField(message: Record<string, unknown>, field: string): readonly unknown[] {
  const value = message[field];
  if (!Array.isArray(value)) {
    throw new Error(`Voice event field ${field} must be an array.`);
  }
  return value;
}

function objectValue(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`Voice event field ${label} must contain objects.`);
  }
  return value as Record<string, unknown>;
}
