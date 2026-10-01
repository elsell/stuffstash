import { t } from '../../presentation/localization';
import type { AssetActivityEntry, AssetActivityRecordViewModel } from '../../application/assets/AssetActivityQuery';
export function groupHistoryRecords(records: readonly AssetActivityRecordViewModel[]): readonly { readonly title: string; readonly data: readonly AssetActivityRecordViewModel[] }[] {
  const formatter = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' });
  const sections = new Map<string, { title: string; data: AssetActivityRecordViewModel[] }>();
  for (const record of records) {
    const date = new Date(record.occurredAt);
    const key = Number.isNaN(date.getTime()) ? 'unknown' : `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
    const title = Number.isNaN(date.getTime()) ? 'Date unavailable' : formatter.format(date);
    const section = sections.get(key) ?? { title, data: [] };
    section.data.push(record);
    sections.set(key, section);
  }
  return [...sections.values()];
}

export function historyLoadError(error: unknown): { readonly title: string; readonly message: string; readonly canRetry: boolean } {
  const status = typeof error === 'object' && error !== null && 'status' in error ? error.status : undefined;
  if (status === 401 || status === 403) return { title: t('mobile.AssetHistoryPresentation.historyUnavailable'), message: t('mobile.AssetHistoryPresentation.youDoNotHaveAccessToThisItemS'), canRetry: false };
  if (status === 404) return { title: t('mobile.AssetHistoryPresentation.historyUnavailable'), message: t('mobile.AssetHistoryPresentation.thisItemIsUnavailableOrYouNoLongerHave'), canRetry: false };
  return { title: t('mobile.AssetHistoryPresentation.couldNotLoadHistory'), message: t('mobile.AssetHistoryPresentation.historyCouldNotBeLoadedTryAgain'), canRetry: true };
}

export function historyRevertConfirmation(entry: AssetActivityEntry): {
  readonly title: string;
  readonly message: string;
  readonly confirmLabel: string;
} {
  const actionOutcome = historicalActionOutcome(entry.action);
  if (actionOutcome) {
    return {
      title: t('mobile.AssetHistoryPresentation.revertThisChange'),
      message: `${actionOutcome} Other changes to the item will stay as they are.`,
      confirmLabel: t('mobile.AssetHistoryPresentation.revertChange')
    };
  }
  const fields = [...new Set(entry.changes.map((change) => userFieldLabel(change.field)))];
  const changeDescription = fields.length === 0
    ? 'this change'
    : fields.length === 1
      ? `the ${fields[0]} change`
      : `the ${fields.slice(0, -1).join(', ')} and ${fields.at(-1)} changes`;
  return {
    title: t('mobile.AssetHistoryPresentation.revertThisChange'),
    message: `This will reverse ${changeDescription} from this entry. Other changes to the item will stay as they are.`,
    confirmLabel: t('mobile.AssetHistoryPresentation.revertChange')
  };
}

function historicalActionOutcome(action: string): string | undefined {
  switch (action) {
    case 'asset.created': return 'This item will be archived.';
    case 'asset.moved': return 'The item’s previous location will be restored.';
    case 'asset.archived': return 'This item will be restored.';
    case 'asset.restored': return 'This item will be archived.';
    case 'asset.checked_out': return 'The checkout will be canceled.';
    case 'asset.returned': return 'The item will be checked out again.';
    default: return undefined;
  }
}

export function historyRevertFailure(error: unknown): { readonly title: string; readonly message: string; readonly isTerminal: boolean } {
  const status = typeof error === 'object' && error !== null && 'status' in error ? error.status : undefined;
  if (status === 401 || status === 403) {
    return {
      title: t('mobile.AssetHistoryPresentation.revertUnavailable'),
      message: t('mobile.AssetHistoryPresentation.youNoLongerHavePermissionToRevertThisChange'),
      isTerminal: true
    };
  }
  if (status === 404) {
    return {
      title: t('mobile.AssetHistoryPresentation.changeCanTBeReverted'),
      message: t('mobile.AssetHistoryPresentation.thisChangeIsNoLongerAvailable'),
      isTerminal: true
    };
  }
  if (status === 409) {
    return {
      title: t('mobile.AssetHistoryPresentation.changeCanTBeReverted'),
      message: t('mobile.AssetHistoryPresentation.thisItemChangedAfterwardSoThisChangeCanT'),
      isTerminal: true
    };
  }
  return {
    title: t('mobile.AssetHistoryPresentation.couldNotRevertChange'),
    message: t('mobile.AssetHistoryPresentation.theChangeCouldNotBeRevertedTryAgain'),
    isTerminal: false
  };
}

function userFieldLabel(field: AssetActivityEntry['changes'][number]['field']): string {
  switch (field) {
    case 'title': return 'name';
    case 'description': return 'description';
    case 'tags': return 'tag';
    case 'parent': return 'location';
    case 'lifecycle_state': return 'status';
    case 'checkout_state': return 'checkout';
  }
}

export function technicalDetailRows(entry: AssetActivityEntry): readonly { readonly label: string; readonly value: string }[] {
  return [
    { label: t('mobile.AssetHistoryPresentation.auditRecord'), value: entry.id },
    { label: t('mobile.AssetHistoryPresentation.action'), value: entry.action },
    { label: t('mobile.AssetHistoryPresentation.source'), value: entry.source },
    ...(entry.requestId ? [{ label: t('mobile.AssetHistoryPresentation.request'), value: entry.requestId }] : []),
    ...Object.entries(entry.technical).map(([label, value]) => ({ label, value }))
  ];
}
