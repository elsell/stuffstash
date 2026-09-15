import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { VoiceConversationComposer } from './VoiceConversationComposer';

it('replaces Send with a single cancellation command during a cancellable request', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' }), addInventoryAssetPhoto: async () => { throw new Error('Unused'); } };
  const controller = new RealtimeVoiceSessionController(context,
    { start: async () => {}, stop: async () => ({ mimeType: 'audio/mp4', sampleRate: 44100, channels: 1, chunksBase64: [] }), cancel: async () => {}, recordingLevel: () => 0 },
    { run: async () => {}, canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => {}, approveActionPlan: async () => {}, cancelActionPlan: async () => {} },
    { playChunk: async () => {}, stop: async () => {} });
  let interaction!: ReturnType<typeof useVoiceInteractionState>; let recordings = 0;
  function Surface() { interaction = useVoiceInteractionState(); return <VoiceConversationComposer onMic={() => { recordings++; }} />; }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}><Surface /></VoiceInteractionStateProvider>
    </MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
    await h.press(h.byLabel('Start recording')); expect(recordings).toBe(1);
    await h.run(() => interaction.setComposerText('Find my drill'));
    expect(h.byLabel('Send message')).toBeDefined();
    await h.run(() => interaction.setStage('processing'));
    expect(h.allByType('Pressable').map(node => node.props.accessibilityLabel)).toEqual(['Cancel request']);
    expect(h.byLabel('Message Stuff Stash')?.props.editable).toBe(false);
    await h.run(() => interaction.setStage('review'));
    expect(h.byLabel('Cancel request')).toBeUndefined();
    expect(h.byLabel('Send message')).toBeUndefined();
  } finally { await h.unmount(); client.clear(); }
});
