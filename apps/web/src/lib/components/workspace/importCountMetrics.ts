import { t } from '$lib/presentation/localization';

/** Stable metric identities; translated labels never determine behavior. */
export type ImportCountMetric = 'asset' | 'assetCreated' | 'assetSaved' | 'assetSkipped' | 'blockingIssue' | 'duplicateSkip' | 'field' | 'fieldCreated' | 'fieldReused' | 'location' | 'locationCreated' | 'locationSaved' | 'photoFile' | 'photoFileImported' | 'photoFileSaved' | 'photoFileSkipped' | 'plannedAsset' | 'plannedLocation' | 'plannedPhotoFile' | 'recordDiscarded' | 'skipped' | 'sourceLinkRemoved' | 'warning';

export function importCountAction(metric: ImportCountMetric): 'issues' | 'records' | undefined {
  if (metric === 'blockingIssue' || metric === 'warning') return 'issues';
  if (['locationCreated', 'assetCreated', 'photoFileImported', 'locationSaved', 'assetSaved', 'photoFileSaved'].includes(metric)) return 'records';
  return undefined;
}

export function importCountIcon(metric: ImportCountMetric): 'issue' | 'location' | 'attachment' | 'asset' | 'skipped' | 'other' {
  if (metric === 'blockingIssue' || metric === 'warning') return 'issue';
  if (['location', 'locationCreated', 'locationSaved', 'plannedLocation'].includes(metric)) return 'location';
  if (['photoFile', 'photoFileImported', 'photoFileSaved', 'photoFileSkipped', 'plannedPhotoFile'].includes(metric)) return 'attachment';
  if (['asset', 'assetCreated', 'assetSaved', 'assetSkipped', 'plannedAsset', 'recordDiscarded'].includes(metric)) return 'asset';
  if (metric === 'skipped' || metric === 'duplicateSkip') return 'skipped';
  return 'other';
}

const actionMessages: Partial<Record<ImportCountMetric, Parameters<typeof t>[0]>> = {
  blockingIssue: 'import.action.blockingIssue',
  warning: 'import.action.warning',
  locationCreated: 'import.action.locationCreated',
  assetCreated: 'import.action.assetCreated',
  photoFileImported: 'import.action.photoFileImported',
  locationSaved: 'import.action.locationSaved',
  assetSaved: 'import.action.assetSaved',
  photoFileSaved: 'import.action.photoFileSaved',
};

export function importCountActionLabel(metric: ImportCountMetric, count: number): string | undefined {
  const key = actionMessages[metric];
  return key ? t(key, { count }) : undefined;
}
