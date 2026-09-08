import type { VoicePlanCommandDrafts } from '../screens/VoicePlanEdits';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';

export const conversationHistoryLimit = 20;
export function appendConversationExchange(history: readonly VoiceRealtimeState[], state: VoiceRealtimeState | null, drafts: VoicePlanCommandDrafts = {}): readonly VoiceRealtimeState[] {
  if (!state || !['completed', 'cancelled', 'failed'].includes(state.status) || (!state.transcript && !state.spokenResponse && !state.actionPlan)) return history;
  const exchange = state.actionPlan && Object.keys(drafts).length ? { ...state, actionPlan: { ...state.actionPlan,
    commands: state.actionPlan.commands.map(command => ({ ...command, ...(command.id && drafts[command.id]?.title ? { title: drafts[command.id].title } : {}) })) } } : state;
  return [...history, exchange].slice(-conversationHistoryLimit);
}
export function canSubmitConversation(stage: string): boolean {
  return ['ready', 'completed', 'cancelled', 'failed'].includes(stage);
}
