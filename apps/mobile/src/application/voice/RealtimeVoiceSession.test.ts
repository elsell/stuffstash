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
      { type: 'session.started', sessionId: 'session-1' },
      { type: 'transcript.final', text: 'Where are my tools?' },
      { type: 'tool.call.started', toolCallId: 'tool-1', toolLabel: 'Search inventory', status: 'searching' },
      { type: 'tool.call.completed', toolCallId: 'tool-1', toolLabel: 'Search inventory', status: 'completed' },
      {
        type: 'assistant.response.completed',
        response: {
          spokenResponse: 'Your tools are in Garage.',
          displayResponse: 'Your tools are in Garage.',
          kind: 'answer'
        }
      },
      { type: 'tts.audio.started', mimeType: 'audio/mpeg' },
      { type: 'tts.audio.chunk', audioBase64: 'YXVkaW8tMQ==', chunkId: 'tts-1' },
      { type: 'tts.audio.completed' },
      { type: 'session.completed' }
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
  });
});

class FakeRecorder implements VoiceAudioRecorder {
  started = false;

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
}

class FakePlayer implements VoiceAudioPlayer {
  readonly played: Array<{ readonly audioBase64: string; readonly mimeType: string }> = [];

  async playChunk(audioBase64: string, mimeType: string): Promise<void> {
    this.played.push({ audioBase64, mimeType });
  }

  async stop(): Promise<void> {}
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
