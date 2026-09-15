import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { Alert } from 'react-native';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
import { shouldConfirmNewConversation } from '../navigation/VoiceConversationHistory';
import type { VoicePlanPhotoDrafts } from './VoicePlanPhotoDraftState';
import type { VoicePlanCommandDrafts } from './VoicePlanEdits';

export function useNewConversation(realtime: VoiceRealtimeState | null, photoDrafts: VoicePlanPhotoDrafts, commandDrafts: VoicePlanCommandDrafts, onReset: () => void) {
  const capturePresentation = useTaskPresentation(undefined, JSON.stringify([
    realtime?.tenantName, realtime?.inventoryName, realtime?.status,
    realtime?.actionPlan?.planId, realtime?.actionPlan?.status,
    realtime?.reviewDecisionPending, realtime?.photoAttachmentStatus?.status,
    photoDrafts, commandDrafts
  ]));
  return () => {
    const isCurrent = capturePresentation();
    if (!isCurrent()) return;
    let accepted = false;
    const accept = () => {
      if (!isCurrent() || accepted) return;
      accepted = true;
      onReset();
    };
    if (shouldConfirmNewConversation(realtime)) {
      Alert.alert('Start a new conversation?', 'This clears the conversation and staged photos. It does not undo changes already submitted.', [
        { text: 'Keep conversation', style: 'cancel' },
        { text: 'New conversation', style: 'destructive', onPress: accept }
      ]);
    } else { accept(); }
  };
}
