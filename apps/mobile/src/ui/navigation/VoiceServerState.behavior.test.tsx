import React from 'react';
import { Text, Pressable } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient, mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from './MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from './VoiceInteractionStateContext';

it('shares focused voice context and cancels recording on inventory replacement', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let selected = 'Garage'; let cancelled = 0;
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId(selected), tenantName: 'Home', inventoryName: selected }) };
  const controller = new RealtimeVoiceSessionController(context, { start: async () => undefined, stop: async () => ({ mimeType: 'audio/mp4', sampleRate: 44100, channels: 1, chunksBase64: [] }), cancel: async () => { cancelled++; }, recordingLevel: () => 0 }, { run: async () => undefined, canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => undefined, approveActionPlan: async () => undefined, cancelActionPlan: async () => undefined }, { playChunk: async () => undefined, stop: async () => undefined });
  function Surface() { const value = useVoiceInteractionState(); return <><Text>{value.state.status === 'ready' ? `${value.state.preview.inventoryName}:${value.state.stage}` : 'Loading'}</Text><Pressable accessibilityLabel="Start" onPress={value.startRealtime} /></>; }
  const query = new VoiceInteractionPreviewQuery(context);
  const settle = () => h.run(() => new Promise(r => setTimeout(r, 10)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: selected })}><VoiceInteractionStateProvider previewQuery={query} realtimeController={controller}><Surface /></VoiceInteractionStateProvider></MobileServerStateProvider>);
    await settle(); await settle(); await h.press(h.byLabel('Start')); expect(h.allText()).toContain('Garage:listening');
    selected = 'Kitchen'; await h.run(() => client.setQueryData(mobileQueryKeys.inventoryScope('scope'), { tenantId: 'tenant', inventoryId: selected })); await settle(); await settle();
    expect(cancelled).toBe(1); expect(h.allText()).toContain('Kitchen:ready'); expect(h.allText()).not.toContain('Garage:listening');
  } finally { await h.unmount(); }
});
