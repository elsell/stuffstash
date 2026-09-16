import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { attemptNavigation, dispatchedActions, resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { createMobileQueryClient } from '../../adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../../application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController, type VoiceRealtimeEvent } from '../../application/voice/RealtimeVoiceSession';
import { MobileServerStateProvider } from '../navigation/MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { VoicePlanLocationRouteScreen } from './VoicePlanLocationRouteScreen';

it.each(['current', 'scope', 'plan', 'left', 'back'] as const)('owns proposal destination choices: %s', async scenario => {
  const h = new MobileRenderHarness(); const client = createMobileQueryClient();
  const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId('tenant'), inventoryId: inventoryId('inventory'), tenantName: 'Home', inventoryName: 'Inventory' }), addInventoryAssetPhoto: async () => { throw new Error('Unused'); } };
  const controller = new RealtimeVoiceSessionController(context,
    { start: async () => {}, stop: async () => { throw new Error('Unused'); }, cancel: async () => {}, recordingLevel: () => 0 },
    { run: async (_input: unknown, receive: (event: VoiceRealtimeEvent) => Promise<void>) => {
      await receive({ type: 'action.plan.proposed', seq: 1, sessionId: 'session', actionPlan: { planId: 'plan', status: 'proposed', confirmationSummary: 'Add drill', risks: [], commands: [{ id: 'item', kind: 'create_asset', operation: 'create', title: 'Drill', summary: 'Add drill' }] } });
    }, canSendFollowUpAudio: () => false, sendFollowUpAudio: async () => {}, approveActionPlan: async () => {}, cancelActionPlan: async () => {} },
    { playChunk: async () => {}, stop: async () => {} });
  let interaction!: ReturnType<typeof useVoiceInteractionState>;
  let routeScope: string | undefined; let routePlan = 'plan'; let reads = 0;
  const lookup = { execute: async () => { reads++; return []; } };
  function Surface() {
    interaction = useVoiceInteractionState();
    return <VoicePlanLocationRouteScreen parentLookupQuery={lookup} params={{ planId: routePlan, commandId: 'item', scope: routeScope ?? interaction.scopeIdentity }} />;
  }
  const tree = () => <MobileServerStateProvider client={client} scopeId="scope" loadInventoryScope={async () => ({ tenantId: 'tenant', inventoryId: 'inventory' })}>
    <VoiceInteractionStateProvider previewQuery={new VoiceInteractionPreviewQuery(context)} realtimeController={controller}><Surface /></VoiceInteractionStateProvider>
  </MobileServerStateProvider>;
  try {
    resetNavigation(); setScreenFocused(true); await h.render(tree());
    await h.run(() => new Promise(resolve => setTimeout(resolve, 20)));
    await h.run(() => interaction.setComposerText('Add a drill'));
    await h.run(() => interaction.sendText());
    await h.run(() => interaction.setCommandDraftState({ planId: 'plan', drafts: { item: { title: 'My drill' } } }));
    await h.run(() => interaction.setTitleEditor({ commandId: 'item', value: 'Pending name' }));
    const retained = h.byLabel('Select Inventory root');
    expect(retained).toBeDefined();
    if (scenario === 'scope') { routeScope = 'other-inventory'; await h.render(tree()); }
    if (scenario === 'plan') { routePlan = 'other-plan'; await h.render(tree()); }
    if (scenario === 'left') await h.run(() => setScreenFocused(false));
    if (scenario === 'back') await h.run(() => { attemptNavigation({ type: 'back' }); setScreenFocused(false); });
    await h.press(retained);
    expect(interaction.commandDraftState.drafts.item).toEqual(scenario === 'current'
      ? { title: 'My drill', parent: { kind: 'root', label: 'Inventory root' } } : { title: 'My drill' });
    expect(interaction.titleEditor?.value).toBe('Pending name');
    expect(dispatchedActions()).toEqual(scenario === 'current' || scenario === 'back' ? [{ type: 'back' }] : []);
    if (scenario === 'scope' || scenario === 'plan') expect(h.allText()).toContain('This proposal is no longer available for editing.');
    expect(reads).toBeGreaterThan(0);
  } finally { await h.unmount(); client.clear(); setScreenFocused(true); }
});
