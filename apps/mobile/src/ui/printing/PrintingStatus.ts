import { t } from '../../presentation/localization';
export function printJobStatus(status: string) {
  switch (status) {
    case 'queued': return t('printing.mobile.status.queued');
    case 'claimed': return t('printing.mobile.status.claimed');
    case 'printing': return t('printing.mobile.status.printing');
    case 'completed': return t('printing.mobile.status.completed');
    case 'failed': return t('printing.mobile.status.failed');
    case 'canceled': return t('printing.mobile.status.canceled');
    case 'uncertain': return t('printing.mobile.status.uncertain');
    default: return t('printing.mobile.status.unknown');
  }
}
export function printerReadiness(state: string) {
  switch (state) { case 'ready': return t('printing.mobile.readiness.ready'); case 'unavailable': return t('printing.mobile.readiness.unavailable'); case 'error': return t('printing.mobile.readiness.error'); default: return t('printing.mobile.readiness.unknown'); }
}
export function connectorState(state: string) {
  switch (state) { case 'active': return t('printing.mobile.state.active'); case 'pending': return t('printing.mobile.state.pending'); case 'revoked': return t('printing.mobile.state.revoked'); default: return t('printing.mobile.state.unknown'); }
}

export function connectorAvailability(state: string) {
  switch (state) { case 'online': return t('printing.mobile.online'); case 'offline': return t('printing.mobile.offline'); default: return t('printing.mobile.connectionUnknown'); }
}
export function printerAttention(reason?: string) {
  switch (reason) { case 'paper_empty': return t('printing.mobile.paperEmpty'); case 'cover_open': return t('printing.mobile.coverOpen'); case 'device_busy': return t('printing.mobile.deviceBusy'); case 'hardware_error': return t('printing.mobile.hardwareError'); default: return undefined; }
}

export function mediaSizeLabel(name: string, media: { readonly widthMicrometers: number; readonly heightMicrometers: number }) {
  if (name.trim()) return name;
  if (!(media.widthMicrometers > 0) || !(media.heightMicrometers > 0) || !Number.isFinite(media.widthMicrometers) || !Number.isFinite(media.heightMicrometers)) return t('printing.mobile.sizeUnavailable');
  return t('labels.mobile.sizeValue', { width: (media.widthMicrometers / 1000).toLocaleString(), height: (media.heightMicrometers / 1000).toLocaleString() });
}
