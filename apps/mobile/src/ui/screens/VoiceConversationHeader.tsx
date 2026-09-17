import { Stack } from 'expo-router';
import { useMemo } from 'react';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';
import type { VoicePlanCommandDrafts } from './VoicePlanEdits';
import type { VoicePlanPhotoDrafts } from './VoicePlanPhotoDraftState';
import { useNewConversation } from './useNewConversation';

export function VoiceConversationHeader({ realtime, photoDrafts, commandDrafts, onReset, onClose }: {
  readonly realtime: VoiceRealtimeState | null;
  readonly photoDrafts: VoicePlanPhotoDrafts;
  readonly commandDrafts: VoicePlanCommandDrafts;
  readonly onReset: () => void;
  readonly onClose: () => void;
}) {
  const startNew = useNewConversation(realtime, photoDrafts, commandDrafts, onReset);
  const leading = useNativeHeaderActionOptions([{ kind: 'close', label: 'Close voice session', onPress: onClose }], 'left');
  const trailing = useNativeHeaderActionOptions([{ kind: 'compose', label: 'New conversation', onPress: startNew }]);
  const options = useMemo(() => ({ title: 'Conversation', headerShown: true, headerBackVisible: false, ...leading, ...trailing }), [leading, trailing]);
  return <Stack.Screen options={options} />;
}
