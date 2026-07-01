import type { VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';

export type VoiceSessionSheetBodyPresentation = {
  readonly hasBodyContent: boolean;
};

export function buildVoiceSessionSheetBodyPresentation(
  state: VoiceInteractionState,
  session: {
    readonly response?: string;
    readonly transcript?: string;
    readonly actionPlan?: unknown;
  },
  diagnosticsEnabled: boolean
): VoiceSessionSheetBodyPresentation {
  return {
    hasBodyContent: Boolean(
      session.response ||
        session.transcript ||
        session.actionPlan ||
        (state.status === 'ready' && state.realtime?.errorMessage) ||
        (diagnosticsEnabled && state.status === 'ready' && state.realtime?.debugEvents.length)
    )
  };
}
