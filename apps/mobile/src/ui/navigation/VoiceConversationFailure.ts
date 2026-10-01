import { t } from "../../presentation/localization";
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
      ? t("voice.disconnectedReview", { reason: failure.errorMessage ?? "" })
      : failure.errorMessage
  };
}
