import { expect, it } from 'vitest';
import { voicePlanProgress } from './VoicePlanProgressPresentation';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
const state: VoiceRealtimeState = { status: 'processing', tenantName: 'Home', inventoryName: 'Main', progressLabel: 'Applying change', debugEvents: [], actionPlan: { planId: 'plan', status: 'approved', confirmationSummary: 'Add Baby', risks: [], commands: [{ id: 'baby', kind: 'create_asset', title: 'Baby', summary: 'Add Baby' }] } };
it('names the save without inventing measured progress', () => {
 expect(voicePlanProgress(state)).toMatchObject({ title: 'Saving Baby…', detail: 'Waiting for confirmation', busy: true });
 expect(voicePlanProgress(state)?.percent).toBeUndefined();
});
it('separates saved changes from measured photo progress and failures', () => {
 const saved = { ...state, actionPlan: { ...state.actionPlan!, status: 'executed' as const }, photoAttachmentStatus: { status: 'uploading' as const, message: 'Adding photos', attachedCount: 1, totalCount: 3, failedCount: 1 } };
 expect(voicePlanProgress(saved)).toMatchObject({ title: 'Saved · adding photos', detail: '1 of 3 photos attached · 1 needs attention', percent: 33, busy: true });
 expect(voicePlanProgress({ ...saved, photoAttachmentStatus: { ...saved.photoAttachmentStatus, status: 'partial_failed', canRetry: true } })).toMatchObject({ title: 'Saved · photos need attention', percent: 33, busy: false });
});
it('does not label proposal rejection as a save', () => {
 expect(voicePlanProgress({ ...state, progressLabel: 'Cancelling change', reviewDecisionPending: true, actionPlan: { ...state.actionPlan!, status: 'proposed' } })?.title).toBe('Cancelling change…');
});
it('uses the name approved in review and preserves upload failure details', () => {
 expect(voicePlanProgress(state, { baby: { title: 'Baby toy' } })?.title).toBe('Saving Baby toy…');
 expect(voicePlanProgress({ ...state, actionPlan: { ...state.actionPlan!, status: 'executed' }, photoAttachmentStatus: { status: 'failed', message: 'Upload permission expired. Retry photos.', attachedCount: 0, totalCount: 1, failedCount: 1 } })?.detail).toContain('Upload permission expired');
});
