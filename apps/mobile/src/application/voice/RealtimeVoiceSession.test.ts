import { describe, expect, it } from 'vitest';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import {
  RecordedVoiceAudio,
  RealtimeVoiceSessionController,
  RealtimeVoiceTransport,
  VoiceAudioPlayer,
  VoiceAudioRecorder,
  VoiceRealtimeEvent
} from './RealtimeVoiceSession';
import {
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { AssetSummary } from '../../domain/assets/AssetSummary';
import { LocationSummary } from '../../domain/locations/LocationSummary';

describe('RealtimeVoiceSessionController', () => {
  it('records the selected inventory context, streams safe progress, and plays TTS chunks', async () => {
    const recorder = new FakeRecorder();
    const transport = new FakeTransport([
      { type: 'session.started', seq: 1, sessionId: 'session-1' },
      { type: 'transcript.final', seq: 2, sessionId: 'session-1', text: 'Where are my tools?' },
      { type: 'tool.call.started', seq: 3, sessionId: 'session-1', toolCallId: 'tool-1', toolLabel: 'Search inventory', status: 'searching' },
      { type: 'tool.call.completed', seq: 4, sessionId: 'session-1', toolCallId: 'tool-1', toolLabel: 'Search inventory', status: 'completed' },
      {
        type: 'assistant.response.completed',
        seq: 5,
        sessionId: 'session-1',
        response: {
          spokenResponse: 'Your tools are in Garage.',
          displayResponse: 'Your tools are in Garage.',
          kind: 'answer'
        }
      },
      { type: 'tts.audio.started', seq: 6, sessionId: 'session-1', mimeType: 'audio/mpeg' },
      { type: 'tts.audio.chunk', seq: 7, sessionId: 'session-1', audioBase64: 'YXVkaW8tMQ==', chunkId: 'tts-1' },
      { type: 'tts.audio.completed', seq: 8, sessionId: 'session-1' },
      { type: 'session.completed', seq: 9, sessionId: 'session-1' }
    ]);
    const player = new FakePlayer();
    const controller = new RealtimeVoiceSessionController(
	  new FakeInventoryRepository(),
	  recorder,
	  transport,
	  player,
	  { diagnosticsEnabled: true }
	);

    const listening = await controller.start();
    expect(listening).toMatchObject({
      status: 'listening',
      tenantName: 'Home tenant',
      inventoryName: 'Home'
    });
    expect(recorder.started).toBe(true);

    const states = await controller.stop();

    expect(transport.lastInput).toMatchObject({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      inputAudio: {
        mimeType: 'audio/mp4',
        sampleRate: 44100,
        channels: 1
      },
      audioChunksBase64: ['ZmFrZS1hdWRpbw==']
    });
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      transcript: 'Where are my tools?',
      spokenResponse: 'Your tools are in Garage.'
    });
    expect(player.played).toEqual([{ audioBase64: 'YXVkaW8tMQ==', mimeType: 'audio/mpeg' }]);
    expect(player.stops).toBe(2);
  });

  it('notifies the mobile state layer as realtime session events arrive', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        { type: 'session.started', seq: 1, sessionId: 'session-1' },
        { type: 'transcript.delta', seq: 2, sessionId: 'session-1', text: 'Where is' },
        { type: 'transcript.final', seq: 3, sessionId: 'session-1', text: 'Where is the drill?' },
        { type: 'agent.progress', seq: 4, sessionId: 'session-1', status: 'searching', message: 'Searching visible inventory' },
        {
          type: 'assistant.response.completed',
          seq: 5,
          sessionId: 'session-1',
          response: {
            spokenResponse: 'The drill is in the garage.',
            displayResponse: 'The drill is in the garage.',
            kind: 'answer'
          }
        },
        { type: 'session.completed', seq: 6, sessionId: 'session-1' }
      ]),
      new FakePlayer()
    );
    const observed: string[] = [];

    await controller.start();
    const states = await controller.stop((state) => {
      observed.push(`${state.status}:${state.progressLabel ?? ''}:${state.transcript ?? state.partialTranscript ?? ''}:${state.spokenResponse ?? ''}`);
    });

    expect(observed).toEqual([
      'processing:Sending audio::',
      'processing:Connected::',
      'processing:Transcribing:Where is:',
      'processing:Thinking:Where is the drill?:',
      'processing:Searching visible inventory:Where is the drill?:',
      'processing:Preparing speech:Where is the drill?:The drill is in the garage.',
      'completed:Done:Where is the drill?:The drill is in the garage.'
    ]);
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      transcript: 'Where is the drill?',
      spokenResponse: 'The drill is in the garage.'
    });
  });

  it('enters review when the API proposes an action plan', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        { type: 'session.started', seq: 1, sessionId: 'session-1' },
        { type: 'transcript.final', seq: 2, sessionId: 'session-1', text: 'Add a water bottle' },
        {
          type: 'action.plan.proposed',
          seq: 3,
          sessionId: 'session-1',
          actionPlan: {
            planId: 'plan-1',
            status: 'proposed',
            confirmationSummary: 'Create item water bottle?',
            commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
            risks: ['Adds a new item to this inventory.']
          }
        },
        { type: 'assistant.response.started', seq: 4, sessionId: 'session-1', responseId: 'response-1' },
        {
          type: 'assistant.response.completed',
          seq: 5,
          sessionId: 'session-1',
          response: {
            spokenResponse: 'I prepared that change for review.',
            displayResponse: 'I prepared that change for review.',
            kind: 'clarification'
          }
        },
        { type: 'session.completed', seq: 6, sessionId: 'session-1' }
      ]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();
    const review = states.find((state) => state.status === 'review');

    expect(review).toMatchObject({
      status: 'review',
      progressLabel: 'Review needed',
      actionPlan: {
        planId: 'plan-1',
        status: 'proposed',
        confirmationSummary: 'Create item water bottle?',
        commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
        risks: ['Adds a new item to this inventory.']
      }
    });
    expect(states.at(-1)).toMatchObject({
      status: 'review',
      progressLabel: 'Review needed',
      actionPlan: {
        planId: 'plan-1',
        status: 'proposed'
      }
    });
  });

  it('approves a proposed action plan through the active realtime transport', async () => {
    const transport = new ReviewDecisionTransport();
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );
    const observed: string[] = [];

    await controller.start();
    const stop = controller.stop((state) => {
      observed.push(`${state.status}:${state.progressLabel ?? ''}:${state.actionPlan?.status ?? ''}`);
    });
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1');
    const states = await stop;

    expect(transport.approvedPlanIds).toEqual(['plan-1']);
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      progressLabel: 'Change applied',
      actionPlan: {
        planId: 'plan-1',
        status: 'executed'
      }
    });
    expect(observed).toContain('review:Review needed:proposed');
    expect(observed).toContain('processing:Applying change:approved');
    expect(observed.at(-1)).toBe('completed:Change applied:executed');
  });

  it('cancels a proposed action plan through the active realtime transport', async () => {
    const transport = new ReviewDecisionTransport();
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.cancelActionPlan('plan-1');
    const states = await stop;

    expect(transport.cancelledPlanIds).toEqual(['plan-1']);
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      progressLabel: 'Change cancelled',
      actionPlan: {
        planId: 'plan-1',
        status: 'cancelled'
      }
    });
  });

  it('shows a safe failure when approved action plan execution fails', async () => {
    const transport = new FailedReviewDecisionTransport();
    const player = new FakePlayer();
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      transport,
      player
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1');
    const states = await stop;

    expect(states.at(-1)).toMatchObject({
      status: 'failed',
      progressLabel: 'Change failed',
      errorMessage: 'The approved change could not be applied safely.',
      actionPlan: {
        planId: 'plan-1',
        status: 'failed'
      }
    });
    expect(player.stops).toBeGreaterThan(0);
  });

  it('returns the server failure state when the realtime session fails safely', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([{ type: 'session.failed', seq: 1, code: 'invalid_request', message: 'Voice is not configured.' }]),
      new FakePlayer()
    );
    const observed: string[] = [];

    await controller.start();
    const states = await controller.stop((state) => {
      observed.push(`${state.status}:${state.progressLabel ?? ''}:${state.errorMessage ?? ''}`);
    });

    expect(states.at(-1)).toMatchObject({
      status: 'failed',
      errorMessage: 'Voice is not configured.',
      progressLabel: 'Voice failed'
    });
    expect(observed).toEqual([
      'processing:Sending audio:',
      'failed:Voice failed:Voice is not configured.'
    ]);
  });

  it('clears partial transcripts when the realtime session fails before a final transcript', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        { type: 'transcript.delta', seq: 1, sessionId: 'session-1', text: 'Where is' },
        { type: 'session.failed', seq: 2, sessionId: 'session-1', code: 'invalid_request', message: 'Voice is not configured.' }
      ]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)).toMatchObject({
      status: 'failed',
      errorMessage: 'Voice is not configured.',
      progressLabel: 'Voice failed'
    });
    expect(states.at(-1)?.partialTranscript).toBeUndefined();
    expect(states.at(-1)?.transcript).toBeUndefined();
  });

  it('clears partial transcripts when the realtime session is cancelled before a final transcript', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        { type: 'transcript.delta', seq: 1, sessionId: 'session-1', text: 'Where is' },
        { type: 'session.cancelled', seq: 2, sessionId: 'session-1' }
      ]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)).toMatchObject({
      status: 'cancelled',
      progressLabel: 'Cancelled'
    });
    expect(states.at(-1)?.partialTranscript).toBeUndefined();
    expect(states.at(-1)?.transcript).toBeUndefined();
  });

  it('maps provider stage failures to safe actionable mobile state', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([{ type: 'session.failed', seq: 1, code: 'speech_to_text_failed', message: 'The voice session failed safely.' }]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)).toMatchObject({
      status: 'failed',
      failureCode: 'speech_to_text_failed',
      errorMessage: 'Speech-to-text provider failed. Check Voice providers and try again.',
      progressLabel: 'Voice failed'
    });
  });

  it('cancels active recording without opening the realtime transport', async () => {
    const recorder = new FakeRecorder();
    const transport = new FakeTransport([]);
    const player = new FakePlayer();
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      recorder,
      transport,
      player
    );

    await controller.start();
    const cancelled = await controller.cancel();

    expect(cancelled).toMatchObject({
      status: 'cancelled',
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Cancelled'
    });
    expect(recorder.cancelled).toBe(true);
    expect(transport.lastInput).toBeUndefined();
    expect(player.stops).toBe(2);
  });

  it('does not open the realtime transport when cancellation races with recorder stop', async () => {
    const recorder = new DelayedStopRecorder();
    const transport = new FakeTransport([]);
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      recorder,
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await recorder.stopStarted;

    await controller.cancel();
    recorder.finishStop();

    await expect(stop).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(transport.lastInput).toBeUndefined();
  });

  it('does not let a new start revive a cancelled delayed stop', async () => {
    const recorder = new DelayedStopRecorder();
    const transport = new FakeTransport([]);
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      recorder,
      transport,
      new FakePlayer()
    );

    await controller.start();
    const oldStop = controller.stop();
    await recorder.stopStarted;

    await controller.cancel();
    await controller.start();
    recorder.finishStop();

    await expect(oldStop).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(transport.lastInput).toBeUndefined();
  });

  it('stops applying transport events after cancellation is requested', async () => {
    const transport = new DelayedEventTransport();
    const player = new FakePlayer();
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      transport,
      player
    );
    const observed: string[] = [];

    await controller.start();
    const stop = controller.stop((state) => {
      observed.push(state.status);
    });
    await transport.started;
    await controller.cancel();
    transport.emit({
      type: 'tts.audio.chunk',
      seq: 1,
      sessionId: 'session-1',
      chunkId: 'late',
      audioBase64: 'bGF0ZQ=='
    });

    await expect(stop).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(observed).toEqual(['processing']);
    expect(player.played).toEqual([]);
  });

  it('cancels an in-flight realtime transport and treats the stop as cancelled', async () => {
    const transport = new CancellableTransport();
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.started;

    const cancelled = await controller.cancel();

    await expect(stop).rejects.toMatchObject({ code: 'voice_cancelled' });
    expect(cancelled).toMatchObject({
      status: 'cancelled',
      progressLabel: 'Cancelled'
    });
    expect(transport.cancelled).toBe(true);
  });

  it('does not start recording when provider profiles are not ready', async () => {
    const recorder = new FakeRecorder();
    const transport = new FakeTransport([]);
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      recorder,
      transport,
      new FakePlayer(),
      {
        readinessChecker: {
          assertReady: async () => {
            throw new Error('Voice provider profiles are not ready: text_to_speech.');
          }
        }
      }
    );

    await expect(controller.start()).rejects.toThrow(
      'Voice provider profiles are not ready: text_to_speech.'
    );
    await expect(controller.stop()).rejects.toThrow('Voice recording has not started.');
    expect(recorder.started).toBe(false);
    expect(transport.lastInput).toBeUndefined();
  });

  it('sanitizes diagnostic tool events before they reach mobile UI state', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        {
          type: 'tool.call.started',
          seq: 1,
          sessionId: 'session-1',
          toolCallId: 'tool-1',
          toolLabel: 'raw prompt bearer secret stack trace',
          status: 'raw query text'
        },
        { type: 'session.completed', seq: 2, sessionId: 'session-1' }
      ]),
      new FakePlayer(),
      { diagnosticsEnabled: true }
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)?.debugEvents).toEqual([
      { label: 'Inventory lookup', status: 'Updated' }
    ]);
  });
});

class FakeRecorder implements VoiceAudioRecorder {
  started = false;
  cancelled = false;

  async start(): Promise<void> {
    this.started = true;
  }

  async stop(): Promise<RecordedVoiceAudio> {
    return {
      mimeType: 'audio/mp4',
      sampleRate: 44100,
      channels: 1,
      chunksBase64: ['ZmFrZS1hdWRpbw==']
    };
  }

  async cancel(): Promise<void> {
    this.cancelled = true;
    this.started = false;
  }
}

class FakeTransport implements RealtimeVoiceTransport {
  lastInput: unknown;

  constructor(private readonly events: readonly VoiceRealtimeEvent[]) {}

  async run(input: Parameters<RealtimeVoiceTransport['run']>[0], onEvent: (event: VoiceRealtimeEvent) => Promise<void>): Promise<void> {
    this.lastInput = input;
    for (const event of this.events) {
      await onEvent(event);
    }
  }

  async approveActionPlan(_planId: string): Promise<void> {}

  async cancelActionPlan(_planId: string): Promise<void> {}
}

class ReviewDecisionTransport implements RealtimeVoiceTransport {
  readonly approvedPlanIds: string[] = [];
  readonly cancelledPlanIds: string[] = [];
  protected onEvent: ((event: VoiceRealtimeEvent) => Promise<void>) | undefined;
  private reviewReadyResolve: (() => void) | undefined;
  protected finishResolve: (() => void) | undefined;
  readonly reviewReady = new Promise<void>((resolve) => {
    this.reviewReadyResolve = resolve;
  });

  async run(
    _input: Parameters<RealtimeVoiceTransport['run']>[0],
    onEvent: (event: VoiceRealtimeEvent) => Promise<void>
  ): Promise<void> {
    this.onEvent = onEvent;
    await onEvent({ type: 'session.started', seq: 1, sessionId: 'session-1' });
    await onEvent({
      type: 'action.plan.proposed',
      seq: 2,
      sessionId: 'session-1',
      actionPlan: {
        planId: 'plan-1',
        status: 'proposed',
        confirmationSummary: 'Create item water bottle?',
        commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
        risks: ['Adds a new item to this inventory.']
      }
    });
    await onEvent({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    this.reviewReadyResolve?.();
    await new Promise<void>((resolve) => {
      this.finishResolve = resolve;
    });
  }

  async approveActionPlan(planId: string): Promise<void> {
    this.approvedPlanIds.push(planId);
    await this.onEvent?.({
      type: 'action.plan.approved',
      seq: 4,
      sessionId: 'session-1',
      planId,
      status: 'approved'
    });
    await this.onEvent?.({
      type: 'action.plan.executed',
      seq: 5,
      sessionId: 'session-1',
      planId,
      status: 'executed',
      message: 'The approved change was applied.'
    });
    this.finishResolve?.();
  }

  async cancelActionPlan(planId: string): Promise<void> {
    this.cancelledPlanIds.push(planId);
    await this.onEvent?.({
      type: 'action.plan.cancelled',
      seq: 4,
      sessionId: 'session-1',
      planId,
      status: 'cancelled'
    });
    this.finishResolve?.();
  }

  protected async emitReviewEvent(event: VoiceRealtimeEvent): Promise<void> {
    await this.onEvent?.(event);
  }

  protected finish(): void {
    this.finishResolve?.();
  }
}

class FailedReviewDecisionTransport extends ReviewDecisionTransport {
  override async approveActionPlan(planId: string): Promise<void> {
    this.approvedPlanIds.push(planId);
    await this.emitReviewEvent({
      type: 'action.plan.approved',
      seq: 4,
      sessionId: 'session-1',
      planId,
      status: 'approved'
    });
    await this.emitReviewEvent({
      type: 'action.plan.failed',
      seq: 5,
      sessionId: 'session-1',
      planId,
      status: 'failed',
      message: 'The approved change could not be applied safely.'
    });
    this.finish();
  }
}

class DelayedStopRecorder extends FakeRecorder {
  private stopStartedResolve: (() => void) | undefined;
  private finishStopResolve: (() => void) | undefined;
  readonly stopStarted = new Promise<void>((resolve) => {
    this.stopStartedResolve = resolve;
  });

  async stop(): Promise<RecordedVoiceAudio> {
    this.stopStartedResolve?.();
    await new Promise<void>((resolve) => {
      this.finishStopResolve = resolve;
    });
    return super.stop();
  }

  finishStop(): void {
    this.finishStopResolve?.();
  }
}

class DelayedEventTransport implements RealtimeVoiceTransport {
  private startedResolve: (() => void) | undefined;
  private emitEvent: ((event: VoiceRealtimeEvent) => void) | undefined;
  readonly started = new Promise<void>((resolve) => {
    this.startedResolve = resolve;
  });

  async run(
    _input: Parameters<RealtimeVoiceTransport['run']>[0],
    onEvent: (event: VoiceRealtimeEvent) => Promise<void>,
    options?: Parameters<RealtimeVoiceTransport['run']>[2]
  ): Promise<void> {
    this.startedResolve?.();
    await new Promise<void>((resolve, reject) => {
      options?.signal?.addEventListener('abort', () => {
        reject(Object.assign(new Error('Voice session cancelled.'), { code: 'voice_cancelled' }));
      });
      this.emitEvent = (event) => {
        onEvent(event).then(resolve, reject);
      };
    });
  }

  emit(event: VoiceRealtimeEvent): void {
    this.emitEvent?.(event);
  }

  async approveActionPlan(_planId: string): Promise<void> {}

  async cancelActionPlan(_planId: string): Promise<void> {}
}

class CancellableTransport implements RealtimeVoiceTransport {
  cancelled = false;
  private startedResolve: (() => void) | undefined;
  readonly started = new Promise<void>((resolve) => {
    this.startedResolve = resolve;
  });

  async run(
    _input: Parameters<RealtimeVoiceTransport['run']>[0],
    _onEvent: (event: VoiceRealtimeEvent) => Promise<void>,
    options?: Parameters<RealtimeVoiceTransport['run']>[2]
  ): Promise<void> {
    this.startedResolve?.();
    await new Promise<void>((resolve, reject) => {
      options?.signal?.addEventListener('abort', () => {
        this.cancelled = true;
        reject(Object.assign(new Error('Voice session cancelled.'), { code: 'voice_cancelled' }));
      });
    });
  }

  async approveActionPlan(_planId: string): Promise<void> {}

  async cancelActionPlan(_planId: string): Promise<void> {}
}

class FakePlayer implements VoiceAudioPlayer {
  readonly played: Array<{ readonly audioBase64: string; readonly mimeType: string }> = [];
  stops = 0;

  async playChunk(audioBase64: string, mimeType: string): Promise<void> {
    this.played.push({ audioBase64, mimeType });
  }

  async stop(): Promise<void> {
    this.stops++;
  }
}

class FakeInventoryRepository implements InventorySummaryRepository {
  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return {
      tenants: [{ id: tenantId('tenant-home'), name: 'Home tenant' }],
      inventories: [{
        id: inventoryId('inventory-home'),
        tenantId: tenantId('tenant-home'),
        name: 'Home',
        role: 'editor',
        permissions: ['view'],
        description: 'Home inventory.',
        updatedAtLabel: 'Updated today',
        locationCount: 0,
        locations: [],
        assets: []
      }],
      defaultInventoryId: inventoryId('inventory-home')
    };
  }

  async getDefaultInventorySummary() {
    return (await this.getInventoryWorkspace()).inventories[0];
  }

  async selectInventory(): Promise<void> {}

  async createAsset(_input: CreateInventoryAssetInput): Promise<AssetSummary> {
    throw new Error('Not used.');
  }

  async addAssetPhoto(): Promise<void> {}
  async archiveAsset(): Promise<void> {}
  async restoreAsset(): Promise<void> {}
  async deleteAsset(): Promise<void> {}

  async browseAssets() {
    return { assets: [], hasMore: false };
  }

  async searchAssets(): Promise<readonly AssetSummary[]> {
    return [];
  }

  async searchLocations(): Promise<readonly LocationSummary[]> {
    return [];
  }
}
