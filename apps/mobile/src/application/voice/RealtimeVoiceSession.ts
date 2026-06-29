import type { InventorySummaryRepository } from '../home/InventorySummaryRepository';

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
  approveActionPlan(planId: string): Promise<void>;
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
  | { readonly type: 'session.started'; readonly sessionId: string }
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
  | { readonly type: 'tts.audio.chunk'; readonly chunkId: string; readonly audioBase64: string }
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
  readonly progressLabel?: string;
  readonly debugEvents: readonly VoiceSafeDiagnosticEvent[];
  readonly failureCode?: VoiceRealtimeFailureCode;
  readonly errorMessage?: string;
};

export type VoiceRealtimeFailureCode =
  | 'provider_readiness'
  | 'speech_to_text_failed'
  | 'language_inference_failed'
  | 'text_to_speech_failed'
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
  private currentContext: { readonly tenantId: string; readonly inventoryId: string; readonly tenantName: string; readonly inventoryName: string } | null = null;
  private recordingStarted = false;
  private activeRunAbortController: AbortController | null = null;
  private activeSessionGeneration = 0;
  private cancelledThroughSessionGeneration = 0;
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
      debugEvents: []
    };
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

  async approveActionPlan(planId: string): Promise<void> {
    await this.transport.approveActionPlan(safeBoundedText(planId, 80));
  }

  async cancelActionPlan(planId: string): Promise<void> {
    await this.transport.cancelActionPlan(safeBoundedText(planId, 80));
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
        return withProgressStep(state, 'Thinking', { status: 'processing', partialTranscript: undefined, transcript: event.text });
      case 'agent.progress':
        return withProgressStep(state, event.message, { status: 'processing' });
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
        return withProgressStep(state, 'Change applied', {
          status: 'completed',
          actionPlan: state.actionPlan && state.actionPlan.planId === event.planId
            ? { ...state.actionPlan, status: 'executed' }
            : state.actionPlan,
          reviewDecisionPending: false
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
        return withProgressStep(state, 'Preparing response', { status: state.actionPlan ? 'review' : 'processing' });
      case 'assistant.response.completed':
        return withProgressStep(state, 'Preparing speech', { status: state.actionPlan ? 'review' : 'processing', spokenResponse: event.response.displayResponse });
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
          : withProgressStep(state, 'Done', { status: 'completed' });
      case 'session.cancelled':
        await this.player.stop();
        return withProgressStep(state, 'Cancelled', { status: 'cancelled', partialTranscript: undefined });
      case 'session.failed':
        await this.player.stop();
        return withProgressStep(state, 'Voice failed', {
          status: 'failed',
          partialTranscript: undefined,
          failureCode: voiceFailureCode(event.code),
          errorMessage: voiceFailureMessage(event.code, event.message)
        });
    }
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
}

const maxVisibleProgressSteps = 12;

function withProgressStep(
  state: VoiceRealtimeState,
  label: string,
  updates: Partial<VoiceRealtimeState>
): VoiceRealtimeState {
  const safeLabel = safeBoundedText(label, 100) || 'Working';
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

function safeDiagnosticEvent(event: Extract<VoiceRealtimeEvent, { readonly type: 'tool.call.started' | 'tool.call.completed' | 'tool.call.failed' }>): VoiceSafeDiagnosticEvent {
  return {
    label: safeDiagnosticLabel(event.toolLabel),
    status: safeDiagnosticStatus(event.status ?? event.code),
    detail: event.detail ? safeBoundedDiagnosticDetail(event.detail, 4000) : undefined
  };
}

function safeAgentDiagnosticEvent(event: Extract<VoiceRealtimeEvent, { readonly type: 'agent.diagnostic' }>): VoiceSafeDiagnosticEvent {
  return {
    label: safeBoundedText(event.message, 120) || 'Agent diagnostic',
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

function voiceFailureCode(code: string): VoiceRealtimeFailureCode {
  switch (code) {
    case 'speech_to_text_failed':
    case 'language_inference_failed':
    case 'text_to_speech_failed':
      return code;
    default:
      return 'voice_failed';
  }
}

function voiceFailureMessage(code: string, fallback: string): string {
  switch (code) {
    case 'speech_to_text_failed':
      return 'Speech-to-text provider failed. Check Voice providers and try again.';
    case 'language_inference_failed':
      return 'Language provider failed. Check Voice providers and try again.';
    case 'text_to_speech_failed':
      return 'Text-to-speech provider failed. Check Voice providers and try again.';
    default:
      return fallback;
  }
}

function safeBoundedText(value: string, maxLength: number): string {
  const normalized = value.replace(/\s+/g, ' ').trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return normalized.slice(0, maxLength).trim();
}

function safeBoundedDiagnosticDetail(value: string, maxLength: number): string {
  const normalized = value
    .replace(/\b(api[-_ ]?key|authorization|credential|password|provider[-_ ]?session[-_ ]?id|secret|token)\s*[:=]\s*["']?[^"',\s}\n]+/gi, '$1: [redacted]')
    .replace(/bearer\s+[a-z0-9._-]+/gi, 'bearer [redacted]')
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
