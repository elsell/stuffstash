import { useCallback, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { Keyboard } from 'react-native';
import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';

export function useVoiceReferenceNavigation({ scopeIdentity, pauseMedia, onOpen }: {
  readonly scopeIdentity: string; readonly pauseMedia: () => Promise<void>;
  readonly onOpen: (reference: VoiceResponseArtifact) => void;
}) {
  const [visit, setVisit] = useState<{ active: boolean; scope: string }>();
  useFocusEffect(useCallback(() => {
    const current = { active: true, scope: scopeIdentity };
    setVisit(current);
    return () => { current.active = false; };
  }, [scopeIdentity]));
  return async (reference: VoiceResponseArtifact) => {
    if (!visit?.active || visit.scope !== scopeIdentity) return;
    Keyboard.dismiss();
    await pauseMedia();
    if (!visit.active) return;
    onOpen(reference);
  };
}
