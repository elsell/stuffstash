import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController, type VoiceRealtimeEvent } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from './MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from './VoiceInteractionStateContext';

it('keeps earlier exchanges once on startup failure and never restores an accepted request for resubmission', async () => {
  const h = new MobileRenderHarness();
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' }), addInventoryAssetPhoto: async () => { throw new Error('Unused'); } };
  let failReadiness = false;
  let failAfterAcceptance = false;
  let value!: ReturnType<typeof useVoiceInteractionState>;
  const transport = {
    run: async (_input: unknown, receive: (event: VoiceRealtimeEvent) => Promise<void>) => {
      await receive({ seq: 1, type: 'transcript.final', text: 'Where is my drill?' });
      if (failAfterAcceptance) throw new Error('Connection interrupted');
      await receive({ seq: 2, type: 'session.completed' });
    },
    canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => undefined,
    approveActionPlan: async () => undefined, cancelActionPlan: async () => undefined
  };
  const controller = new RealtimeVoiceSessionController(context,
    { start: async () => undefined, stop: async () => ({ mimeType: 'audio/mp4', sampleRate: 44100, channels: 1, chunksBase64: [] }), cancel: async () => undefined, recordingLevel: () => 0 },
    transport, { playChunk: async () => undefined, stop: async () => undefined },
    { readinessChecker: { assertReady: async () => { if (failReadiness) throw new Error('Not ready'); } } });
  function Surface() { value = useVoiceInteractionState(); return null; }
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 10)));
  try {
    await h.render(<MobileServerStateProvider client={createMobileQueryClient()} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}><VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}><Surface /></VoiceInteractionStateProvider></MobileServerStateProvider>);
    await settle(); await settle();
    await h.run(() => value.setComposerText('Where is my drill?'));
    await h.run(() => value.sendText());
    failReadiness = true;
    await h.run(() => value.startRealtime());
    await h.run(() => value.startRealtime());
    expect(value.history).toHaveLength(1);
    expect(value.state.status === 'ready' && value.state.realtime?.transcript).toBeUndefined();
    failReadiness = false; failAfterAcceptance = true;
    await h.run(() => value.setComposerText('Where is my drill?'));
    await h.run(() => value.sendText());
    expect(value.composerText).toBe('');
    expect(value.history).toHaveLength(1);
  } finally { await h.unmount(); }
});
