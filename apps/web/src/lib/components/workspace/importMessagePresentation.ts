import type { ImportMessage } from '$lib/domain/inventory';
import { t } from '$lib/presentation/localization';

type Guidance = 'duplicate' | 'partialDate' | 'download' | 'validation' | 'blocking' | 'warning';

/** Interpret source diagnostics before translating; display copy never controls behavior. */
export function importIssuePresentation(message: ImportMessage): { identity: string; cause: string; guidance: Guidance } {
  const detail = message.detail || '';
  const lower = detail.toLowerCase();
  let category = 'source';
  let cause = detail;
  if (message.code === 'duplicate-asset' || lower.includes('homebox-source-id')) {
    category = 'linked';
    cause = t('import.issue.linked');
  } else if (message.code === 'source-link-duplicate') {
    category = 'imported';
    cause = t('import.issue.imported');
  } else if (message.code === 'attachment-unavailable') {
    category = 'download';
    cause = t('import.issue.download');
  } else if (message.code === 'attachment-session-unavailable') {
    category = 'session';
    cause = detail || t('import.issue.session');
  } else if (message.code === 'attachment-storage-unavailable') {
    category = 'storage';
    cause = detail || t('import.issue.storage');
  } else if (lower.includes('import validation failed')) {
    category = 'validation';
    cause = t('import.issue.validation');
  }
  let guidance: Guidance = message.severity === 'error' ? 'blocking' : 'warning';
  if (category === 'linked' || category === 'imported') guidance = 'duplicate';
  else if (message.code === 'partial-date' || message.summary.toLowerCase().includes('partial date')) guidance = 'partialDate';
  else if (category === 'download' || category === 'session') guidance = 'download';
  else if (category === 'validation') guidance = 'validation';
  else if (category === 'source') {
    if (lower.includes('already')) guidance = 'duplicate';
    else if (lower.includes('download')) guidance = 'download';
    else if (lower.includes('attachment validation') || lower.includes('unsupported file type')) guidance = 'validation';
  }
  const identity = ['source', 'session', 'storage'].includes(category) ? `${category}:${detail}` : category;
  return { identity, cause, guidance };
}
