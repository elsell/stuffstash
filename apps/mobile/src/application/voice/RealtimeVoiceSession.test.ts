import { describe, expect, it } from 'vitest';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import {
  RecordedVoiceAudio,
  RealtimeVoiceSessionController,
  RealtimeVoiceTransport,
  VoiceActionPlanAttachmentUploadIntent,
  VoiceActionPlanCommand,
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

  it('includes bounded recorder level while listening', async () => {
    const recorder = new FakeRecorder();
    recorder.level = 1.4;
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      recorder,
      new FakeTransport([]),
      new FakePlayer()
    );

    const listening = await controller.start();

    expect(listening.recordingLevel).toBe(1);

    recorder.level = Number.NaN;
    expect(controller.recordingLevel()).toBe(0);

    recorder.level = -0.4;
    expect(controller.recordingLevel()).toBe(0);
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
      'processing:Understanding request:Where is the drill?:',
      'processing:Searching visible inventory:Where is the drill?:',
      'processing:Preparing speech:Where is the drill?:The drill is in the garage.',
      'completed:Done:Where is the drill?:The drill is in the garage.'
    ]);
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      transcript: 'Where is the drill?',
      spokenResponse: 'The drill is in the garage.',
      progressSteps: [
        'Sending audio',
        'Connected',
        'Transcribing',
        'Understanding request',
        'Searching visible inventory',
        'Preparing speech',
        'Done'
      ]
    });
  });

  it('records and sends follow-up audio after a clarification completion', async () => {
    const recorder = new FakeRecorder();
    const transport = new FakeTransport([
      {
        type: 'assistant.response.completed',
        seq: 1,
        sessionId: 'session-1',
        response: {
          spokenResponse: 'Which item should I update?',
          displayResponse: 'Which item should I update?',
          kind: 'clarification'
        }
      },
      { type: 'session.completed', seq: 2, sessionId: 'session-1' }
    ]);
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      recorder,
      transport,
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      responseKind: 'clarification',
      progressLabel: 'Needs detail'
    });

    const followUp = await controller.startFollowUp();
    expect(followUp).toMatchObject({ status: 'listening', progressLabel: 'Listening' });
    const followUpStates = await controller.stopFollowUp();
    expect(transport.followUpAudio).toEqual([['ZmFrZS1hdWRpbw==']]);
    expect(followUpStates.at(-1)).toMatchObject({ status: 'completed' });
  });

  it('bounds progress steps and collapses adjacent duplicate labels', async () => {
    const events: VoiceRealtimeEvent[] = [
      { type: 'session.started', seq: 1, sessionId: 'session-1' },
      { type: 'agent.progress', seq: 2, sessionId: 'session-1', status: 'duplicate', message: 'Checking shelves' },
      { type: 'agent.progress', seq: 3, sessionId: 'session-1', status: 'duplicate', message: 'Checking shelves' },
      ...Array.from({ length: 13 }, (_, index) => ({
        type: 'agent.progress' as const,
        seq: index + 4,
        sessionId: 'session-1',
        status: 'step',
        message: `Safe progress ${index + 1}`
      })),
      { type: 'session.completed', seq: 17, sessionId: 'session-1' }
    ];
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport(events),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();
    const progressSteps = states.at(-1)?.progressSteps ?? [];

    expect(progressSteps).toHaveLength(12);
    expect(progressSteps).not.toContain('Checking shelves');
    expect(progressSteps.at(-1)).toBe('Done');
    for (let index = 1; index < progressSteps.length; index++) {
      expect(progressSteps[index]).not.toBe(progressSteps[index - 1]);
    }
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

  it('preserves action plan values that merely look like redaction terms', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        {
          type: 'action.plan.proposed',
          seq: 1,
          sessionId: 'session-1',
          actionPlan: {
            planId: 'plan-raw-prompt-stack-trace',
            status: 'proposed',
            confirmationSummary: 'Create Stack Trace book?',
            commands: [{
              id: 'cmd-raw-query-notes',
              kind: 'create_asset',
              summary: 'Create Raw Query notes',
              title: 'Stack Trace book'
            }],
            risks: ['Uses the title the user provided.']
          }
        }
      ]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)?.actionPlan).toMatchObject({
      planId: 'plan-raw-prompt-stack-trace',
      confirmationSummary: 'Create Stack Trace book?',
      commands: [{
        id: 'cmd-raw-query-notes',
        summary: 'Create Raw Query notes',
        title: 'Stack Trace book'
      }]
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

  it('attaches staged photos to executed command result assets after approval', async () => {
    const transport = new ReviewDecisionTransport({
      commandResults: [{
        commandId: 'cmd-water-bottle',
        assetId: 'asset-water-bottle',
        operation: 'create',
        assetKind: 'item'
      }],
      attachmentUploadIntents: [testUploadIntent('cmd-water-bottle', 'asset-water-bottle', 'water-bottle.jpg')]
    });
    const repository = new FakeInventoryRepository();
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1', {
      'cmd-water-bottle': [{
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'cGhvdG8=',
        sizeBytes: 5
      }]
    });
    const states = await stop;

    expect(repository.addedPhotos).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-water-bottle',
        fileName: 'water-bottle.jpg',
        uploadId: 'upload-cmd-water-bottle'
      }
    ]);
    expect(states.at(-1)).toMatchObject({
      status: 'completed',
      photoAttachmentStatus: {
        status: 'attached',
        message: '1 photo attached.'
      }
    });
  });

  it('attaches staged photos to reviewed move command results after approval', async () => {
    const transport = new ReviewDecisionTransport({
      commands: [{
        id: 'cmd-water-bottle',
        kind: 'move_asset',
        summary: 'Move Water bottle to Kitchen',
        operation: 'move',
        assetKind: 'item',
        title: 'Move Water bottle to Kitchen'
      }],
      commandResults: [{
        commandId: 'cmd-water-bottle',
        assetId: 'asset-water-bottle',
        operation: 'move',
        assetKind: 'item'
      }],
      attachmentUploadIntents: [testUploadIntent('cmd-water-bottle', 'asset-water-bottle', 'water-bottle.jpg')]
    });
    const repository = new FakeInventoryRepository();
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1', {
      'cmd-water-bottle': [{
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'cGhvdG8=',
        sizeBytes: 5
      }]
    });
    const states = await stop;

    expect(repository.addedPhotos).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-water-bottle',
        fileName: 'water-bottle.jpg',
        uploadId: 'upload-cmd-water-bottle'
      }
    ]);
    expect(states.at(-1)?.photoAttachmentStatus).toMatchObject({
      status: 'attached',
      message: '1 photo attached.'
    });
  });

  it('retains failed photo uploads so the user can retry after the plan applied', async () => {
    const transport = new ReviewDecisionTransport({
      commandResults: [{
        commandId: 'cmd-water-bottle',
        assetId: 'asset-water-bottle',
        operation: 'create',
        assetKind: 'item'
      }],
      attachmentUploadIntents: [testUploadIntent('cmd-water-bottle', 'asset-water-bottle', 'water-bottle.jpg')]
    });
    const repository = new FakeInventoryRepository();
    repository.failPhotoUploads = 1;
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1', {
      'cmd-water-bottle': [{
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'cGhvdG8=',
        sizeBytes: 5
      }]
    });
    const states = await stop;

    expect(states.at(-1)?.photoAttachmentStatus).toMatchObject({
      status: 'failed',
      canRetry: true
    });

    const retry = await controller.retryPhotoAttachments('plan-1');

    expect(retry).toEqual({
      status: 'attached',
      message: '1 photo attached.'
    });
    expect(repository.addedPhotos).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-water-bottle',
        fileName: 'water-bottle.jpg',
        uploadId: 'upload-cmd-water-bottle'
      }
    ]);
  });

  it('surfaces the safe photo upload failure reason when all staged photos fail', async () => {
    const transport = new ReviewDecisionTransport({
      commandResults: [{
        commandId: 'cmd-water-bottle',
        assetId: 'asset-water-bottle',
        operation: 'create',
        assetKind: 'item'
      }],
      attachmentUploadIntents: [testUploadIntent('cmd-water-bottle', 'asset-water-bottle', 'water-bottle.jpg', 123)]
    });
    const repository = new FakeInventoryRepository();
    repository.failPhotoUploads = 1;
    repository.photoUploadFailureMessage = 'Attachment content is not available for JSON upload fallback.';
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1', {
      'cmd-water-bottle': [{
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        uri: 'ph://photo-one',
        sizeBytes: 123
      }]
    });
    const states = await stop;

    expect(states.at(-1)?.photoAttachmentStatus).toEqual({
      status: 'failed',
      message: 'The change was applied, but photos could not be attached: Attachment content is not available for JSON upload fallback.',
      canRetry: true
    });
  });

  it('does not offer retry when the server omits the required upload intent', async () => {
    const transport = new ReviewDecisionTransport({
      commandResults: [{
        commandId: 'cmd-water-bottle',
        assetId: 'asset-water-bottle',
        operation: 'create',
        assetKind: 'item'
      }]
    });
    const repository = new FakeInventoryRepository();
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1', {
      'cmd-water-bottle': [{
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'cGhvdG8=',
        sizeBytes: 5
      }]
    });
    const states = await stop;

    expect(states.at(-1)?.photoAttachmentStatus).toEqual({
      status: 'failed',
      message: 'The change was applied, but photos could not be attached: The server did not return an upload intent for this photo.',
      canRetry: false
    });
    expect(repository.addedPhotos).toEqual([]);
    await expect(controller.retryPhotoAttachments('plan-1')).resolves.toEqual({
      status: 'failed',
      message: 'There are no photos ready to retry.'
    });
  });

  it('rejects staged photos without size metadata before approving a plan', async () => {
    const transport = new ReviewDecisionTransport();
    const repository = new FakeInventoryRepository();
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await expect(controller.approveActionPlan('plan-1', {
      'cmd-water-bottle': [{
        fileName: 'water-bottle.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'cGhvdG8='
      }]
    })).rejects.toThrow('Photo size is required before approving this change.');

    await transport.cancelActionPlan('plan-1');
    await stop;
    expect(transport.approvedPhotoRequests).toEqual([]);
    expect(repository.addedPhotos).toEqual([]);
  });

  it('attaches staged photos through server upload intents for dependent created items', async () => {
    const transport = new ReviewDecisionTransport({
      commands: [
        { id: 'cmd-room', kind: 'create_location', summary: 'Create Henry room', operation: 'create', assetKind: 'location', title: 'Henry room' },
        { id: 'cmd-closet', kind: 'create_asset', summary: 'Create closet', operation: 'create', assetKind: 'container', title: 'closet', parentCommandId: 'cmd-room' },
        { id: 'cmd-refills', kind: 'create_asset', summary: 'Create diaper genie refills', operation: 'create', assetKind: 'item', title: 'diaper genie refills', parentCommandId: 'cmd-closet' }
      ],
      commandResults: [
        { commandId: 'cmd-room', assetId: 'asset-room', operation: 'create', assetKind: 'location' },
        { commandId: 'cmd-closet', assetId: 'asset-closet', operation: 'create', assetKind: 'container' },
        { commandId: 'cmd-refills', assetId: 'asset-refills', operation: 'create', assetKind: 'item' }
      ],
      attachmentUploadIntents: [{
        commandId: 'cmd-refills',
        photoIndex: 0,
        assetId: 'asset-refills',
        fileName: 'refills.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 4,
        directUpload: {
          uploadId: 'upload-one',
          attachmentId: 'attachment-one',
          method: 'PUT',
          url: 'stuffstash-local://direct-uploads/upload-one',
          headers: { 'content-type': 'image/jpeg' },
          formFields: {},
          expiresAt: '2026-07-01T12:00:00Z'
        }
      }]
    });
    const repository = new FakeInventoryRepository();
    const controller = new RealtimeVoiceSessionController(
      repository,
      new FakeRecorder(),
      transport,
      new FakePlayer()
    );

    await controller.start();
    const stop = controller.stop();
    await transport.reviewReady;

    await controller.approveActionPlan('plan-1', {
      'cmd-refills': [{
        fileName: 'refills.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ==',
        uri: 'file:///refills.jpg',
        sizeBytes: 4
      }]
    });
    const states = await stop;

    expect(transport.approvedPhotoRequests).toEqual([{
      commandId: 'cmd-refills',
      photoIndex: 0,
      fileName: 'refills.jpg',
      contentType: 'image/jpeg',
      sizeBytes: 4
    }]);
    expect(repository.addedPhotos).toEqual([{
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-refills',
      fileName: 'refills.jpg',
      uploadId: 'upload-one'
    }]);
    expect(states.at(-1)?.photoAttachmentStatus).toMatchObject({
      status: 'attached',
      message: '1 photo attached.'
    });
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

  it('maps late language provider failures to continuation-specific copy', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([{ type: 'session.failed', seq: 1, code: 'language_inference_failed', message: 'The voice session failed safely.' }]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)).toMatchObject({
      status: 'failed',
      failureCode: 'language_inference_failed',
      errorMessage: 'Language model stopped while continuing this request. Check Voice providers and try again.',
      progressLabel: 'Voice failed'
    });
  });

  it('maps clarification turn limits to a fresh-request failure state', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([{ type: 'session.failed', seq: 1, code: 'clarification_turn_limit', message: 'The voice session needs a fresh start.' }]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)).toMatchObject({
      status: 'failed',
      failureCode: 'clarification_turn_limit',
      errorMessage: 'That thread needs a fresh voice request. Start again with the missing detail included.',
      progressLabel: 'Voice needs a fresh start'
    });
  });

  it('mentions diagnostics for language provider failures only when diagnostics are enabled', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([{ type: 'session.failed', seq: 1, code: 'language_inference_failed', message: 'The voice session failed safely.' }]),
      new FakePlayer(),
      { diagnosticsEnabled: true }
    );

    await controller.start();
    const states = await controller.stop();

    expect(states.at(-1)?.errorMessage).toBe('Language model stopped while continuing this request. Check diagnostics or Voice providers and try again.');
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
          status: 'raw query text',
          detail: 'name: search_authorized_assets\napiKey: should-not-leak\nquery: water bottle'
        },
        {
          type: 'agent.diagnostic',
          seq: 2,
          sessionId: 'session-1',
          message: 'Language prompt',
          detail: 'Transcript: move water bottle\nbearer abc123'
        },
        { type: 'session.completed', seq: 3, sessionId: 'session-1' }
      ]),
      new FakePlayer(),
      { diagnosticsEnabled: true }
    );
    const visibleProgressLabels: string[] = [];

    await controller.start();
    const states = await controller.stop((state) => {
      visibleProgressLabels.push(state.progressLabel ?? '');
    });

    expect(states.at(-1)?.debugEvents).toEqual([
      {
        label: 'Inventory lookup',
        status: 'Updated',
        detail: 'name: search_authorized_assets\napiKey: [redacted]\nquery: water bottle'
      },
      {
        label: 'Language prompt',
        status: 'Details',
        detail: 'Transcript: move water bottle\nbearer [redacted]'
      }
    ]);
    expect(visibleProgressLabels.join(' ')).not.toContain('bearer secret');
    expect(states.at(-1)?.progressSteps?.join(' ')).not.toContain('bearer secret');
  });

  it('redacts unsafe agent progress text before it reaches visible state', async () => {
    const controller = new RealtimeVoiceSessionController(
      new FakeInventoryRepository(),
      new FakeRecorder(),
      new FakeTransport([
        {
          type: 'agent.progress',
          seq: 1,
          sessionId: 'session-1',
          status: 'exploring',
          message: 'raw prompt bearer abc/def== stack trace provider session id: gemini-live-1'
        },
        {
          type: 'agent.progress',
          seq: 2,
          sessionId: 'session-1',
          status: 'answering',
          message: 'Authorization: tok+en/with~punctuation bearer eyJhbGciOi.test.sig=='
        },
        { type: 'session.completed', seq: 3, sessionId: 'session-1' }
      ]),
      new FakePlayer()
    );

    await controller.start();
    const states = await controller.stop();
    const visibleText = `${states.at(-1)?.progressLabel ?? ''} ${(states.at(-1)?.progressSteps ?? []).join(' ')}`;

    expect(visibleText).toContain('[redacted]');
    expect(visibleText).not.toContain('raw prompt');
    expect(visibleText).not.toContain('abc/def');
    expect(visibleText).not.toContain('tok+en');
    expect(visibleText).not.toContain('eyJhbGciOi');
    expect(visibleText).not.toContain('stack trace');
    expect(visibleText).not.toContain('gemini-live-1');
  });
});

class FakeRecorder implements VoiceAudioRecorder {
  started = false;
  cancelled = false;
  level = 0;

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

  recordingLevel(): number {
    return this.level;
  }
}

class FakeTransport implements RealtimeVoiceTransport {
  lastInput: unknown;
  followUpAudio: readonly string[][] = [];

  constructor(private readonly events: readonly VoiceRealtimeEvent[]) {}

  async run(input: Parameters<RealtimeVoiceTransport['run']>[0], onEvent: (event: VoiceRealtimeEvent) => Promise<void>): Promise<void> {
    this.lastInput = input;
    for (const event of this.events) {
      await onEvent(event);
    }
  }

  canSendFollowUpAudio(): boolean {
    return true;
  }

  async sendFollowUpAudio(audioChunksBase64: readonly string[], onEvent?: (event: VoiceRealtimeEvent) => Promise<void>): Promise<void> {
    this.followUpAudio = [...this.followUpAudio, [...audioChunksBase64]];
    for (const event of this.events) {
      await onEvent?.(event);
    }
  }

  async approveActionPlan(_planId: string): Promise<void> {}

  async cancelActionPlan(_planId: string): Promise<void> {}
}

class ReviewDecisionTransport implements RealtimeVoiceTransport {
  readonly approvedPlanIds: string[] = [];
  readonly approvedPhotoRequests: unknown[] = [];
  readonly cancelledPlanIds: string[] = [];
  protected onEvent: ((event: VoiceRealtimeEvent) => Promise<void>) | undefined;
  private reviewReadyResolve: (() => void) | undefined;
  protected finishResolve: (() => void) | undefined;
  readonly reviewReady = new Promise<void>((resolve) => {
    this.reviewReadyResolve = resolve;
  });

  constructor(private readonly options: {
    readonly commands?: readonly VoiceActionPlanCommand[];
    readonly commandResults?: readonly {
      readonly commandId: string;
      readonly assetId: string;
      readonly operation: string;
      readonly assetKind: string;
    }[];
    readonly attachmentUploadIntents?: readonly VoiceActionPlanAttachmentUploadIntent[];
  } = {}) {}

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
        commands: this.options.commands ?? [{ id: 'cmd-water-bottle', kind: 'create_asset', summary: 'Create item water bottle' }],
        risks: ['Adds a new item to this inventory.']
      }
    });
    await onEvent({ type: 'session.completed', seq: 3, sessionId: 'session-1' });
    this.reviewReadyResolve?.();
    await new Promise<void>((resolve) => {
      this.finishResolve = resolve;
    });
  }

  async approveActionPlan(planId: string, photos: readonly unknown[] = []): Promise<void> {
    this.approvedPlanIds.push(planId);
    this.approvedPhotoRequests.push(...photos);
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
      message: 'The approved change was applied.',
      ...(this.options.commandResults ? { commandResults: this.options.commandResults } : {}),
      ...(this.options.attachmentUploadIntents ? { attachmentUploadIntents: this.options.attachmentUploadIntents } : {})
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

  canSendFollowUpAudio(): boolean {
    return false;
  }

  async sendFollowUpAudio(): Promise<void> {
    throw new Error('Voice follow-up session is not active.');
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
      message: 'stack trace raw prompt bearer secret'
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

  canSendFollowUpAudio(): boolean {
    return false;
  }

  async sendFollowUpAudio(): Promise<void> {
    throw new Error('Voice follow-up session is not active.');
  }
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

  canSendFollowUpAudio(): boolean {
    return false;
  }

  async sendFollowUpAudio(): Promise<void> {
    throw new Error('Voice follow-up session is not active.');
  }
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

function testUploadIntent(commandId: string, assetIdValue: string, fileName: string, sizeBytes = 5): VoiceActionPlanAttachmentUploadIntent {
  return {
    commandId,
    photoIndex: 0,
    assetId: assetIdValue,
    fileName,
    contentType: 'image/jpeg',
    sizeBytes,
    directUpload: {
      uploadId: `upload-${commandId}`,
      attachmentId: `attachment-${commandId}`,
      method: 'PUT',
      url: `stuffstash-local://direct-uploads/upload-${commandId}`,
      headers: { 'content-type': 'image/jpeg' },
      formFields: {},
      expiresAt: '2026-07-01T12:00:00Z'
    }
  };
}

class FakeInventoryRepository implements InventorySummaryRepository {
  readonly addedPhotos: Array<{ readonly tenantId?: string; readonly inventoryId?: string; readonly assetId: string; readonly fileName: string; readonly uploadId?: string }> = [];
  failPhotoUploads = 0;
  photoUploadFailureMessage = 'Photo upload failed.';

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

  async addAssetPhoto(assetIdValue: string, input: { readonly fileName: string }): Promise<void> {
    this.addedPhotos.push({ assetId: assetIdValue, fileName: input.fileName });
  }
  async addInventoryAssetPhoto(input: { readonly tenantId: string; readonly inventoryId: string; readonly assetId: string; readonly fileName: string; readonly directUpload?: { readonly uploadId: string } }): Promise<void> {
    if (this.failPhotoUploads > 0) {
      this.failPhotoUploads -= 1;
      throw new Error(this.photoUploadFailureMessage);
    }
    this.addedPhotos.push({
      tenantId: input.tenantId,
      inventoryId: input.inventoryId,
      assetId: input.assetId,
      fileName: input.fileName,
      uploadId: input.directUpload?.uploadId
    });
  }
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
