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
import { VoicePreviewRecovery } from '../screens/VoicePreviewRecovery';
import { setScreenFocused } from '../../test-support/navigation';

it.each(['scope', 'context'])('recovers failed initial %s from the conversation Retry command', async failure => {
  setScreenFocused(true);
  const h = new MobileRenderHarness(); const client = createMobileQueryClient(); let attempts = 0;
  const failTwice = () => { if (++attempts <= 2) throw new Error('Context temporarily unavailable'); };
  const context = { getVoiceInventoryContext: async () => {
    if (failure === 'context') failTwice();
    return { tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' };
  } };
  const unused = async (): Promise<never> => { throw new Error('Unused'); };
  const controller = new RealtimeVoiceSessionController(context,
    { start: unused, stop: unused, cancel: async () => {}, recordingLevel: () => 0 },
    { run: unused, canSendFollowUpAudio: () => false, sendFollowUpAudio: unused, approveActionPlan: unused, cancelActionPlan: unused },
    { playChunk: async () => {}, stop: async () => {} });
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
  function Surface() {
    const { state, scopeIdentity, retryPreview } = useVoiceInteractionState();
    return state.status === 'error'
      ? <VoicePreviewRecovery key={scopeIdentity} message={state.message} identity={scopeIdentity} onRetry={retryPreview} />
      : <Text>{state.status === 'ready' ? 'Conversation ready' : 'Loading'}</Text>;
  }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => {
      if (failure === 'scope') failTwice(); return { tenantId: 'tenant', inventoryId: 'inventory' };
    }}><VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}>
      <Surface />
    </VoiceInteractionStateProvider></MobileServerStateProvider>);
    await settle(); await settle();
    expect(h.allText()).toContain('Voice unavailable');
    await h.press(h.byLabel('Retry conversation')); await settle();
    expect(attempts).toBe(2); expect(h.allText()).toContain('Voice unavailable');
    await h.press(h.byLabel('Retry conversation')); await settle(); await settle();
    expect(attempts).toBe(3);
    expect(h.allText()).not.toContain('Voice unavailable');
    expect(h.allText()).toContain('Conversation ready');
  } finally { await h.unmount(); client.clear(); }
});

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
