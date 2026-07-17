import type {
  RealtimeVoiceTransport,
  RealtimeVoiceTransportInput,
  RealtimeVoiceTransportRunOptions,
  VoiceAssistantResponseKind,
  VoiceActionPlanPhotoApprovalRequest,
  VoiceActionPlanCommandEdit,
  VoiceRealtimeEvent
} from '../../application/voice/RealtimeVoiceSession';
import { isValidVoiceActionPlanProposal, VoiceRealtimeCancelledError } from '../../application/voice/RealtimeVoiceSession';
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

const voiceClientMessage = {
  sessionStart: 'session.start',
  audioChunk: 'audio.chunk',
  audioEnd: 'audio.end',
  sessionCancel: 'session.cancel',
  actionPlanApprove: 'action.plan.approve',
  actionPlanCancel: 'action.plan.cancel'
} as const;

type VoiceClientActionPlanDecisionMessage =
  typeof voiceClientMessage.actionPlanApprove | typeof voiceClientMessage.actionPlanCancel;

const voiceServerMessage = {
  sessionStarted: 'session.started',
  sessionFailed: 'session.failed',
  transcriptDelta: 'transcript.delta',
  transcriptFinal: 'transcript.final',
  agentProgress: 'agent.progress',
  agentDiagnostic: 'agent.diagnostic',
  toolCallStarted: 'tool.call.started',
  toolCallCompleted: 'tool.call.completed',
  toolCallFailed: 'tool.call.failed',
  actionPlanProposed: 'action.plan.proposed',
  actionPlanApproved: 'action.plan.approved',
  actionPlanCancelled: 'action.plan.cancelled',
  actionPlanExecuted: 'action.plan.executed',
  actionPlanFailed: 'action.plan.failed',
  assistantResponseStarted: 'assistant.response.started',
  assistantResponseDelta: 'assistant.response.delta',
  assistantResponseCompleted: 'assistant.response.completed',
  textToSpeechAudioStarted: 'tts.audio.started',
  textToSpeechAudioChunk: 'tts.audio.chunk',
  textToSpeechAudioCompleted: 'tts.audio.completed',
  sessionCompleted: 'session.completed',
  sessionCancelled: 'session.cancelled'
} as const;

type VoiceServerActionPlanTerminalMessage =
  typeof voiceServerMessage.actionPlanApproved |
  typeof voiceServerMessage.actionPlanCancelled |
  typeof voiceServerMessage.actionPlanExecuted |
  typeof voiceServerMessage.actionPlanFailed;

type ActiveRealtimeReviewSession = {
  readonly sendDecision: (type: VoiceClientActionPlanDecisionMessage, planId: string, photos?: readonly VoiceActionPlanPhotoApprovalRequest[], edits?: readonly VoiceActionPlanCommandEdit[]) => void;
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
      let responseCompletedForTurn = false;
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
            type: voiceClientMessage.audioChunk,
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
        socket.send(JSON.stringify({ type: voiceClientMessage.audioEnd, seq: seq++, sessionId }));
      };
      const sendDecision = (type: VoiceClientActionPlanDecisionMessage, planId: string, photos: readonly VoiceActionPlanPhotoApprovalRequest[] = [], edits: readonly VoiceActionPlanCommandEdit[] = []) => {
        if (!sessionId || settled) {
          throw new Error('Voice review session is not active.');
        }
        if (decisionSent) {
          throw new Error('Voice review decision has already been sent.');
        }
        const safePhotos = type === voiceClientMessage.actionPlanApprove
          ? safePhotoApprovalRequests(photos)
          : [];
        decisionSent = true;
        thisTransport.activeReviewSession = null;
        socket.send(JSON.stringify({
          type,
          seq: seq++,
          sessionId,
          planId,
          ...(safePhotos.length > 0 ? { photoAttachments: safePhotos } : {}),
          ...(type === voiceClientMessage.actionPlanApprove && edits.length > 0 ? { commandEdits: edits } : {})
        }));
      };
      const cancelSocketForUser = () => {
        completed = true;
        try {
          if (sessionId) {
            socket.send(JSON.stringify({
              type: voiceClientMessage.sessionCancel,
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
          type: voiceClientMessage.sessionStart,
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
          if (
            message.type === voiceServerMessage.sessionCompleted &&
            !hasPendingActionPlan &&
            !responseCompletedForTurn
          ) {
            throw new Error('Voice session completed without a structured response.');
          }
          lastServerSeq = message.seq;
          if (message.type === voiceServerMessage.sessionStarted) {
            sessionId = message.sessionId;
          }
          await currentOnEvent(message);
          if (message.type === voiceServerMessage.sessionStarted) {
            if (settled || options.signal?.aborted) {
              return;
            }
            sendAudioTurn(input.audioChunksBase64);
          }
          if (message.type === voiceServerMessage.assistantResponseCompleted) {
            lastResponseKind = message.response.kind;
            responseCompletedForTurn = true;
          }
          if (message.type === voiceServerMessage.actionPlanProposed) {
            hasPendingActionPlan = true;
            this.activeReviewSession = { sendDecision };
          }
          if (message.type === voiceServerMessage.sessionCompleted && !hasPendingActionPlan) {
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
                  responseCompletedForTurn = false;
                  lastResponseKind = '';
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
          if (
            message.type === voiceServerMessage.actionPlanCancelled ||
            message.type === voiceServerMessage.actionPlanExecuted ||
            message.type === voiceServerMessage.actionPlanFailed
          ) {
            completed = true;
            this.activeReviewSession = null;
            this.activeFollowUpSession = null;
            followUpResolve?.();
            followUpResolve = null;
            followUpReject = null;
            socket.close();
            settleResolve();
          }
          if (message.type === voiceServerMessage.sessionCancelled) {
            completed = true;
            this.activeReviewSession = null;
            this.activeFollowUpSession = null;
            followUpResolve?.();
            followUpResolve = null;
            followUpReject = null;
            socket.close();
            settleResolve();
          }
          if (message.type === voiceServerMessage.sessionFailed) {
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

  async approveActionPlan(planId: string, photos: readonly VoiceActionPlanPhotoApprovalRequest[] = [], edits: readonly VoiceActionPlanCommandEdit[] = []): Promise<void> {
    if (!this.activeReviewSession) {
      throw new Error('Voice review session is not active.');
    }
    this.activeReviewSession.sendDecision(voiceClientMessage.actionPlanApprove, planId, photos, edits);
  }

  async cancelActionPlan(planId: string): Promise<void> {
    if (!this.activeReviewSession) {
      throw new Error('Voice review session is not active.');
    }
    this.activeReviewSession.sendDecision(voiceClientMessage.actionPlanCancel, planId);
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

function safePhotoApprovalRequests(photos: readonly VoiceActionPlanPhotoApprovalRequest[]): readonly VoiceActionPlanPhotoApprovalRequest[] {
  return photos.map((photo) => {
    const raw = photo as unknown as Record<string, unknown>;
    for (const key of Object.keys(raw)) {
      if (!voicePhotoApprovalRequestFieldAllowed(key)) {
        throw new Error('Voice photo approval metadata may only include reviewed metadata.');
      }
    }
    if (
      typeof raw.commandId !== 'string' ||
      typeof raw.photoIndex !== 'number' ||
      !Number.isInteger(raw.photoIndex) ||
      raw.photoIndex < 0 ||
      typeof raw.fileName !== 'string' ||
      !voicePhotoApprovalContentTypeAllowed(raw.contentType) ||
      typeof raw.sizeBytes !== 'number' ||
      !Number.isInteger(raw.sizeBytes) ||
      raw.sizeBytes <= 0
    ) {
      throw new Error('Voice photo approval metadata may only include reviewed metadata.');
    }
    return {
      commandId: raw.commandId,
      photoIndex: raw.photoIndex,
      fileName: raw.fileName,
      contentType: raw.contentType,
      sizeBytes: raw.sizeBytes
    };
  });
}

function voicePhotoApprovalRequestFieldAllowed(key: string): boolean {
  return key === 'commandId' ||
    key === 'photoIndex' ||
    key === 'fileName' ||
    key === 'contentType' ||
    key === 'sizeBytes';
}

function voicePhotoApprovalContentTypeAllowed(value: unknown): value is VoiceActionPlanPhotoApprovalRequest['contentType'] {
  return value === 'image/jpeg' || value === 'image/png' || value === 'image/webp';
}

function parseServerMessage(raw: string, directUploadPolicy: DirectUploadTargetPolicy): VoiceRealtimeEvent {
  const message = JSON.parse(raw) as Record<string, unknown>;
  const metadata = eventMetadata(message);
  switch (message.type) {
    case voiceServerMessage.sessionStarted:
      return {
        ...metadata,
        type: voiceServerMessage.sessionStarted,
        sessionId: stringField(message, 'sessionId'),
        acceptedInputAudio: acceptedInputAudioField(message),
        acceptedOutputAudio: acceptedOutputAudioField(message),
        acceptedCapabilities: acceptedCapabilitiesField(message)
      };
    case voiceServerMessage.sessionFailed:
      return {
        ...metadata,
        type: voiceServerMessage.sessionFailed,
        sessionId: optionalStringField(message, 'sessionId'),
        code: stringField(message, 'code'),
        message: stringField(message, 'message')
      };
    case voiceServerMessage.transcriptDelta:
    case voiceServerMessage.transcriptFinal:
      return { ...metadata, type: message.type, sessionId: stringField(message, 'sessionId'), text: stringField(message, 'text') };
    case voiceServerMessage.agentProgress:
      return {
        ...metadata,
        type: voiceServerMessage.agentProgress,
        sessionId: stringField(message, 'sessionId'),
        status: stringField(message, 'status'),
        message: stringField(message, 'message')
      };
    case voiceServerMessage.agentDiagnostic:
      return {
        ...metadata,
        type: voiceServerMessage.agentDiagnostic,
        sessionId: stringField(message, 'sessionId'),
        message: stringField(message, 'message'),
        detail: optionalStringField(message, 'detail')
      };
    case voiceServerMessage.toolCallStarted:
    case voiceServerMessage.toolCallCompleted:
    case voiceServerMessage.toolCallFailed:
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
    case voiceServerMessage.actionPlanProposed:
      return {
        ...metadata,
        type: voiceServerMessage.actionPlanProposed,
        sessionId: stringField(message, 'sessionId'),
        actionPlan: actionPlanField(message)
      };
    case voiceServerMessage.actionPlanApproved:
    case voiceServerMessage.actionPlanCancelled:
    case voiceServerMessage.actionPlanExecuted:
    case voiceServerMessage.actionPlanFailed: {
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
    case voiceServerMessage.assistantResponseStarted:
      return { ...metadata, type: voiceServerMessage.assistantResponseStarted, sessionId: stringField(message, 'sessionId'), responseId: stringField(message, 'responseId') };
    case voiceServerMessage.assistantResponseDelta:
      throw new Error('Voice server sent a reserved response delta event.');
    case voiceServerMessage.assistantResponseCompleted: {
      const response = objectField(message, 'response');
      return {
        ...metadata,
        type: voiceServerMessage.assistantResponseCompleted,
        sessionId: stringField(message, 'sessionId'),
        response: {
          kind: assistantResponseKindField(response),
          spokenResponse: stringField(response, 'spokenResponse'),
          displayResponse: stringField(response, 'displayResponse'),
          artifacts: voiceResponseArtifactsField(response)
        }
      };
    }
    case voiceServerMessage.textToSpeechAudioStarted: {
      const format = objectField(message, 'format');
      return { ...metadata, type: voiceServerMessage.textToSpeechAudioStarted, sessionId: stringField(message, 'sessionId'), mimeType: stringField(format, 'mimeType') };
    }
    case voiceServerMessage.textToSpeechAudioChunk:
      return {
        ...metadata,
        type: voiceServerMessage.textToSpeechAudioChunk,
        sessionId: stringField(message, 'sessionId'),
        chunkId: stringField(message, 'chunkId'),
        audioBase64: stringField(message, 'audioBase64'),
        isFinalChunk: booleanField(message, 'isFinalChunk')
      };
    case voiceServerMessage.textToSpeechAudioCompleted:
      return { ...metadata, type: voiceServerMessage.textToSpeechAudioCompleted, sessionId: stringField(message, 'sessionId') };
    case voiceServerMessage.sessionCompleted:
      return { ...metadata, type: voiceServerMessage.sessionCompleted, sessionId: stringField(message, 'sessionId') };
    case voiceServerMessage.sessionCancelled:
      return { ...metadata, type: voiceServerMessage.sessionCancelled, sessionId: stringField(message, 'sessionId') };
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
  if (message.type === voiceServerMessage.sessionStarted) {
    if (currentSessionId && message.sessionId !== currentSessionId) {
      throw new Error('Voice server event changed session.');
    }
    return;
  }
  if (!currentSessionId) {
    if (message.type !== voiceServerMessage.sessionFailed) {
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
  const proposal = {
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
  if (!isValidVoiceActionPlanProposal(proposal)) {
    throw new Error('Voice action plan could not be reviewed safely.');
  }
  return proposal;
}

function assistantResponseKindField(message: Record<string, unknown>): VoiceAssistantResponseKind {
  const value = stringField(message, 'kind');
  switch (value) {
    case 'answer':
    case 'clarification':
    case 'unsupported_action':
    case 'safe_failure':
      return value;
    default:
      throw new Error('Voice structured response kind is not supported.');
  }
}

function voiceResponseArtifactsField(message: Record<string, unknown>) {
  const raw = message.artifacts;
  if (raw === undefined) {
    return [];
  }
  if (!Array.isArray(raw) || raw.length > 16) {
    throw new Error('Voice response artifacts must be a bounded array.');
  }
  const seen = new Set<string>();
  return raw.map((item) => {
    const artifact = objectValue(item, 'response.artifacts');
    const allowed = new Set(['type', 'assetId', 'title', 'assetKind', 'context']);
    if (Object.keys(artifact).some((key) => !allowed.has(key))) {
      throw new Error('Voice response artifact contains unsupported fields.');
    }
    const type = stringField(artifact, 'type');
    const assetIdValue = stringField(artifact, 'assetId').trim();
    const title = stringField(artifact, 'title').trim();
    const assetKind = voiceResponseAssetKind(stringField(artifact, 'assetKind'));
    const rawContext = artifact.context;
    const context = typeof rawContext === 'string' ? rawContext.trim() : undefined;
    if (
      type !== 'asset_reference' || assetIdValue.length > 200 || title.length > 500 ||
      (rawContext !== undefined && (typeof rawContext !== 'string' || !context || context.length > 500 || context !== rawContext)) ||
      seen.has(assetIdValue)
    ) {
      throw new Error('Voice response artifact is invalid.');
    }
    seen.add(assetIdValue);
    return { type: 'asset_reference' as const, assetId: assetIdValue, title, assetKind, ...(context ? { context } : {}) };
  });
}

function voiceResponseAssetKind(value: string): 'item' | 'container' | 'location' {
  if (value === 'item' || value === 'container' || value === 'location') {
    return value;
  }
  throw new Error('Voice response artifact asset kind is invalid.');
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
  type: VoiceServerActionPlanTerminalMessage
): 'approved' | 'cancelled' | 'executed' | 'failed' {
  const status = stringField(message, 'status');
  switch (type) {
    case voiceServerMessage.actionPlanApproved:
      if (status === 'approved') {
        return status;
      }
      break;
    case voiceServerMessage.actionPlanCancelled:
      if (status === 'cancelled') {
        return status;
      }
      break;
    case voiceServerMessage.actionPlanExecuted:
      if (status === 'executed') {
        return status;
      }
      break;
    case voiceServerMessage.actionPlanFailed:
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
