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
    readonly progressTrace?: readonly string[];
  },
  diagnosticsEnabled: boolean
): VoiceSessionSheetBodyPresentation {
  return {
    hasBodyContent: Boolean(
      session.response ||
        session.transcript ||
        session.actionPlan ||
        (session.progressTrace?.length ?? 0) > 0 ||
        (state.status === 'ready' && state.realtime?.errorMessage) ||
        (diagnosticsEnabled && state.status === 'ready' && state.realtime?.debugEvents.length)
    )
  };
}
