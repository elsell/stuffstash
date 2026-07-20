import type { AssetActivityEntry, AssetActivityRecordViewModel, AssetActivityView } from '../../application/assets/AssetActivityQuery';
import type { NativeActionMenuGroup } from '../components/NativeActionMenu';

export function historyFilterMenuGroups(
  value: AssetActivityView,
  onChange: (view: AssetActivityView) => void
): readonly NativeActionMenuGroup[] {
  return [{
    id: 'history-filter',
    items: [
      {
        id: 'changes',
        label: 'Changes',
        isSelected: value === 'changes',
        onPress: () => {
          if (value !== 'changes') onChange('changes');
        }
      },
      {
        id: 'all',
        label: 'All events',
        isSelected: value === 'all',
        onPress: () => {
          if (value !== 'all') onChange('all');
        }
      }
    ]
  }];
}

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
  if (status === 401 || status === 403) return { title: 'History unavailable', message: 'You do not have access to this item’s History.', canRetry: false };
  if (status === 404) return { title: 'History unavailable', message: 'This item is unavailable or you no longer have access.', canRetry: false };
  return { title: 'Could not load History', message: 'History could not be loaded. Try again.', canRetry: true };
}

export function historyRevertConfirmation(entry: AssetActivityEntry): {
  readonly title: string;
  readonly message: string;
  readonly confirmLabel: string;
} {
  const actionOutcome = historicalActionOutcome(entry.action);
  if (actionOutcome) {
    return {
      title: 'Revert this change?',
      message: `${actionOutcome} Other changes to the item will stay as they are.`,
      confirmLabel: 'Revert Change'
    };
  }
  const fields = [...new Set(entry.changes.map((change) => userFieldLabel(change.field)))];
  const changeDescription = fields.length === 0
    ? 'this change'
    : fields.length === 1
      ? `the ${fields[0]} change`
      : `the ${fields.slice(0, -1).join(', ')} and ${fields.at(-1)} changes`;
  return {
    title: 'Revert this change?',
    message: `This will reverse ${changeDescription} from this entry. Other changes to the item will stay as they are.`,
    confirmLabel: 'Revert Change'
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
      title: 'Revert unavailable',
      message: 'You no longer have permission to revert this change.',
      isTerminal: true
    };
  }
  if (status === 404) {
    return {
      title: 'Change can’t be reverted',
      message: 'This change is no longer available.',
      isTerminal: true
    };
  }
  if (status === 409) {
    return {
      title: 'Change can’t be reverted',
      message: 'This item changed afterward, so this change can’t be safely reverted.',
      isTerminal: true
    };
  }
  return {
    title: 'Could not revert change',
    message: 'The change could not be reverted. Try again.',
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
    { label: 'Audit record', value: entry.id },
    { label: 'Action', value: entry.action },
    { label: 'Source', value: entry.source },
    ...(entry.requestId ? [{ label: 'Request', value: entry.requestId }] : []),
    ...Object.entries(entry.technical).map(([label, value]) => ({ label, value }))
  ];
}
