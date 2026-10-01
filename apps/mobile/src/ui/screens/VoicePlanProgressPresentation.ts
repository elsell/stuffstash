import { t } from '../../presentation/localization';
import type { VoicePlanCommandDrafts } from './VoicePlanEdits';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
export type VoicePlanProgressModel = { readonly title: string; readonly detail: string; readonly busy: boolean; readonly needsAttention?: boolean; readonly percent?: number };
export function voicePlanProgress(state: VoiceRealtimeState | null, drafts: VoicePlanCommandDrafts = {}): VoicePlanProgressModel | null {
  const plan = state?.actionPlan;
  if (!state || !plan) return null;
  if (plan.status === 'approved' || (plan.status === 'proposed' && state.reviewDecisionPending)) {
    if (state.progressLabel === 'Cancelling change') return { title: t('mobile.VoicePlanProgressPresentation.cancellingChange'), detail: t('mobile.voice.waitingConfirmation'), busy: true };
    const titles = plan.commands.map(command => (command.id && drafts[command.id]?.title) || command.title).filter(Boolean);
    return { title: titles.length === 1 ? t('mobile.VoicePlanProgressPresentation.saving', { value: String(titles[0]) }) : t('mobile.VoicePlanProgressPresentation.savingChanges'), detail: t('mobile.voice.waitingConfirmation'), busy: true };
  }
  if (plan.status !== 'executed') return null;
  const photos = state.photoAttachmentStatus;
  if (!photos) return { title: t('mobile.VoicePlanProgressPresentation.saved'), detail: t('mobile.voice.inventoryUpToDate'), busy: false };
  const busy = photos.status === 'uploading';
  const complete = photos.status === 'attached';
  const needsAttention = photos.status === 'failed' || photos.status === 'partial_failed';
  const counts = photos.totalCount !== undefined && photos.attachedCount !== undefined;
  const failed = photos.failedCount ?? 0;
  return {
    title: busy ? t('mobile.VoicePlanProgressPresentation.savedAddingPhotos') : complete ? t('mobile.VoicePlanProgressPresentation.saved') : t('mobile.VoicePlanProgressPresentation.savedPhotosNeedAttention'),
    detail: counts ? t(failed ? 'mobile.voice.photoProgressAttention' : 'mobile.voice.photoProgress', { attached: photos.attachedCount!, total: photos.totalCount!, count: failed, detail: needsAttention ? t('mobile.voice.photoProgressDetail', { message: photos.message }) : '' }) : photos.message,
    busy,
    needsAttention,
    ...(counts && photos.totalCount! > 0 ? { percent: Math.min(100, Math.max(0, Math.round(100 * photos.attachedCount! / photos.totalCount!))) } : {})
  };
}
