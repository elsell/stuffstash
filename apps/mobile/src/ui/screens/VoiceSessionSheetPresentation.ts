import type { VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';

export type VoiceSessionSheetBodyPresentation = {
  readonly hasBodyContent: boolean;
};

export function buildVoiceSessionSheetBodyPresentation(
  state: VoiceInteractionState,
  session: {
    readonly progressSteps?: readonly string[];
    readonly response?: string;
    readonly transcript?: string;
  },
  diagnosticsEnabled: boolean
): VoiceSessionSheetBodyPresentation {
  return {
    hasBodyContent: Boolean(
      session.response ||
        session.transcript ||
        session.progressSteps?.length ||
        (state.status === 'ready' && state.realtime?.errorMessage) ||
        (diagnosticsEnabled && state.status === 'ready' && state.realtime?.debugEvents.length)
    )
  };
}
