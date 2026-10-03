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
