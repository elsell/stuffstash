import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';

export function retainFailedConversation(current: VoiceRealtimeState | null, failure: VoiceRealtimeState): VoiceRealtimeState {
  const plan = current?.actionPlan;
  const interrupted = plan?.status === 'proposed' || plan?.status === 'approved';
  return {
    ...current, ...failure,
    transcript: current?.transcript,
    spokenResponse: current?.spokenResponse,
    responseArtifacts: current?.responseArtifacts,
    actionPlan: interrupted ? { ...plan, status: 'failed' } : plan,
    reviewDecisionPending: false,
    errorMessage: interrupted
      ? `${failure.errorMessage} This review is disconnected. Your draft is kept here. Check the inventory before submitting the change again.`
      : failure.errorMessage
  };
}
