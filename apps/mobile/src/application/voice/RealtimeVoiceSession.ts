import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';
import type { CreateInventoryAssetPhotoInput } from '../home/InventorySummaryRepository';
import { assetId } from '../../domain/assets/AssetSummary';
import type { InventoryId, TenantId } from '../../domain/inventories/InventorySummary';

export type RecordedVoiceAudio = {
  readonly mimeType: 'audio/mp4';
  readonly sampleRate: number;
  readonly channels: number;
  readonly chunksBase64: readonly string[];
};

export interface VoiceAudioRecorder {
  start(): Promise<void>;
  stop(): Promise<RecordedVoiceAudio>;
  cancel(): Promise<void>;
  recordingLevel(): number;
}

export interface VoiceAudioPlayer {
  playChunk(audioBase64: string, mimeType: string): Promise<void>;
  stop(): Promise<void>;
}

export type RealtimeVoiceTransportInput = {
  readonly tenantId: string;
  readonly inventoryId: string;
  readonly source: 'mobile_voice';
  readonly inputAudio: {
    readonly mimeType: string;
    readonly sampleRate: number;
    readonly channels: number;
  };
  readonly outputAudioMimeTypes: readonly string[];
  readonly audioChunksBase64: readonly string[];
};

export interface RealtimeVoiceTransport {
  run(
    input: RealtimeVoiceTransportInput,
    onEvent: (event: VoiceRealtimeEvent) => Promise<void>,
    options?: RealtimeVoiceTransportRunOptions
  ): Promise<void>;
  canSendFollowUpAudio(): boolean;
  sendFollowUpAudio(audioChunksBase64: readonly string[], onEvent?: (event: VoiceRealtimeEvent) => Promise<void>): Promise<void>;
  approveActionPlan(planId: string, photos?: readonly VoiceActionPlanPhotoApprovalRequest[]): Promise<void>;
  cancelActionPlan(planId: string): Promise<void>;
}

export type RealtimeVoiceTransportRunOptions = {
  readonly signal?: AbortSignal;
};

export type RealtimeVoiceSessionControllerOptions = {
  readonly diagnosticsEnabled?: boolean;
  readonly readinessChecker?: VoiceProviderReadinessChecker;
};

export interface VoiceProviderReadinessChecker {
  assertReady(): Promise<void>;
}

type VoiceRealtimeEventMetadata = {
  readonly seq: number;
  readonly sessionId?: string;
};

export type VoiceRealtimeEvent = VoiceRealtimeEventMetadata & (
  | {
      readonly type: 'session.started';
      readonly sessionId: string;
      readonly acceptedInputAudio: {
        readonly mimeType: string;
        readonly sampleRate: number;
        readonly channels: number;
      };
      readonly acceptedOutputAudio: {
        readonly mimeTypes: readonly string[];
      };
      readonly acceptedCapabilities?: readonly string[];
    }
  | { readonly type: 'session.failed'; readonly code: string; readonly message: string }
  | { readonly type: 'transcript.delta'; readonly text: string }
  | { readonly type: 'transcript.final'; readonly text: string }
  | { readonly type: 'agent.progress'; readonly status: string; readonly message: string }
  | { readonly type: 'agent.diagnostic'; readonly message: string; readonly detail?: string }
  | {
      readonly type: 'tool.call.started' | 'tool.call.completed' | 'tool.call.failed';
      readonly toolCallId: string;
      readonly toolLabel: string;
      readonly status?: string;
      readonly code?: string;
      readonly message?: string;
      readonly detail?: string;
    }
  | { readonly type: 'action.plan.proposed'; readonly actionPlan: VoiceActionPlanProposal }
  | {
      readonly type: 'action.plan.approved' | 'action.plan.cancelled' | 'action.plan.executed' | 'action.plan.failed';
      readonly planId: string;
      readonly status: 'approved' | 'cancelled' | 'executed' | 'failed';
      readonly message?: string;
      readonly commandResults?: readonly VoiceActionPlanCommandResult[];
      readonly attachmentUploadIntents?: readonly VoiceActionPlanAttachmentUploadIntent[];
    }
  | { readonly type: 'assistant.response.started'; readonly responseId: string }
  | {
      readonly type: 'assistant.response.completed';
      readonly response: {
        readonly kind: string;
        readonly spokenResponse: string;
        readonly displayResponse: string;
      };
    }
  | { readonly type: 'tts.audio.started'; readonly mimeType: string }
  | { readonly type: 'tts.audio.chunk'; readonly chunkId: string; readonly audioBase64: string; readonly isFinalChunk: boolean }
  | { readonly type: 'tts.audio.completed' }
  | { readonly type: 'session.completed' }
  | { readonly type: 'session.cancelled' }
);

export type VoiceActionPlanProposal = {
  readonly planId: string;
  readonly status: 'proposed' | 'approved' | 'cancelled' | 'executed' | 'failed';
  readonly confirmationSummary: string;
  readonly commands: readonly VoiceActionPlanCommand[];
  readonly risks: readonly string[];
};

type VoiceActionPlanStatus = VoiceActionPlanProposal['status'];

export type VoiceActionPlanCommand = {
  readonly id?: string;
  readonly kind: string;
  readonly summary: string;
  readonly operation?: string;
  readonly title?: string;
  readonly assetKind?: string;
  readonly parentAssetId?: string;
  readonly parentTitle?: string;
  readonly parentKind?: string;
  readonly parentCommandId?: string;
};

export type VoiceActionPlanCommandResult = {
  readonly commandId: string;
  readonly assetId: string;
  readonly operation: string;
  readonly assetKind: string;
};

export type VoiceActionPlanAttachmentUploadIntent = {
  readonly commandId: string;
  readonly photoIndex: number;
  readonly assetId: string;
  readonly fileName: string;
  readonly contentType: CreateInventoryAssetPhotoInput['contentType'];
  readonly sizeBytes: number;
  readonly directUpload: NonNullable<CreateInventoryAssetPhotoInput['directUpload']>;
};

export type VoiceActionPlanPhotoApprovalRequest = {
  readonly commandId: string;
  readonly photoIndex: number;
  readonly fileName: string;
  readonly contentType: CreateInventoryAssetPhotoInput['contentType'];
  readonly sizeBytes: number;
};

export type VoiceActionPlanPhotoDrafts = Record<string, readonly CreateInventoryAssetPhotoInput[]>;

type VoiceActionPlanExecutedEvent = VoiceRealtimeEventMetadata & {
  readonly type: 'action.plan.executed';
  readonly planId: string;
  readonly status: 'executed';
  readonly message?: string;
  readonly commandResults?: readonly VoiceActionPlanCommandResult[];
  readonly attachmentUploadIntents?: readonly VoiceActionPlanAttachmentUploadIntent[];
};

export type VoiceRealtimeState = {
  readonly status: 'ready' | 'listening' | 'review' | 'processing' | 'speaking' | 'completed' | 'cancelled' | 'failed';
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly actionPlan?: VoiceActionPlanProposal;
  readonly reviewDecisionPending?: boolean;
  readonly progressSteps?: readonly string[];
  readonly partialTranscript?: string;
  readonly transcript?: string;
  readonly spokenResponse?: string;
  readonly responseKind?: string;
  readonly clarificationFollowUpAvailable?: boolean;
  readonly conversationPhase?: VoiceConversationPhase;
  readonly progressLabel?: string;
  readonly debugEvents: readonly VoiceSafeDiagnosticEvent[];
  readonly failureCode?: VoiceRealtimeFailureCode;
  readonly errorMessage?: string;
  readonly photoAttachmentStatus?: VoicePhotoAttachmentStatus;
  readonly recordingLevel?: number;
};

export type VoiceConversationPhase =
  | 'understanding'
  | 'exploring'
  | 'planning'
  | 'reviewing'
  | 'answering'
  | 'recovering';

export type VoicePhotoAttachmentStatus = {
  readonly status: 'attached' | 'partial_failed' | 'failed';
  readonly message: string;
  readonly canRetry?: boolean;
};

export type VoiceRealtimeFailureCode =
  | 'provider_readiness'
  | 'speech_to_text_failed'
  | 'language_inference_failed'
  | 'text_to_speech_failed'
  | 'clarification_turn_limit'
  | 'voice_failed';

export class VoiceRealtimeCancelledError extends Error {
  readonly code = 'voice_cancelled';

  constructor() {
    super('Voice session cancelled.');
  }
}

export type VoiceRealtimeStateHandler = (state: VoiceRealtimeState) => void;

export type VoiceSafeDiagnosticEvent = {
  readonly label: string;
  readonly status: string;
  readonly detail?: string;
};

export class RealtimeVoiceSessionController {
  private currentContext: { readonly tenantId: TenantId; readonly inventoryId: InventoryId; readonly tenantName: string; readonly inventoryName: string } | null = null;
  private recordingStarted = false;
  private activeRunAbortController: AbortController | null = null;
  private activeSessionGeneration = 0;
  private cancelledThroughSessionGeneration = 0;
  private pendingPhotoDraftsByPlanId = new Map<string, VoiceActionPlanPhotoDrafts>();
  private pendingPhotoRetriesByPlanId = new Map<string, VoiceActionPlanPhotoRetry>();
  private ttsMimeType = 'audio/mpeg';

  constructor(
    private readonly inventories: InventorySummaryRepository,
    private readonly recorder: VoiceAudioRecorder,
    private readonly transport: RealtimeVoiceTransport,
    private readonly player: VoiceAudioPlayer,
    private readonly options: RealtimeVoiceSessionControllerOptions = {}
  ) {}

  async start(): Promise<VoiceRealtimeState> {
    const context = await this.selectedInventoryContext();
    await this.options.readinessChecker?.assertReady();
    this.activeSessionGeneration++;
    this.currentContext = context;
    await this.player.stop();
    await this.recorder.start();
    this.recordingStarted = true;
    return {
      status: 'listening',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Listening',
      recordingLevel: this.recordingLevel(),
      debugEvents: []
    };
  }

  recordingLevel(): number {
    return boundedRecordingLevel(this.recorder.recordingLevel());
  }

  async stop(onState?: VoiceRealtimeStateHandler): Promise<readonly VoiceRealtimeState[]> {
    if (!this.recordingStarted) {
      throw new Error('Voice recording has not started.');
    }

    const generation = this.activeSessionGeneration;
    const context = this.currentContext ?? (await this.selectedInventoryContext());
    const recorded = await this.recorder.stop();
    this.recordingStarted = false;
    if (this.isSessionGenerationCancelled(generation)) {
      throw new VoiceRealtimeCancelledError();
    }
    const states: VoiceRealtimeState[] = [{
      status: 'processing',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Sending audio',
      progressSteps: ['Sending audio'],
      debugEvents: []
    }];
    onState?.(states[0]);

    const abortController = new AbortController();
    this.activeRunAbortController = abortController;
    try {
      await this.transport.run({
        tenantId: context.tenantId,
        inventoryId: context.inventoryId,
        source: 'mobile_voice',
        inputAudio: {
          mimeType: recorded.mimeType,
          sampleRate: recorded.sampleRate,
          channels: recorded.channels
        },
        outputAudioMimeTypes: ['audio/mpeg'],
        audioChunksBase64: recorded.chunksBase64
      }, async (event) => {
        if (this.isSessionGenerationCancelled(generation)) {
          throw new VoiceRealtimeCancelledError();
        }
        const previous = states[states.length - 1];
        const next = await this.reduceEvent(previous, event);
        states.push(next);
        onState?.(next);
      }, { signal: abortController.signal });
    } finally {
      if (this.activeRunAbortController === abortController) {
        this.activeRunAbortController = null;
      }
    }
    this.syncFinalClarificationFollowUpAvailability(states, onState);

    return states;
  }

  canSendFollowUpAudio(): boolean {
    return this.transport.canSendFollowUpAudio();
  }

  async startFollowUp(): Promise<VoiceRealtimeState> {
    if (!this.transport.canSendFollowUpAudio()) {
      throw new Error('Voice follow-up session is not active.');
    }
    const context = this.currentContext ?? (await this.selectedInventoryContext());
    await this.options.readinessChecker?.assertReady();
    await this.player.stop();
    await this.recorder.start();
    this.recordingStarted = true;
    return {
      status: 'listening',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Listening',
      recordingLevel: this.recordingLevel(),
      responseKind: 'clarification',
      clarificationFollowUpAvailable: true,
      debugEvents: []
    };
  }

  async stopFollowUp(onState?: VoiceRealtimeStateHandler): Promise<readonly VoiceRealtimeState[]> {
    if (!this.recordingStarted) {
      throw new Error('Voice recording has not started.');
    }
    const context = this.currentContext ?? (await this.selectedInventoryContext());
    const recorded = await this.recorder.stop();
    this.recordingStarted = false;
    const states: VoiceRealtimeState[] = [{
      status: 'processing',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Sending audio',
      progressSteps: ['Sending audio'],
      debugEvents: []
    }];
    onState?.(states[0]);
    await this.transport.sendFollowUpAudio(recorded.chunksBase64, async (event) => {
      const previous = states[states.length - 1];
      const next = await this.reduceEvent(previous, event);
      states.push(next);
      onState?.(next);
    });
    if (states.length === 1) {
      states.push(withProgressStep(states[0], 'Done', { status: 'completed' }));
      onState?.(states[1]);
    }
    this.syncFinalClarificationFollowUpAvailability(states, onState);
    return states;
  }

  async cancel(): Promise<VoiceRealtimeState> {
    this.cancelledThroughSessionGeneration = Math.max(
      this.cancelledThroughSessionGeneration,
      this.activeSessionGeneration
    );
    const context = this.currentContext ?? (await this.selectedInventoryContext());
    if (this.recordingStarted) {
      this.recordingStarted = false;
      await this.recorder.cancel();
    }
    this.activeRunAbortController?.abort();
    await this.player.stop();
    return {
      status: 'cancelled',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Cancelled',
      progressSteps: ['Cancelled'],
      debugEvents: []
    };
  }

  async approveActionPlan(planId: string, photoDrafts: VoiceActionPlanPhotoDrafts = {}): Promise<void> {
    const safePlanId = safeBoundedText(planId, 80);
    const boundedDrafts = boundedPhotoDrafts(photoDrafts);
    validatePhotoApprovalMetadata(boundedDrafts);
    if (Object.keys(boundedDrafts).length > 0) {
      this.pendingPhotoDraftsByPlanId.set(safePlanId, boundedDrafts);
    }
    try {
      await this.transport.approveActionPlan(safePlanId, photoApprovalRequests(boundedDrafts));
    } catch (error) {
      this.pendingPhotoDraftsByPlanId.delete(safePlanId);
      throw error;
    }
  }

  async cancelActionPlan(planId: string): Promise<void> {
    await this.transport.cancelActionPlan(safeBoundedText(planId, 80));
  }

  async retryPhotoAttachments(planId: string): Promise<VoicePhotoAttachmentStatus> {
    const safePlanId = safeBoundedText(planId, 80);
    const retry = this.pendingPhotoRetriesByPlanId.get(safePlanId);
    if (!retry) {
      return {
        status: 'failed',
        message: 'There are no photos ready to retry.'
      };
    }
    return await this.uploadPhotoRetry(safePlanId, retry) ?? {
      status: 'failed',
      message: 'There are no photos ready to retry.'
    };
  }

  private isSessionGenerationCancelled(generation: number): boolean {
    return generation > 0 && generation <= this.cancelledThroughSessionGeneration;
  }

  private async reduceEvent(state: VoiceRealtimeState, event: VoiceRealtimeEvent): Promise<VoiceRealtimeState> {
    switch (event.type) {
      case 'session.started':
        return withProgressStep(state, 'Connected', { status: 'processing' });
      case 'transcript.delta':
        return withProgressStep(state, 'Transcribing', { status: 'processing', partialTranscript: event.text });
      case 'transcript.final':
        return withProgressStep(state, 'Understanding request', { status: 'processing', partialTranscript: undefined, transcript: event.text, conversationPhase: 'understanding' });
      case 'agent.progress':
        return withProgressStep(state, event.message, { status: 'processing', conversationPhase: voiceConversationPhase(event.status) });
      case 'agent.diagnostic':
        return {
          ...state,
          debugEvents: this.options.diagnosticsEnabled
            ? [...state.debugEvents, safeAgentDiagnosticEvent(event)]
            : state.debugEvents
        };
      case 'tool.call.started':
      case 'tool.call.completed':
      case 'tool.call.failed':
        return {
          ...state,
          status: 'processing',
          debugEvents: this.options.diagnosticsEnabled
            ? [...state.debugEvents, safeDiagnosticEvent(event)]
            : state.debugEvents
        };
      case 'action.plan.proposed':
        return withProgressStep(state, 'Review needed', { status: 'review', actionPlan: safeActionPlanProposal(event.actionPlan) });
      case 'action.plan.approved':
        return withProgressStep(state, 'Applying change', {
          status: 'processing',
          actionPlan: state.actionPlan && state.actionPlan.planId === event.planId
            ? { ...state.actionPlan, status: 'approved' }
            : state.actionPlan,
          reviewDecisionPending: true
        });
      case 'action.plan.cancelled':
        return withProgressStep(state, 'Change cancelled', {
          status: 'completed',
          actionPlan: state.actionPlan && state.actionPlan.planId === event.planId
            ? { ...state.actionPlan, status: 'cancelled' }
            : state.actionPlan
        });
      case 'action.plan.executed':
        const photoAttachmentStatus = await this.attachApprovedPlanPhotos({
          ...event,
          type: 'action.plan.executed',
          status: 'executed'
        }, state.actionPlan);
        return withProgressStep(state, 'Change applied', {
          status: 'completed',
          actionPlan: state.actionPlan && state.actionPlan.planId === event.planId
            ? { ...state.actionPlan, status: 'executed' }
            : state.actionPlan,
          reviewDecisionPending: false,
          photoAttachmentStatus
        });
      case 'action.plan.failed':
        await this.player.stop();
        return withProgressStep(state, 'Change failed', {
          status: 'failed',
          actionPlan: state.actionPlan && state.actionPlan.planId === event.planId
            ? { ...state.actionPlan, status: 'failed' }
            : state.actionPlan,
          reviewDecisionPending: false,
          errorMessage: 'The approved change could not be applied safely.'
        });
      case 'assistant.response.started':
        return withProgressStep(state, 'Preparing response', { status: state.actionPlan ? 'review' : 'processing', conversationPhase: 'answering' });
      case 'assistant.response.completed':
        return withProgressStep(state, event.response.kind === 'clarification' ? 'Needs detail' : 'Preparing speech', {
          status: state.actionPlan ? 'review' : 'processing',
          spokenResponse: event.response.displayResponse,
          responseKind: event.response.kind,
          conversationPhase: 'answering'
        });
      case 'tts.audio.started':
        this.ttsMimeType = event.mimeType;
        return withProgressStep(state, 'Speaking', { status: state.actionPlan ? 'review' : 'speaking' });
      case 'tts.audio.chunk':
        await this.player.playChunk(event.audioBase64, this.ttsMimeType);
        return withProgressStep(state, 'Speaking', { status: state.actionPlan ? 'review' : 'speaking' });
      case 'tts.audio.completed':
        return withProgressStep(state, 'Speech complete', { status: state.actionPlan ? 'review' : 'speaking' });
      case 'session.completed':
        await this.player.stop();
        return state.actionPlan
          ? withProgressStep(state, 'Review needed', { status: 'review' })
          : withProgressStep(state, state.responseKind === 'clarification' ? 'Needs detail' : 'Done', {
              status: 'completed',
              clarificationFollowUpAvailable: state.responseKind === 'clarification'
                ? this.transport.canSendFollowUpAudio()
                : undefined
            });
      case 'session.cancelled':
        await this.player.stop();
        return withProgressStep(state, 'Cancelled', { status: 'cancelled', partialTranscript: undefined });
      case 'session.failed':
        await this.player.stop();
        return withProgressStep(state, voiceFailureProgressLabel(event.code), {
          status: 'failed',
          partialTranscript: undefined,
          failureCode: voiceFailureCode(event.code),
          errorMessage: voiceFailureMessage(event.code, event.message, this.options.diagnosticsEnabled === true)
        });
    }
  }

  private syncFinalClarificationFollowUpAvailability(
    states: VoiceRealtimeState[],
    onState: VoiceRealtimeStateHandler | undefined
  ): void {
    const current = states[states.length - 1];
    if (!current || current.status !== 'completed' || current.responseKind !== 'clarification') {
      return;
    }
    const next = {
      ...current,
      clarificationFollowUpAvailable: this.transport.canSendFollowUpAudio()
    };
    if (next.clarificationFollowUpAvailable === current.clarificationFollowUpAvailable) {
      return;
    }
    states.push(next);
    onState?.(next);
  }

  private async selectedInventoryContext() {
    const workspace = await this.inventories.getInventoryWorkspace();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('Inventory workspace must include at least one inventory.');
    }

    const tenant = workspace.tenants.find((item) => item.id === inventory.tenantId);
    if (!tenant) {
      throw new Error('Selected inventory must belong to a tenant.');
    }

    return {
      tenantId: inventory.tenantId,
      inventoryId: inventory.id,
      tenantName: tenant.name,
      inventoryName: inventory.name
    };
  }

  private async attachApprovedPlanPhotos(
    event: VoiceActionPlanExecutedEvent,
    reviewedPlan: VoiceActionPlanProposal | undefined
  ): Promise<VoicePhotoAttachmentStatus | undefined> {
    const drafts = this.pendingPhotoDraftsByPlanId.get(event.planId);
    this.pendingPhotoDraftsByPlanId.delete(event.planId);
    if (!drafts || Object.keys(drafts).length === 0) {
      return undefined;
    }
    this.pendingPhotoRetriesByPlanId.delete(event.planId);

    const reviewedPhotoCommands = new Map(
      (reviewedPlan?.commands ?? [])
        .filter((command) => command.id && isPhotoAttachableReviewedCommand(command))
        .map((command) => [command.id ?? '', command])
    );
    const context = this.currentContext ?? (await this.selectedInventoryContext());
    const uploadIntentsByPhotoKey = new Map(
      (event.attachmentUploadIntents ?? []).map((intent) => [photoIntentKey(intent.commandId, intent.photoIndex), intent])
    );
    const retry: VoiceActionPlanPhotoRetry = {
      tenantId: context.tenantId,
      inventoryId: context.inventoryId,
      commandAssetIds: {},
      photos: {},
      nonRetryableFailures: []
    };
    const nonRetryableFailures: string[] = [];

    for (const [commandId, photos] of Object.entries(drafts)) {
      const reviewedCommand = reviewedPhotoCommands.get(commandId);
      const commandResult = (event.commandResults ?? []).find((result) => result.commandId === commandId);
      if (!reviewedCommand || !commandResult || !commandResultMatchesReviewedCommand(commandResult, reviewedCommand)) {
        nonRetryableFailures.push(...photos.map(() => 'The server did not return an upload intent for this photo.'));
        continue;
      }
      const photosWithIntents: CreateInventoryAssetPhotoInput[] = [];
      const photosWithoutIntents: CreateInventoryAssetPhotoInput[] = [];
      photos.forEach((photo, index) => {
        const intent = uploadIntentsByPhotoKey.get(photoIntentKey(commandId, index));
        if (!intent || intent.assetId !== commandResult.assetId || intent.contentType !== photo.contentType || intent.fileName !== photo.fileName || intent.sizeBytes !== photo.sizeBytes) {
          photosWithoutIntents.push(photo);
          return;
        }
        photosWithIntents.push({ ...photo, directUpload: intent.directUpload, sizeBytes: intent.sizeBytes });
      });
      if (photosWithIntents.length > 0) {
        retry.commandAssetIds[commandId] = commandResult.assetId;
        retry.photos[commandId] = photosWithIntents;
      }
      if (photosWithoutIntents.length > 0) {
        nonRetryableFailures.push('The server did not return an upload intent for this photo.');
      }
    }

    return this.uploadPhotoRetry(event.planId, { ...retry, nonRetryableFailures });
  }

  private async uploadPhotoRetry(planId: string, retry: VoiceActionPlanPhotoRetry): Promise<VoicePhotoAttachmentStatus | undefined> {
    let attempted = retry.nonRetryableFailures.length;
    let failed = retry.nonRetryableFailures.length;
    const remaining: VoiceActionPlanPhotoRetry = {
      tenantId: retry.tenantId,
      inventoryId: retry.inventoryId,
      commandAssetIds: { ...retry.commandAssetIds },
      photos: {},
      nonRetryableFailures: []
    };
    const failureMessages: string[] = [...retry.nonRetryableFailures];
    for (const [commandId, photos] of Object.entries(retry.photos)) {
      const targetAssetId = retry.commandAssetIds[commandId];
      for (const photo of photos) {
        attempted += 1;
        if (!targetAssetId) {
          failed += 1;
          failureMessages.push('The server did not return an upload intent for this photo.');
          remaining.photos[commandId] = [...(remaining.photos[commandId] ?? []), photo];
          continue;
        }
        try {
          if (!this.inventories.addInventoryAssetPhoto) {
            throw new Error('Photo attachments are not available in this build.');
          }
          await this.inventories.addInventoryAssetPhoto({
            tenantId: retry.tenantId,
            inventoryId: retry.inventoryId,
            assetId: assetId(targetAssetId),
            ...photo
          });
        } catch (error) {
          failed += 1;
          failureMessages.push(safePhotoUploadFailureReason(error));
          remaining.photos[commandId] = [...(remaining.photos[commandId] ?? []), photo];
        }
      }
    }
    if (failed > 0 && hasRetryablePhotos(remaining)) {
      this.pendingPhotoRetriesByPlanId.set(planId, remaining);
    } else {
      this.pendingPhotoRetriesByPlanId.delete(planId);
    }
    if (attempted === 0) {
      return undefined;
    }
    if (failed === 0) {
      return {
        status: 'attached',
        message: `${attempted.toString()} ${attempted === 1 ? 'photo' : 'photos'} attached.`
      };
    }
    if (failed < attempted) {
      return {
        status: 'partial_failed',
        message: `${(attempted - failed).toString()} of ${attempted.toString()} photos attached.`,
        canRetry: hasRetryablePhotos(remaining)
      };
    }
    return {
      status: 'failed',
      message: photoUploadFailureMessage(failureMessages),
      canRetry: hasRetryablePhotos(remaining)
    };
  }
}

function boundedRecordingLevel(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.max(0, Math.min(1, value));
}

function voiceConversationPhase(status: string): VoiceConversationPhase | undefined {
  switch (status) {
    case 'understanding':
    case 'exploring':
    case 'planning':
    case 'reviewing':
    case 'answering':
    case 'recovering':
      return status;
    default:
      return undefined;
  }
}

type VoiceActionPlanPhotoRetry = {
  readonly tenantId: TenantId;
  readonly inventoryId: InventoryId;
  readonly commandAssetIds: Record<string, string>;
  readonly photos: VoiceActionPlanPhotoDrafts;
  readonly nonRetryableFailures: readonly string[];
};

const maxVisibleProgressSteps = 12;

function withProgressStep(
  state: VoiceRealtimeState,
  label: string,
  updates: Partial<VoiceRealtimeState>
): VoiceRealtimeState {
  const safeLabel = safeVisibleProgressText(label, 100) || 'Working';
  const currentSteps = state.progressSteps ?? [];
  const nextSteps = currentSteps.at(-1) === safeLabel
    ? currentSteps
    : [...currentSteps, safeLabel].slice(-maxVisibleProgressSteps);
  return {
    ...state,
    ...updates,
    progressLabel: safeLabel,
    progressSteps: nextSteps
  };
}

function safePhotoUploadFailureReason(error: unknown): string {
  if (error instanceof Error && error.message.trim().length > 0) {
    return safeBoundedText(error.message, 160);
  }
  return 'Photo upload failed.';
}

function photoUploadFailureMessage(reasons: readonly string[]): string {
  const firstReason = reasons.find((reason) => reason.trim().length > 0);
  if (!firstReason) {
    return 'The change was applied, but photos could not be attached.';
  }
  return `The change was applied, but photos could not be attached: ${firstReason}`;
}

function photoApprovalRequests(drafts: VoiceActionPlanPhotoDrafts): readonly VoiceActionPlanPhotoApprovalRequest[] {
  const requests: VoiceActionPlanPhotoApprovalRequest[] = [];
  for (const [commandId, photos] of Object.entries(drafts)) {
    photos.forEach((photo, index) => {
      if (photo.sizeBytes && photo.sizeBytes > 0) {
        requests.push({
          commandId,
          photoIndex: index,
          fileName: photo.fileName,
          contentType: photo.contentType,
          sizeBytes: photo.sizeBytes
        });
      }
    });
  }
  return requests;
}

function validatePhotoApprovalMetadata(drafts: VoiceActionPlanPhotoDrafts): void {
  for (const photos of Object.values(drafts)) {
    for (const photo of photos) {
      if (!photo.sizeBytes || photo.sizeBytes <= 0) {
        throw new Error('Photo size is required before approving this change.');
      }
    }
  }
}

function photoIntentKey(commandId: string, photoIndex: number): string {
  return `${commandId}:${photoIndex.toString()}`;
}

function hasRetryablePhotos(retry: VoiceActionPlanPhotoRetry): boolean {
  return Object.values(retry.photos).some((photos) => photos.length > 0);
}

function safeDiagnosticEvent(event: Extract<VoiceRealtimeEvent, { readonly type: 'tool.call.started' | 'tool.call.completed' | 'tool.call.failed' }>): VoiceSafeDiagnosticEvent {
  return {
    label: safeDiagnosticLabel(event.toolLabel),
    status: safeDiagnosticStatus(event.status ?? event.code),
    detail: event.detail ? safeBoundedDiagnosticDetail(event.detail, 4000) : undefined
  };
}

function safeAgentDiagnosticEvent(event: Extract<VoiceRealtimeEvent, { readonly type: 'agent.diagnostic' }>): VoiceSafeDiagnosticEvent {
  return {
    label: safeVisibleProgressText(event.message, 120) || 'Agent diagnostic',
    status: 'Details',
    detail: event.detail ? safeBoundedDiagnosticDetail(event.detail, 4000) : undefined
  };
}

function safeDiagnosticLabel(toolLabel: string): string {
  const normalized = toolLabel.toLowerCase();

  if (normalized.includes('search')) {
    return 'Inventory search';
  }

  if (normalized.includes('asset') || normalized.includes('detail')) {
    return 'Asset lookup';
  }

  if (normalized.includes('location')) {
    return 'Location contents';
  }

  if (normalized.includes('list')) {
    return 'Inventory list';
  }

  return 'Inventory lookup';
}

function safeDiagnosticStatus(status: string | undefined): string {
  switch (status) {
    case 'completed':
      return 'Completed';
    case 'failed':
    case 'invalid_request':
    case 'unauthorized':
    case 'forbidden':
      return 'Failed safely';
    case 'needs_more_context':
      return 'Needs more context';
    case 'no_visible_match':
      return 'No visible match';
    case 'looking_up_item':
    case 'searching':
    case 'checking_location':
      return 'Looking';
    default:
      return 'Updated';
  }
}

function safeActionPlanProposal(proposal: VoiceActionPlanProposal): VoiceActionPlanProposal {
  const terminalStatuses: readonly VoiceActionPlanStatus[] = ['approved', 'cancelled', 'executed', 'failed'];
  const safeStatus: VoiceActionPlanStatus = terminalStatuses.includes(proposal.status)
    ? proposal.status
    : 'proposed';
  return {
    planId: safeBoundedText(proposal.planId, 80),
    status: safeStatus,
    confirmationSummary: safeBoundedText(proposal.confirmationSummary, 180),
    commands: proposal.commands.map((command) => ({
      id: command.id ? safeBoundedText(command.id, 80) : undefined,
      kind: safeBoundedText(command.kind, 40),
      summary: safeBoundedText(command.summary, 180),
      operation: command.operation ? safeBoundedText(command.operation, 40) : undefined,
      title: command.title ? safeBoundedText(command.title, 120) : undefined,
      assetKind: command.assetKind ? safeBoundedText(command.assetKind, 40) : undefined,
      parentAssetId: command.parentAssetId ? safeBoundedText(command.parentAssetId, 80) : undefined,
      parentTitle: command.parentTitle ? safeBoundedText(command.parentTitle, 120) : undefined,
      parentKind: command.parentKind ? safeBoundedText(command.parentKind, 40) : undefined,
      parentCommandId: command.parentCommandId ? safeBoundedText(command.parentCommandId, 80) : undefined
    })),
    risks: proposal.risks.slice(0, 6).map((risk) => safeBoundedText(risk, 180)).filter(Boolean)
  };
}

function boundedPhotoDrafts(drafts: VoiceActionPlanPhotoDrafts): VoiceActionPlanPhotoDrafts {
  const bounded: VoiceActionPlanPhotoDrafts = {};
  for (const [commandId, photos] of Object.entries(drafts).slice(0, 10)) {
    const safeCommandId = safeBoundedText(commandId, 80);
    if (!safeCommandId || photos.length === 0) {
      continue;
    }
    bounded[safeCommandId] = photos.slice(0, 10).map((photo) => ({
      fileName: safeBoundedText(photo.fileName, 160) || 'voice-photo.jpg',
      contentType: photo.contentType,
      contentBase64: photo.contentBase64,
      uri: photo.uri ? safeBoundedText(photo.uri, 2048) : undefined,
      sizeBytes: photo.sizeBytes
    }));
  }
  return bounded;
}

function isPhotoAttachableReviewedCommand(command: VoiceActionPlanCommand): boolean {
  return command.operation === 'create' ||
    command.operation === 'move' ||
    command.kind === 'create_asset' ||
    command.kind === 'create_location' ||
    command.kind === 'move_asset';
}

function commandResultMatchesReviewedCommand(result: VoiceActionPlanCommandResult, command: VoiceActionPlanCommand): boolean {
  if (!operationMatchesReviewedCommand(result.operation, command)) {
    return false;
  }
  return assetKindMatchesReviewedCommand(result.assetKind, command);
}

function operationMatchesReviewedCommand(resultOperation: string, command: VoiceActionPlanCommand): boolean {
  const expectedOperation = command.operation || (command.kind === 'create_asset' || command.kind === 'create_location' ? 'create' : undefined);
  return !expectedOperation || resultOperation === expectedOperation;
}

function assetKindMatchesReviewedCommand(resultKind: string, command: VoiceActionPlanCommand): boolean {
  const expectedKind = command.assetKind || (command.kind === 'create_location' ? 'location' : undefined);
  if (!expectedKind) {
    return command.operation === 'create' || command.kind === 'create_asset';
  }
  return resultKind === expectedKind;
}

function voiceFailureCode(code: string): VoiceRealtimeFailureCode {
  switch (code) {
    case 'speech_to_text_failed':
    case 'language_inference_failed':
    case 'text_to_speech_failed':
    case 'clarification_turn_limit':
      return code;
    default:
      return 'voice_failed';
  }
}

function voiceFailureMessage(code: string, fallback: string, diagnosticsEnabled: boolean): string {
  switch (code) {
    case 'speech_to_text_failed':
      return 'Speech-to-text provider failed. Check Voice providers and try again.';
    case 'language_inference_failed':
      return diagnosticsEnabled
        ? 'Language model stopped while continuing this request. Check diagnostics or Voice providers and try again.'
        : 'Language model stopped while continuing this request. Check Voice providers and try again.';
    case 'text_to_speech_failed':
      return 'Speech output failed after Stuff Stash prepared the answer. Check Voice providers and try again.';
    case 'clarification_turn_limit':
      return 'That thread needs a fresh voice request. Start again with the missing detail included.';
    default:
      return fallback;
  }
}

function voiceFailureProgressLabel(code: string): string {
  switch (code) {
    case 'clarification_turn_limit':
      return 'Voice needs a fresh start';
    case 'speech_to_text_failed':
      return 'Speech input failed';
    case 'language_inference_failed':
      return 'Agent brain failed';
    case 'text_to_speech_failed':
      return 'Speech output failed';
    default:
      return 'Voice failed';
  }
}

function safeBoundedText(value: string, maxLength: number): string {
  const normalized = value.replace(/\s+/g, ' ').trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return normalized.slice(0, maxLength).trim();
}

function safeVisibleProgressText(value: string, maxLength: number): string {
  const normalized = redactUnsafeVoiceText(value)
    .replace(/\s+/g, ' ')
    .trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return normalized.slice(0, maxLength).trim();
}

function safeBoundedDiagnosticDetail(value: string, maxLength: number): string {
  const normalized = redactUnsafeVoiceText(value)
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .replace(/[ \t]{2,}/g, ' ')
    .replace(/\n{4,}/g, '\n\n\n')
    .trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return `${normalized.slice(0, Math.max(0, maxLength - 3)).trimEnd()}...`;
}

function redactUnsafeVoiceText(value: string): string {
  return value
    .replace(/\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|secret|token)\s*[:=]\s*["']?[^"',\s}\n]+/gi, '$1: [redacted]')
    .replace(/bearer\s+[^"',\s}\]\)]+/gi, 'bearer [redacted]')
    .replace(/\b(raw prompt|stack trace|raw query|raw transcript|provider session id)\b/gi, '[redacted]');
}
