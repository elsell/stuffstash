import { NativeSheetActions } from '../components/NativeSheetActions';
import { useFocusedSheetActions } from '../components/useFocusedSheetActions';
import { useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { voicePlanCommandEdits } from './VoicePlanEdits';

/** Retired review events must never be retargeted to a replacement plan. */
export function VoiceReviewActions({ planId }: { readonly planId: string }) {
  const { scopeIdentity, state } = useVoiceInteractionState();
  const realtime = state.status === 'ready' ? state.realtime : null;
  if (realtime?.actionPlan?.planId !== planId || realtime.actionPlan.status !== 'proposed' || realtime.reviewDecisionPending) return null;
  return <ReviewDecision key={JSON.stringify([scopeIdentity, planId])} planId={planId} />;
}

function ReviewDecision({ planId }: { readonly planId: string }) {
  const { titleEditor, photoDrafts, commandDraftState, approveRealtimeActionPlan, cancelRealtimeActionPlan } = useVoiceInteractionState();
  const drafts = commandDraftState.planId === planId ? commandDraftState.drafts : {};
  const actions = useFocusedSheetActions({
    primaryLabel: 'Approve', primaryAccessibilityLabel: 'Approve voice change',
    secondaryLabel: 'Cancel', secondaryAccessibilityLabel: 'Cancel voice change',
    disabled: !!titleEditor && !titleEditor.value.trim(),
    onApply: () => { void approveRealtimeActionPlan(planId, photoDrafts, voicePlanCommandEdits(drafts)); },
    onBack: () => { void cancelRealtimeActionPlan(planId); }
  });
  return <NativeSheetActions keyboardAvoidance="container" {...actions} />;
}
