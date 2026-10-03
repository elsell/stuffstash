import { t } from './localization';
import { PrintingFailure, type PrinterReadiness, type PrintJobStatus, type PrintConnector } from '$lib/domain/printing';
export function readinessLabel(value: PrinterReadiness) { return t(value === 'ready' ? 'web.Printing.ready' : value === 'unavailable' ? 'web.Printing.offline' : value === 'error' ? 'web.Printing.attention' : 'web.Printing.unknown'); }
export function jobStatusLabel(value: PrintJobStatus) { const keys = { queued: 'web.Printing.queued', claimed: 'web.Printing.claimed', printing: 'web.Printing.printing', completed: 'web.Printing.completed', failed: 'web.Printing.failed', uncertain: 'web.Printing.uncertain', canceled: 'web.Printing.canceled' } as const; return t(keys[value]); }
export function connectorStatusLabel(value: PrintConnector) { return t(value.authorizationPending ? 'web.Printing.pendingAuth' : value.state === 'revoked' ? 'web.Printing.revoked' : value.state === 'awaiting_activation' ? 'web.Printing.awaiting' : 'web.Printing.active'); }
export function printingFailureMessage(error: unknown) { if (error instanceof PrintingFailure) {
    if (error.kind === 'authentication_required')
        return t('web.Printing.authentication');
    if (error.kind === 'denied')
        return t('web.Printing.denied');
    if (error.kind === 'conflict')
        return t('web.Printing.conflict');
    if (error.kind === 'invalid')
        return t('web.Printing.invalid');
} return t('web.Printing.unavailable'); }
