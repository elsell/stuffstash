import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react';
import { useLocalSearchParams } from 'expo-router';
import { createMobileQueryClient } from '../src/adapters/serverState/MobileQueryClient';
import { inventoryId, tenantId } from '../src/domain/inventories/InventorySummary';
import { VoiceInteractionPreviewQuery } from '../src/application/voice/VoiceInteractionPreviewQuery';
import { RealtimeVoiceSessionController, type VoiceRealtimeEvent } from '../src/application/voice/RealtimeVoiceSession';
import type { ParentLookupQuery, ParentLookupResult } from '../src/application/add/ParentLookupQuery';
import { MobileServerStateProvider } from '../src/ui/navigation/MobileServerStateProvider';
import { VoiceInteractionStateProvider, useVoiceInteractionState } from '../src/ui/navigation/VoiceInteractionStateContext';
import { VoiceSessionWorkspace } from '../src/ui/screens/VoiceSessionSheetScreen';
import { VoicePlanLocationRouteScreen } from '../src/ui/screens/VoicePlanLocationRouteScreen';

const scope = { tenantId: 'audit-voice-tenant', inventoryId: 'audit-voice-inventory' };
const unavailable = async (): Promise<never> => { throw new Error('Not supported by the proposal selection fixture'); };
const LookupContext = createContext<Pick<ParentLookupQuery, 'execute'> | null>(null);
const ActivateProposalContext = createContext<((active: boolean) => void) | null>(null);

export function VoiceProposalFixtureProvider({ children }: { readonly children: ReactNode }) {
  const [proposalActivated, setProposalActivated] = useState(false);
  const [fixture] = useState(() => {
    const context = { getVoiceInventoryContext: async () => ({ tenantId: tenantId(scope.tenantId), inventoryId: inventoryId(scope.inventoryId), tenantName: 'Audit home', inventoryName: 'Audit inventory' }), addInventoryAssetPhoto: unavailable };
    const controller = new RealtimeVoiceSessionController(context,
      { start: unavailable, stop: unavailable, cancel: async () => {}, recordingLevel: () => 0 },
      { run: async (_input: unknown, receive: (event: VoiceRealtimeEvent) => Promise<void>) => {
        await receive({ type: 'action.plan.proposed', seq: 1, sessionId: 'audit-session', actionPlan: {
          planId: 'audit-plan', status: 'proposed', confirmationSummary: 'Add the audit drill', risks: [],
          commands: [{ id: 'audit-drill', kind: 'create_asset', operation: 'create', title: 'Audit drill', summary: 'Add audit drill', assetKind: 'item' }]
        } });
      }, canSendFollowUpAudio: () => false, sendFollowUpAudio: unavailable, approveActionPlan: unavailable, cancelActionPlan: unavailable },
      { playChunk: async () => {}, stop: async () => {} });
    let failed = false;
    const lookup = { execute: async (query: string): Promise<readonly ParentLookupResult[]> => {
      if (!failed) { failed = true; throw new Error('Synthetic location lookup failure'); }
      if (query.trim() && !'garage bin'.includes(query.trim().toLowerCase())) return [];
      return [{ id: 'audit-bin', title: 'Garage bin', kind: 'container', subtitle: 'Container', pathLabel: 'Garage / Garage bin', selectionHint: '', willPromoteToContainer: false }];
    } };
    return { client: createMobileQueryClient(), preview: new VoiceInteractionPreviewQuery(context), controller, lookup };
  });
  useEffect(() => () => fixture.client.clear(), [fixture]);
  return <MobileServerStateProvider client={fixture.client} scopeId="audit-voice" loadInventoryScope={async () => scope}>
    <VoiceInteractionStateProvider previewQuery={fixture.preview} realtimeController={fixture.controller}>
      {proposalActivated && <VoiceProposalSeed />}
      <ActivateProposalContext.Provider value={setProposalActivated}>
        <LookupContext.Provider value={fixture.lookup}>{children}</LookupContext.Provider>
      </ActivateProposalContext.Provider>
    </VoiceInteractionStateProvider>
  </MobileServerStateProvider>;
}

function VoiceProposalSeed() {
  const { state, composerText, setComposerText, sendText } = useVoiceInteractionState();
  const seeded = useRef(false);
  const sent = useRef(false);
  useEffect(() => {
    if (state.status === 'ready' && !state.realtime?.actionPlan && !seeded.current) {
      seeded.current = true; setComposerText('Add the audit drill');
    }
  }, [state, setComposerText]);
  useEffect(() => {
    if (composerText === 'Add the audit drill' && !sent.current) { sent.current = true; void sendText(); }
  }, [composerText, sendText]);
  return null;
}

export function VoiceProposalFixture() {
  const activate = useContext(ActivateProposalContext);
  useEffect(() => { activate?.(true); }, [activate]);
  return <VoiceSessionWorkspace photoSelectionQuery={{ selectFromLibrary: unavailable, captureFromCamera: unavailable }} />;
}

export function VoicePlanLocationFixture() {
  const params = useLocalSearchParams<{ planId: string; commandId: string; scope: string }>();
  const lookup = useContext(LookupContext);
  if (!lookup) throw new Error('Missing proposal fixture provider');
  return <VoicePlanLocationRouteScreen params={params} parentLookupQuery={lookup} />;
}
