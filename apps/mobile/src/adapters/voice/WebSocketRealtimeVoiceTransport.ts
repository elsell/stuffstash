import type {
  RealtimeVoiceTransport,
  RealtimeVoiceTransportInput,
  RealtimeVoiceTransportRunOptions,
  VoiceActionPlanPhotoApprovalRequest,
  VoiceRealtimeEvent
} from '../../application/voice/RealtimeVoiceSession';
import { VoiceRealtimeCancelledError } from '../../application/voice/RealtimeVoiceSession';
import {
  directUploadMethod,
  isDirectUploadTargetSupported,
  type DirectUploadTargetPolicy
} from '../uploads/DirectUploadPolicy';

type VoiceWebSocket = {
  onopen: (() => void) | null;
  onmessage: ((event: { readonly data: string }) => void) | null;
  onerror: ((event: unknown) => void) | null;
  onclose: ((event?: { readonly code?: number; readonly reason?: string }) => void) | null;
  send(payload: string): void;
  close(): void;
};

type VoiceWebSocketFactory = (url: string, headers: Record<string, string>) => VoiceWebSocket;

const requiredRealtimeVoiceCapabilities = ['speech_to_text', 'language_inference', 'text_to_speech'] as const;

type ActiveRealtimeReviewSession = {
  readonly sendDecision: (type: 'action.plan.approve' | 'action.plan.cancel', planId: string, photos?: readonly VoiceActionPlanPhotoApprovalRequest[]) => void;
};

type ActiveRealtimeFollowUpSession = {
  readonly sendAudio: (
    audioChunksBase64: readonly string[],
    onEvent?: (event: VoiceRealtimeEvent) => Promise<void>,
    options?: RealtimeVoiceTransportRunOptions
  ) => Promise<void>;
  readonly close: () => void;
};

export type WebSocketRealtimeVoiceTransportOptions = {
  readonly apiBaseUrl: string;
  readonly tokenProvider: () => string | Promise<string>;
  readonly diagnosticsEnabled?: boolean;
  readonly directUploadPolicy?: DirectUploadTargetPolicy;
  readonly webSocketFactory?: VoiceWebSocketFactory;
};

export class WebSocketRealtimeVoiceTransport implements RealtimeVoiceTransport {
  private readonly apiBaseUrl: string;
  private readonly tokenProvider: () => string | Promise<string>;
  private readonly diagnosticsEnabled: boolean;
  private readonly directUploadPolicy: DirectUploadTargetPolicy;
  private readonly webSocketFactory: VoiceWebSocketFactory;
  private activeReviewSession: ActiveRealtimeReviewSession | null = null;
  private activeFollowUpSession: ActiveRealtimeFollowUpSession | null = null;

  constructor(options: WebSocketRealtimeVoiceTransportOptions) {
    this.apiBaseUrl = options.apiBaseUrl.replace(/\/+$/, '');
    this.tokenProvider = options.tokenProvider;
    this.diagnosticsEnabled = options.diagnosticsEnabled ?? false;
    this.directUploadPolicy = options.directUploadPolicy ?? {};
    this.webSocketFactory = options.webSocketFactory ?? createReactNativeWebSocket;
  }

  async run(
    input: RealtimeVoiceTransportInput,
    onEvent: (event: VoiceRealtimeEvent) => Promise<void>,
    options: RealtimeVoiceTransportRunOptions = {}
  ): Promise<void> {
    const maybeToken = this.tokenProvider();
    const token = typeof maybeToken === 'string' ? maybeToken : await maybeToken;
    if (!token) {
      throw new Error('Sign in before starting a voice session.');
    }
    const socket = this.webSocketFactory(realtimeVoiceUrl(this.apiBaseUrl), {
      Authorization: `Bearer ${token}`
    });
    const thisTransport = this;

    await new Promise<void>((resolve, reject) => {
      let sessionId = '';
      let seq = 1;
      let lastServerSeq = 0;
      let completed = false;
      let hasPendingActionPlan = false;
      let lastResponseKind = '';
      let followUpPending = false;
      let followUpResolve: (() => void) | null = null;
      let followUpReject: ((error: Error) => void) | null = null;
      let currentOnEvent = onEvent;
      let settled = false;
      let decisionSent = false;
      let messageChain = Promise.resolve();
      const sendAudioTurn = (audioChunksBase64: readonly string[], signal?: AbortSignal) => {
        audioChunksBase64.forEach((audioBase64, index) => {
          if (signal?.aborted) {
            throw new VoiceRealtimeCancelledError();
          }
          socket.send(JSON.stringify({
            type: 'audio.chunk',
            seq: seq++,
            sessionId,
            chunkId: `mobile-${seq}-${index + 1}`,
            audioBase64,
            isFinalChunk: index === audioChunksBase64.length - 1
          }));
        });
        if (signal?.aborted) {
          throw new VoiceRealtimeCancelledError();
        }
        socket.send(JSON.stringify({ type: 'audio.end', seq: seq++, sessionId }));
      };
      const sendDecision = (type: 'action.plan.approve' | 'action.plan.cancel', planId: string, photos: readonly VoiceActionPlanPhotoApprovalRequest[] = []) => {
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
          planId,
          ...(type === 'action.plan.approve' && photos.length > 0 ? { photoAttachments: photos } : {})
        }));
      };
      const cancelSocketForUser = () => {
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
        }
      };
      const abortHandler = () => {
        if (settled) {
          return;
        }
        cancelSocketForUser();
        settleReject(new VoiceRealtimeCancelledError());
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
          if (thisTransport.activeFollowUpSession?.close === closeFollowUpSession) {
            thisTransport.activeFollowUpSession = null;
          }
          options.signal?.removeEventListener('abort', abortHandler);
          reject(error);
        }
      }

      const closeFollowUpSession = () => {
        thisTransport.activeFollowUpSession = null;
        socket.close();
      };

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
          if (thisTransport.activeFollowUpSession?.close === closeFollowUpSession) {
            thisTransport.activeFollowUpSession = null;
          }
          if (followUpPending) {
            const error = new Error(prematureCloseMessage(event));
            followUpPending = false;
            followUpResolve = null;
            followUpReject?.(error);
            followUpReject = null;
            return;
          }
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
          requestedCapabilities: [...requiredRealtimeVoiceCapabilities],
          developerDiagnostics: this.diagnosticsEnabled,
          ...(input.clientCorrelationId ? { clientCorrelationId: input.clientCorrelationId } : {}),
          inputAudio: input.inputAudio,
          outputAudio: { mimeTypes: input.outputAudioMimeTypes }
        }));
      };
      socket.onmessage = (event) => {
        messageChain = messageChain.then(async () => {
          if (settled) {
            if (!followUpPending) {
              return;
            }
          }
          const message = parseServerMessage(event.data, thisTransport.directUploadPolicy);
          validateServerMessage(message, sessionId, lastServerSeq);
          lastServerSeq = message.seq;
          if (message.type === 'session.started') {
            sessionId = message.sessionId;
          }
          await currentOnEvent(message);
          if (message.type === 'session.started') {
            if (settled || options.signal?.aborted) {
              return;
            }
            sendAudioTurn(input.audioChunksBase64);
          }
          if (message.type === 'assistant.response.completed') {
            lastResponseKind = message.response.kind;
          }
          if (message.type === 'action.plan.proposed') {
            hasPendingActionPlan = true;
            this.activeReviewSession = { sendDecision };
          }
          if (message.type === 'session.completed' && !hasPendingActionPlan) {
            completed = true;
            if (lastResponseKind === 'clarification') {
              followUpPending = false;
              followUpResolve?.();
              followUpResolve = null;
              followUpReject = null;
              this.activeFollowUpSession = {
                sendAudio: (audioChunksBase64, followUpOnEvent, followUpOptions) => {
                  if (followUpOptions?.signal?.aborted) {
                    thisTransport.activeFollowUpSession = null;
                    cancelSocketForUser();
                    return Promise.reject(new VoiceRealtimeCancelledError());
                  }
                  currentOnEvent = followUpOnEvent ?? onEvent;
                  followUpPending = true;
                  const abortFollowUpHandler = () => {
                    if (!followUpPending) {
                      return;
                    }
                    followUpPending = false;
                    thisTransport.activeFollowUpSession = null;
                    cancelSocketForUser();
                    const rejectFollowUp = followUpReject;
                    followUpResolve = null;
                    followUpReject = null;
                    rejectFollowUp?.(new VoiceRealtimeCancelledError());
                  };
                  const promise = new Promise<void>((resolveFollowUp, rejectFollowUp) => {
                    followUpResolve = () => {
                      followUpOptions?.signal?.removeEventListener('abort', abortFollowUpHandler);
                      resolveFollowUp();
                    };
                    followUpReject = (error) => {
                      followUpOptions?.signal?.removeEventListener('abort', abortFollowUpHandler);
                      rejectFollowUp(error);
                    };
                  });
                  followUpOptions?.signal?.addEventListener('abort', abortFollowUpHandler, { once: true });
                  try {
                    sendAudioTurn(audioChunksBase64, followUpOptions?.signal);
                  } catch (error) {
                    followUpOptions?.signal?.removeEventListener('abort', abortFollowUpHandler);
                    followUpPending = false;
                    followUpResolve = null;
                    followUpReject = null;
                    return Promise.reject(error instanceof Error ? error : new Error(String(error)));
                  }
                  return promise;
                },
                close: closeFollowUpSession
              };
            } else {
              followUpPending = false;
              socket.close();
              this.activeFollowUpSession = null;
              followUpResolve?.();
              followUpResolve = null;
              followUpReject = null;
            }
            settleResolve();
          }
          if (message.type === 'action.plan.cancelled' || message.type === 'action.plan.executed' || message.type === 'action.plan.failed') {
            completed = true;
            this.activeReviewSession = null;
            this.activeFollowUpSession = null;
            followUpResolve?.();
            followUpResolve = null;
            followUpReject = null;
            socket.close();
            settleResolve();
          }
          if (message.type === 'session.cancelled') {
            completed = true;
            this.activeReviewSession = null;
            this.activeFollowUpSession = null;
            followUpResolve?.();
            followUpResolve = null;
            followUpReject = null;
            socket.close();
            settleResolve();
          }
          if (message.type === 'session.failed') {
            completed = true;
            this.activeReviewSession = null;
            this.activeFollowUpSession = null;
            followUpResolve?.();
            followUpResolve = null;
            followUpReject = null;
            socket.close();
            settleResolve();
          }
        }).catch(settleReject);
      };
    });
  }

  canSendFollowUpAudio(): boolean {
    return this.activeFollowUpSession !== null;
  }

  async sendFollowUpAudio(
    audioChunksBase64: readonly string[],
    onEvent?: (event: VoiceRealtimeEvent) => Promise<void>,
    options?: RealtimeVoiceTransportRunOptions
  ): Promise<void> {
    if (!this.activeFollowUpSession) {
      throw new Error('Voice follow-up session is not active.');
    }
    await this.activeFollowUpSession.sendAudio(audioChunksBase64, onEvent, options);
  }

  async approveActionPlan(planId: string, photos: readonly VoiceActionPlanPhotoApprovalRequest[] = []): Promise<void> {
    if (!this.activeReviewSession) {
      throw new Error('Voice review session is not active.');
    }
    this.activeReviewSession.sendDecision('action.plan.approve', planId, photos);
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

function parseServerMessage(raw: string, directUploadPolicy: DirectUploadTargetPolicy): VoiceRealtimeEvent {
  const message = JSON.parse(raw) as Record<string, unknown>;
  const metadata = eventMetadata(message);
  switch (message.type) {
    case 'session.started':
      return {
        ...metadata,
        type: 'session.started',
        sessionId: stringField(message, 'sessionId'),
        acceptedInputAudio: acceptedInputAudioField(message),
        acceptedOutputAudio: acceptedOutputAudioField(message),
        acceptedCapabilities: acceptedCapabilitiesField(message)
      };
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
    case 'action.plan.failed': {
      const commandResults = actionPlanCommandResultsField(message);
      const attachmentUploadIntents = actionPlanAttachmentUploadIntentsField(message, directUploadPolicy);
      return {
        ...metadata,
        type: message.type,
        sessionId: stringField(message, 'sessionId'),
        planId: stringField(message, 'planId'),
        status: actionPlanStatusField(message, message.type),
        message: optionalStringField(message, 'message'),
        ...(commandResults ? { commandResults } : {}),
        ...(attachmentUploadIntents ? { attachmentUploadIntents } : {})
      };
    }
    case 'assistant.response.started':
      return { ...metadata, type: 'assistant.response.started', sessionId: stringField(message, 'sessionId'), responseId: stringField(message, 'responseId') };
    case 'assistant.response.delta':
      throw new Error('Voice server sent a reserved response delta event.');
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
        audioBase64: stringField(message, 'audioBase64'),
        isFinalChunk: booleanField(message, 'isFinalChunk')
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

function nonNegativeNumberField(message: Record<string, unknown>, field: string): number {
  const value = message[field];
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 0) {
    throw new Error(`Voice event field ${field} must be a non-negative integer.`);
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

function booleanField(message: Record<string, unknown>, field: string): boolean {
  const value = message[field];
  if (typeof value !== 'boolean') {
    throw new Error(`Voice event field ${field} must be a boolean.`);
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

function acceptedInputAudioField(message: Record<string, unknown>) {
  const acceptedInputAudio = objectField(message, 'acceptedInputAudio');
  return {
    mimeType: stringField(acceptedInputAudio, 'mimeType'),
    sampleRate: numberField(acceptedInputAudio, 'sampleRate'),
    channels: numberField(acceptedInputAudio, 'channels')
  };
}

function acceptedOutputAudioField(message: Record<string, unknown>) {
  const acceptedOutputAudio = objectField(message, 'acceptedOutputAudio');
  return {
    mimeTypes: stringArrayField(acceptedOutputAudio, 'mimeTypes')
  };
}

function acceptedCapabilitiesField(message: Record<string, unknown>): readonly string[] {
  let capabilities: readonly string[];
  try {
    capabilities = stringArrayField(message, 'acceptedCapabilities');
  } catch {
    throw new Error('Voice event field acceptedCapabilities must match requested voice capabilities.');
  }
  const matchesRequested = capabilities.length === requiredRealtimeVoiceCapabilities.length &&
    requiredRealtimeVoiceCapabilities.every((capability, index) => capabilities[index] === capability);
  if (!matchesRequested) {
    throw new Error('Voice event field acceptedCapabilities must match requested voice capabilities.');
  }
  return capabilities;
}

function actionPlanCommandResultsField(message: Record<string, unknown>) {
  const raw = message.commandResults;
  if (raw === undefined) {
    return undefined;
  }
  if (!Array.isArray(raw)) {
    throw new Error('Voice event field commandResults must be an array.');
  }
  return raw.map((item) => {
    const result = objectValue(item, 'commandResults');
    return {
      commandId: stringField(result, 'commandId'),
      assetId: stringField(result, 'assetId'),
      operation: stringField(result, 'operation'),
      assetKind: stringField(result, 'assetKind')
    };
  });
}

function actionPlanAttachmentUploadIntentsField(message: Record<string, unknown>, directUploadPolicy: DirectUploadTargetPolicy) {
  const raw = message.attachmentUploadIntents;
  if (raw === undefined) {
    return undefined;
  }
  if (!Array.isArray(raw)) {
    throw new Error('Voice event field attachmentUploadIntents must be an array.');
  }
  return raw.map((item) => {
    const intent = objectValue(item, 'attachmentUploadIntents');
    const directUpload = objectField(intent, 'directUpload');
    return {
      commandId: stringField(intent, 'commandId'),
      photoIndex: nonNegativeNumberField(intent, 'photoIndex'),
      assetId: stringField(intent, 'assetId'),
      fileName: stringField(intent, 'fileName'),
      contentType: photoContentTypeField(intent, 'contentType'),
      sizeBytes: numberField(intent, 'sizeBytes'),
      directUpload: {
        uploadId: stringField(directUpload, 'uploadId'),
        attachmentId: stringField(directUpload, 'attachmentId'),
        method: directUploadMethodField(directUpload, 'method'),
        url: directUploadURLField(directUpload, 'url', directUploadPolicy),
        headers: stringRecordField(directUpload, 'headers'),
        formFields: stringRecordField(directUpload, 'formFields'),
        expiresAt: stringField(directUpload, 'expiresAt')
      }
    };
  });
}

function photoContentTypeField(message: Record<string, unknown>, field: string): 'image/jpeg' | 'image/png' | 'image/webp' {
  const value = stringField(message, field);
  if (value === 'image/jpeg' || value === 'image/png' || value === 'image/webp') {
    return value;
  }
  throw new Error(`Voice event field ${field} has unsupported photo content type.`);
}

function directUploadMethodField(message: Record<string, unknown>, field: string): 'POST' | 'PUT' | 'PATCH' {
  try {
    return directUploadMethod(stringField(message, field));
  } catch {
    throw new Error(`Voice event field ${field} has unsupported direct upload method.`);
  }
}

function directUploadURLField(message: Record<string, unknown>, field: string, directUploadPolicy: DirectUploadTargetPolicy): string {
  const value = stringField(message, field);
  if (isDirectUploadTargetSupported(value, directUploadPolicy)) {
    return value;
  }
  throw new Error(`Voice event field ${field} has unsupported direct upload URL.`);
}

function stringRecordField(message: Record<string, unknown>, field: string): Readonly<Record<string, string>> {
  const raw = objectField(message, field);
  const values: Record<string, string> = {};
  for (const [key, value] of Object.entries(raw)) {
    if (typeof value !== 'string') {
      throw new Error(`Voice event field ${field} must be a string record.`);
    }
    values[key] = value;
  }
  return values;
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

function stringArrayField(message: Record<string, unknown>, field: string): readonly string[] {
  const value = arrayField(message, field);
  if (value.some((item) => typeof item !== 'string' || item.trim().length === 0)) {
    throw new Error(`Voice event field ${field} must be a string array.`);
  }
  return value as readonly string[];
}

function objectValue(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`Voice event field ${label} must contain objects.`);
  }
  return value as Record<string, unknown>;
}
