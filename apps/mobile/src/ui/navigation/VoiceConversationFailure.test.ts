import { expect, it } from 'vitest';
import { retainFailedConversation } from './VoiceConversationFailure';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
it('keeps the visible review and transcript when its socket disconnects', () => {
  const current: VoiceRealtimeState = {status:'review',tenantName:'Home',inventoryName:'Main',transcript:'Add wipes under sink',progressLabel:'Review',debugEvents:[],actionPlan:{planId:'plan',status:'proposed',confirmationSummary:'Add wipes',commands:[{id:'wipes',kind:'create_asset',title:'Clorox wipes',summary:'Create wipes'}],risks:[]}};
  const failed: VoiceRealtimeState = {status:'failed',tenantName:'Home',inventoryName:'Main',progressLabel:'Connection lost',debugEvents:[],errorMessage:'Connection interrupted.'};
  const result = retainFailedConversation(current,failed);
  expect(result.transcript).toBe(current.transcript);
  expect(result.actionPlan?.commands).toEqual(current.actionPlan?.commands);
  expect(result.actionPlan?.planId).toBe('plan');
  expect(result.actionPlan?.status).toBe('failed');
  expect(result.errorMessage).toContain('review');
  expect(result.reviewDecisionPending).toBe(false);
});
