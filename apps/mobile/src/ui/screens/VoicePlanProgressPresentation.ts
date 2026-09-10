import type { VoicePlanCommandDrafts } from './VoicePlanEdits';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
export type VoicePlanProgressModel = { readonly title: string; readonly detail: string; readonly busy: boolean; readonly percent?: number };
export function voicePlanProgress(state: VoiceRealtimeState | null, drafts: VoicePlanCommandDrafts = {}): VoicePlanProgressModel | null {
  const plan = state?.actionPlan;
  if (!state || !plan) return null;
  if (plan.status === 'approved' || (plan.status === 'proposed' && state.reviewDecisionPending)) {
    if (state.progressLabel === 'Cancelling change') return { title: 'Cancelling change…', detail: 'Waiting for confirmation', busy: true };
    const titles = plan.commands.map(command => (command.id && drafts[command.id]?.title) || command.title).filter(Boolean);
    return { title: titles.length === 1 ? `Saving ${titles[0]}…` : 'Saving changes…', detail: 'Waiting for confirmation', busy: true };
  }
  if (plan.status !== 'executed') return null;
  const photos = state.photoAttachmentStatus;
  if (!photos) return { title: 'Saved', detail: 'Your inventory is up to date', busy: false };
  const busy = photos.status === 'uploading';
  const complete = photos.status === 'attached';
  const counts = photos.totalCount !== undefined && photos.attachedCount !== undefined;
  const failed = photos.failedCount ?? 0;
  return {
    title: busy ? 'Saved · adding photos' : complete ? 'Saved' : 'Saved · photos need attention',
    detail: counts ? `${photos.attachedCount} of ${photos.totalCount} photos attached${failed ? ` · ${failed} ${failed === 1 ? 'needs' : 'need'} attention` : ''}${photos.status === 'failed' ? `. ${photos.message}` : ''}` : photos.message,
    busy,
    ...(counts && photos.totalCount! > 0 ? { percent: Math.min(100, Math.max(0, Math.round(100 * photos.attachedCount! / photos.totalCount!))) } : {})
  };
}
