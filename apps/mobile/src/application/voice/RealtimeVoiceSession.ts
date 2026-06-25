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
  run(input: RealtimeVoiceTransportInput, onEvent: (event: VoiceRealtimeEvent) => Promise<void>): Promise<void>;
}

export type VoiceRealtimeEvent =
  | { readonly type: 'session.started'; readonly sessionId: string }
  | { readonly type: 'session.failed'; readonly code: string; readonly message: string }
  | { readonly type: 'transcript.final'; readonly text: string }
  | { readonly type: 'agent.progress'; readonly status: string; readonly message: string }
  | {
      readonly type: 'tool.call.started' | 'tool.call.completed' | 'tool.call.failed';
      readonly toolCallId: string;
      readonly toolLabel: string;
      readonly status?: string;
      readonly code?: string;
      readonly message?: string;
    }
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
  | { readonly type: 'session.completed' };

export type VoiceRealtimeState = {
  readonly status: 'ready' | 'listening' | 'processing' | 'speaking' | 'completed' | 'failed';
  readonly tenantName: string;
  readonly inventoryName: string;
  readonly transcript?: string;
  readonly spokenResponse?: string;
  readonly progressLabel?: string;
  readonly debugEvents: readonly string[];
  readonly errorMessage?: string;
};

export class RealtimeVoiceSessionController {
  private currentContext: { readonly tenantId: string; readonly inventoryId: string; readonly tenantName: string; readonly inventoryName: string } | null = null;
  private ttsMimeType = 'audio/mpeg';

  constructor(
    private readonly inventories: InventorySummaryRepository,
    private readonly recorder: VoiceAudioRecorder,
    private readonly transport: RealtimeVoiceTransport,
    private readonly player: VoiceAudioPlayer
  ) {}

  async start(): Promise<VoiceRealtimeState> {
    const context = await this.selectedInventoryContext();
    this.currentContext = context;
    await this.player.stop();
    await this.recorder.start();
    return {
      status: 'listening',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Listening',
      debugEvents: []
    };
  }

  async stop(): Promise<readonly VoiceRealtimeState[]> {
    const context = this.currentContext ?? (await this.selectedInventoryContext());
    const recorded = await this.recorder.stop();
    const states: VoiceRealtimeState[] = [{
      status: 'processing',
      tenantName: context.tenantName,
      inventoryName: context.inventoryName,
      progressLabel: 'Sending audio',
      debugEvents: []
    }];

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
      const previous = states[states.length - 1];
      const next = await this.reduceEvent(previous, event);
      states.push(next);
    });

    return states;
  }

  private async reduceEvent(state: VoiceRealtimeState, event: VoiceRealtimeEvent): Promise<VoiceRealtimeState> {
    switch (event.type) {
      case 'session.started':
        return { ...state, status: 'processing', progressLabel: 'Connected' };
      case 'transcript.final':
        return { ...state, status: 'processing', transcript: event.text, progressLabel: 'Thinking' };
      case 'agent.progress':
        return { ...state, status: 'processing', progressLabel: event.message };
      case 'tool.call.started':
      case 'tool.call.completed':
      case 'tool.call.failed':
        return {
          ...state,
          status: 'processing',
          debugEvents: [...state.debugEvents, `${event.toolLabel}: ${event.status ?? event.code ?? 'updated'}`],
          progressLabel: event.toolLabel
        };
      case 'assistant.response.completed':
        return {
          ...state,
          status: 'processing',
          spokenResponse: event.response.displayResponse,
          progressLabel: 'Preparing speech'
        };
      case 'tts.audio.started':
        this.ttsMimeType = event.mimeType;
        return { ...state, status: 'speaking', progressLabel: 'Speaking' };
      case 'tts.audio.chunk':
        await this.player.playChunk(event.audioBase64, this.ttsMimeType);
        return { ...state, status: 'speaking', progressLabel: 'Speaking' };
      case 'tts.audio.completed':
        return { ...state, status: 'speaking', progressLabel: 'Speech complete' };
      case 'session.completed':
        return { ...state, status: 'completed', progressLabel: 'Done' };
      case 'session.failed':
        return { ...state, status: 'failed', errorMessage: event.message, progressLabel: 'Voice failed' };
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
