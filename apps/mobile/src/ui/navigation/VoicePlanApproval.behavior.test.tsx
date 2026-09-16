import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController, type VoiceActionPlanCommandEdit } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from './MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from './VoiceInteractionStateContext';

it('blocks blank pending names and approves the visible name without losing placement or retry drafts', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' }), addInventoryAssetPhoto: async () => { throw new Error('Unused'); } };
  const submitted: (readonly VoiceActionPlanCommandEdit[] | undefined)[] = [];
  const controller = new RealtimeVoiceSessionController(context,
    { start: async () => {}, stop: async () => { throw new Error('Unused'); }, cancel: async () => {}, recordingLevel: () => 0 },
    { run: async () => {}, canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => {},
      approveActionPlan: async (_plan, _photos, edits) => { submitted.push(edits); throw new Error('Connection interrupted'); },
      cancelActionPlan: async () => {} },
    { playChunk: async () => {}, stop: async () => {} });
  let interaction!: ReturnType<typeof useVoiceInteractionState>;
  function Surface() { interaction = useVoiceInteractionState(); return null; }
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}><Surface /></VoiceInteractionStateProvider>
    </MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
    await h.run(() => interaction.setCommandDraftState({ planId: 'plan', drafts: { item: { title: 'Old name', parent: { kind: 'root', label: 'Inventory root' } } } }));
    await h.run(() => interaction.setTitleEditor({ commandId: 'item', value: '   ' }));
    await h.run(() => interaction.approveRealtimeActionPlan('plan', {}, [{ commandId: 'item', title: 'Old name', parent: { kind: 'root' } }]));
    expect(submitted).toEqual([]);
    expect(interaction.titleEditor?.value).toBe('   ');
    await h.run(() => interaction.setTitleEditor({ commandId: 'item', value: '  New   name  ' }));
    await h.run(() => interaction.approveRealtimeActionPlan('plan', {}, [{ commandId: 'item', title: 'Old name', parent: { kind: 'root' } }, { commandId: 'other', title: 'Other item' }]));
    expect(submitted).toEqual([[{ commandId: 'item', title: 'New name', parent: { kind: 'root' } }, { commandId: 'other', title: 'Other item' }]]);
    expect(interaction.commandDraftState.drafts.item).toEqual({ title: 'New name', parent: { kind: 'root', label: 'Inventory root' } });
    expect(interaction.titleEditor).toBeNull();
  } finally { await h.unmount(); client.clear(); }
});
