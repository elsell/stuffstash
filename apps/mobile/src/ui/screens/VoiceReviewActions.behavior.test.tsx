import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController, type VoiceRealtimeEvent } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { VoiceReviewActions } from './VoiceReviewActions';

it('retained voice decisions use current drafts only within the visible review plan', async () => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' }), addInventoryAssetPhoto: async () => { throw new Error('Unused'); } };
  const submitted: unknown[] = []; const cancelled: string[] = [];
  let plan = 'first';
  let deliver!: (event: VoiceRealtimeEvent) => Promise<void>;
  const controller = new RealtimeVoiceSessionController(context,
    { start: async () => {}, stop: async () => { throw new Error('Unused'); }, cancel: async () => {}, recordingLevel: () => 0 },
    { run: async (_input: unknown, receive: (event: VoiceRealtimeEvent) => Promise<void>) => {
        deliver = receive;
        await receive({ type: 'action.plan.proposed', seq: 1, actionPlan: { planId: plan, status: 'proposed', confirmationSummary: 'Create a tent?', commands: [{ id: 'item', kind: 'create_asset', operation: 'create', summary: 'Create tent' }], risks: [] } });
        await receive({ type: 'session.completed', seq: 2 });
      }, canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => {},
      approveActionPlan: async (id, photos, edits) => { submitted.push({ id, photos, edits }); throw new Error('Connection interrupted'); },
      cancelActionPlan: async id => { cancelled.push(id); } },
    { playChunk: async () => {}, stop: async () => {} });
  let interaction!: ReturnType<typeof useVoiceInteractionState>;
  function Surface() { interaction = useVoiceInteractionState(); const current = interaction.state.status === 'ready' ? interaction.state.realtime?.actionPlan : undefined; return current?.status === 'proposed' ? <VoiceReviewActions planId={current.planId} /> : null; }
  const propose = async () => {
    await h.run(() => interaction.setComposerText('Create a tent'));
    await h.run(() => interaction.sendText());
  };
  let approve!: () => void; let cancel!: () => void;
  try {
    await h.render(<MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
      <VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}><Surface /></VoiceInteractionStateProvider>
    </MobileServerStateProvider>);
    await h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
    await propose();
    await h.run(() => interaction.setCommandDraftState({ planId: plan, drafts: { item: { title: 'Earlier name' } } }));
    await h.run(() => interaction.setPhotoDrafts({ item: [{ id: 'old', uri: 'file:///old.jpg', fileName: 'old.jpg', contentType: 'image/jpeg', sizeBytes: 5 }] }));
    approve = h.byLabel('Approve voice change')!.props.onPress;
    cancel = h.byLabel('Cancel voice change')!.props.onPress;
    await h.run(() => interaction.setTitleEditor({ commandId: 'item', value: ' ' }));
    await h.run(approve); expect(submitted).toEqual([]);
    await h.run(() => interaction.setTitleEditor(null));
    await h.run(() => interaction.setCommandDraftState({ planId: plan, drafts: { item: { title: 'Current name', parent: { kind: 'root', label: 'Inventory root' } } } }));
    await h.run(() => interaction.setPhotoDrafts({ item: [{ id: 'new', uri: 'file:///new.jpg', fileName: 'new.jpg', contentType: 'image/jpeg', sizeBytes: 7 }] }));
    await h.run(() => setScreenFocused(false));
    await h.run(approve); await h.run(cancel);
    expect(submitted).toEqual([]); expect(cancelled).toEqual([]);
    await h.run(() => setScreenFocused(true));
    await h.run(approve);
    expect(submitted).toEqual([{ id: 'first', photos: [{ commandId: 'item', photoIndex: 0, fileName: 'new.jpg', contentType: 'image/jpeg', sizeBytes: 7 }], edits: [{ commandId: 'item', title: 'Current name', parent: { kind: 'root' } }] }]);
    submitted.length = 0;
    await h.run(() => interaction.reset()); plan = 'second'; await propose();
    await h.run(approve); await h.run(cancel);
    expect(submitted).toEqual([]); expect(cancelled).toEqual([]);
    const secondApprove = h.byLabel('Approve voice change')!.props.onPress;
    const secondCancel = h.byLabel('Cancel voice change')!.props.onPress;
    await h.run(() => deliver({ type: 'action.plan.proposed', seq: 3, actionPlan: { planId: 'third', status: 'proposed', confirmationSummary: 'Create a box?', commands: [{ id: 'item', kind: 'create_asset', operation: 'create', summary: 'Create box' }], risks: [] } }));
    expect(interaction.state.status === 'ready' && interaction.state.realtime?.actionPlan?.planId).toBe('third');
    await h.run(secondApprove); await h.run(secondCancel);
    expect(submitted).toEqual([]); expect(cancelled).toEqual([]);
    await h.press(h.byLabel('Cancel voice change'));
    expect(cancelled).toEqual(['third']);
  } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
  cancelled.length = 0;
  await h.run(approve); await h.run(cancel);
  expect(submitted).toEqual([]); expect(cancelled).toEqual([]);
});
