import { describe, expect, it } from 'vitest';
import { appendConversationExchange, canSubmitConversation } from './VoiceConversationHistory';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
const answer: VoiceRealtimeState = { status: 'completed', transcript: 'Where is the drill?', spokenResponse: 'In the toolbox.', tenantName: 'Home', inventoryName: 'Home', progressLabel: 'Done', debugEvents: [] };
describe('conversation history', () => {
  it('preserves prior user and assistant words and bounds memory', () => {
    const history = Array.from({ length: 20 }, () => answer);
    const next = { ...answer, transcript: 'Which toolbox?' };
    const result = appendConversationExchange(history, next);
    expect(result).toHaveLength(20);
    expect(result[19]).toBe(next);
    expect(result[0].spokenResponse).toBe('In the toolbox.');
    expect(history[19]).toBe(answer);
  });
  it('does not archive unfinished work or allow a request to replace review', () => {
    expect(appendConversationExchange([], { ...answer, status: 'processing' })).toEqual([]);
    for (const stage of ['review', 'processing', 'speaking', 'listening']) expect(canSubmitConversation(stage)).toBe(false);
    expect(canSubmitConversation('completed')).toBe(true);
  });
});

it('retains edited proposal titles in the completed exchange', () => {
  const result = appendConversationExchange([], { ...answer, actionPlan: { planId: 'plan', status: 'executed', confirmationSummary: 'Add drill', risks: [], commands: [{ id: 'drill', kind: 'create_asset', title: 'Drill', summary: 'Add drill' }] } }, { drill: { title: 'Cordless drill' } });
  expect(result[0].actionPlan?.commands[0].title).toBe('Cordless drill');
});

it('does not offer cancellation after a review decision or confirmed save', async () => {
  const { canCancelConversation } = await import('./VoiceConversationHistory');
  const base = { status: 'processing' as const, tenantName: 'Home', inventoryName: 'Main', progressLabel: 'Working', debugEvents: [] };
  const actionPlan = { planId: 'plan', status: 'proposed' as const, commands: [], risks: [], confirmationSummary: 'Add bottle' };
  expect(canCancelConversation('processing', base)).toBe(true);
  expect(canCancelConversation('processing', { ...base, actionPlan, reviewDecisionPending: true })).toBe(false);
  expect(canCancelConversation('processing', { ...base, actionPlan: { ...actionPlan, status: 'approved' } })).toBe(false);
  expect(canCancelConversation('processing', { ...base, actionPlan: { ...actionPlan, status: 'executed' } })).toBe(false);
});

it('requires reset confirmation while photos remain unfinished', async () => {
  const { shouldConfirmNewConversation } = await import('./VoiceConversationHistory');
  expect(shouldConfirmNewConversation(answer)).toBe(false);
  expect(shouldConfirmNewConversation({ ...answer, photoAttachmentStatus: { status: 'uploading', message: 'Adding photos' } })).toBe(true);
  expect(shouldConfirmNewConversation({ ...answer, photoAttachmentStatus: { status: 'failed', message: 'Retry photos', canRetry: true } })).toBe(true);
});
