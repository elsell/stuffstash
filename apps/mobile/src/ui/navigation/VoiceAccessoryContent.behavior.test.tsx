import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController, type VoiceRealtimeEvent } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from './MobileServerStateProvider';
import { VoiceInteractionStateProvider } from './VoiceInteractionStateContext';
import { VoiceAccessoryContent } from './VoiceAccessoryContent';
import { VoiceTabContent } from './VoiceTabContent';
import { dispatchedActions, resetNavigation } from '../../test-support/navigation';
import { setPathname } from '../../test-support/expo-router';

it('starts and sends a fresh voice interaction from the fallback and reopens it after returning to Browse', async () => {
  const h = new MobileRenderHarness();
  resetNavigation(); setPathname('/search');
  let starts = 0;
  let stops = 0;
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' }), addInventoryAssetPhoto: async () => { throw new Error('Unused'); } };
  const controller = new RealtimeVoiceSessionController(context,
    { start: async () => { starts++; }, stop: async () => { stops++; return { mimeType: 'audio/mp4', sampleRate: 44100, channels: 1, chunksBase64: [] }; }, cancel: async () => undefined, recordingLevel: () => 0 },
    { run: async (_input: unknown, receive: (event: VoiceRealtimeEvent) => Promise<void>) => { await receive({ seq: 1, type: 'session.completed' }); },
      canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => undefined, approveActionPlan: async () => undefined, cancelActionPlan: async () => undefined },
    { playChunk: async () => undefined, stop: async () => undefined });
  const client = createMobileQueryClient();
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 10)));
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}>
        <VoiceTabContent platform="android" version={35} accessory={<VoiceAccessoryContent placement="regular" />}>{null}</VoiceTabContent>
      </VoiceInteractionStateProvider>
    </MobileServerStateProvider>);
    await settle(); await settle();
    await h.press(h.byLabel('Start voice interaction')); await settle();
    expect(starts).toBe(1);
    expect(dispatchedActions()).toEqual([{ type: 'navigate', href: '/voice' }]);
    await h.run(() => setPathname('/voice'));
    await h.press(h.byLabel('Send voice request')); await settle();
    expect(stops).toBe(1);
    expect(dispatchedActions()).toHaveLength(1);
    await h.run(() => setPathname('/search'));
    const open = h.allByType('Pressable').find(node => String(node.props.accessibilityLabel).startsWith('Open voice session.'));
    await h.press(open);
    expect(dispatchedActions()).toHaveLength(2);
  } finally { await h.unmount(); client.clear(); setPathname('/'); resetNavigation(); }
});
