import type { ImportCountMetric } from './importCountMetrics';
import { localization, t } from '$lib/presentation/localization';
import type { ImportJob, ImportMessage, Principal } from '$lib/domain/inventory';

export type CountCell = {
  metric: ImportCountMetric;
  value: number;
  label: string;
  muted?: boolean;
  tone?: 'default' | 'success' | 'warning' | 'action' | 'muted';
  actionLabel?: string;
};
export type ImportIssueTone = 'none' | 'warning' | 'action';

export function statusLabel(job: ImportJob): string {
  switch (job.status) {
    case 'previewed':
      return t('web.importWorkspacePresentation.ready');
    case 'running':
      return t('web.importWorkspacePresentation.inProgress');
    case 'succeeded':
      return t('web.importWorkspacePresentation.completed');
    case 'failed':
      return t('web.importWorkspacePresentation.failed');
    case 'cancel_requested':
      return t('web.importWorkspacePresentation.cancelling');
    case 'cancelled_kept':
      return t('web.importWorkspacePresentation.kept');
    case 'cancelled_discarded':
      return t('web.importWorkspacePresentation.discarded');
    case 'discard_failed':
      return t('web.importWorkspacePresentation.discardFailed');
    default:
      return String(job.status).replaceAll('_', ' ');
  }
}

export function jobNeedsAttention(job: ImportJob): boolean {
  return job.status === 'failed' || job.status === 'discard_failed' || job.counts.errors > 0;
}

export function jobHasWarnings(job: ImportJob): boolean {
  return !jobNeedsAttention(job) && job.counts.warnings > 0;
}

export function importIssueTone(job: ImportJob): ImportIssueTone {
  const messages = allJobMessages(job);
  if (jobNeedsAttention(job) || messages.some((message) => message.severity === 'error')) {
    return 'action';
  }
  if (jobHasWarnings(job) || messages.some((message) => message.severity === 'warning')) {
    return 'warning';
  }
  return 'none';
}

export function attentionSummary(job: ImportJob): string {
  if (job.status === 'discard_failed') return t('web.importWorkspacePresentation.cancellationCleanupNeedsReview');
  if (job.status === 'failed') return t('web.importWorkspacePresentation.importFailedBeforeItFinished');
  if (job.counts.errors === 0 && job.counts.warnings === 0) return t('web.importWorkspacePresentation.noIssues');
  return countParts([
    [job.counts.errors, 'blockingIssue'],
    [job.counts.warnings, 'warning']
  ]);
}

export function statusSentence(job: ImportJob): string {
  switch (job.status) {
    case 'previewed':
      return t('web.importWorkspacePresentation.readyForYourReview');
    case 'running':
      return t('web.importWorkspacePresentation.importIsRunningInTheBackground');
    case 'cancel_requested':
      return t('web.importWorkspacePresentation.cancellationIsWaitingForASafeStoppingPoint');
    case 'succeeded':
      return job.counts.warnings > 0 ? t('web.importWorkspacePresentation.completedWithWarnings') : t('web.importWorkspacePresentation.completedSuccessfully');
    case 'failed': {
      const created = job.counts.fieldsCreated + (job.counts.tagsCreated ?? 0) + job.counts.locationsCreated + job.counts.assetsCreated + job.counts.attachmentsCreated;
      if (allJobMessages(job).some((message) => message.code === 'attachment-session-unavailable' || message.code === 'attachment-storage-unavailable')) {
        return created > 0
          ? t('web.importWorkspacePresentation.imageImportDidNotStartEarlierRecordsWereKept', { created: String(created) })
          : t('web.importWorkspacePresentation.imageImportDidNotStart');
      }
      return created > 0 ? t('web.importWorkspacePresentation.importStoppedRecordsWereKept', { created: String(created) }) : t('web.importWorkspacePresentation.importFailedBeforeItCouldFinish');
    }
    case 'cancelled_kept':
      return t('web.importWorkspacePresentation.cancelledPartialProgressWasKept');
    case 'cancelled_discarded':
      return t('web.importWorkspacePresentation.cancelledPartialProgressWasDiscarded');
    case 'discard_failed':
      return t('web.importWorkspacePresentation.cancellationCleanupNeedsAttention');
    default:
      return statusLabel(job);
  }
}

export function phaseLabel(job: ImportJob): string {
  const phase = job.progress.phase || job.status;
  switch (phase) {
    case 'ready':
      return t('web.importWorkspacePresentation.ready');
    case 'reading_source':
      return t('web.importWorkspacePresentation.readingSource');
    case 'creating_fields':
    case 'fields':
      return t('web.importWorkspacePresentation.creatingFields');
    case 'creating_locations':
    case 'locations':
      return t('web.importWorkspacePresentation.creatingLocations');
    case 'creating_assets':
    case 'assets':
      return t('web.importWorkspacePresentation.creatingAssets');
    case 'importing_attachments':
    case 'attachments':
      return t('web.importWorkspacePresentation.importingPhotosAndFiles');
    case 'terminal':
      return statusLabel(job);
    default:
      return String(phase)
        .replaceAll('_', ' ')
        .replace(/^\w/, (first) => first.toUpperCase());
  }
}

export function isTerminal(job: ImportJob): boolean {
  return ['succeeded', 'failed', 'cancelled_kept', 'cancelled_discarded', 'discard_failed'].includes(job.status);
}

export function canRequestCancellation(job: ImportJob): boolean {
  return job.status === 'running';
}

export function terminalJobMayHaveChangedInventory(job: ImportJob): boolean {
  return ['succeeded', 'failed', 'cancelled_kept', 'discard_failed'].includes(job.status);
}

export function canRemoveJobFromHistory(job: ImportJob): boolean {
  return ['succeeded', 'failed', 'cancelled_kept', 'cancelled_discarded'].includes(job.status);
}

export function progressKnown(job: ImportJob): boolean {
  return job.progress.total > 0;
}

export function jobTotal(job: ImportJob): number {
  return job.progress.total;
}

export function jobDone(job: ImportJob): number {
  return Math.min(job.progress.done || 0, jobTotal(job));
}

export function progressPercent(job: ImportJob): number {
  const total = jobTotal(job);
  if (total <= 0) return 0;
  return Math.max(0, Math.min(100, Math.round((jobDone(job) / total) * 100)));
}

export function progressSummary(job: ImportJob): string {
  if (progressKnown(job)) {
    return `${jobDone(job)} / ${jobTotal(job)}`;
  }
  if (isTerminal(job)) {
    return statusLabel(job);
  }
  return t('web.importWorkspacePresentation.totalNotKnownYet');
}

export function progressBarLabel(job: ImportJob): string {
  if (progressKnown(job)) {
    return t('web.importWorkspacePresentation.importProgressPercent', { value: String(progressPercent(job)) });
  }
  return t('web.importWorkspacePresentation.importProgressForTotalNotKnownYet', { value: String(phaseLabel(job)) });
}

export function progressBarStyle(job: ImportJob): string | undefined {
  return progressKnown(job) ? t('web.importWorkspacePresentation.width', { value: String(progressPercent(job)) }) : undefined;
}

export function progressTimeline(job: ImportJob): ImportJob['progressHistory'] {
  return job.progressHistory.length > 0 ? job.progressHistory : [job.progress];
}

export function sourceDescription(job: ImportJob): string {
  const parts = [job.source.type === 'legacy_homebox_csv' ? 'CSV upload' : compactSourceURL(job.source.baseUrl) || 'Homebox'];
  if (job.source.version) parts.push(job.source.version);
  return parts.join(' · ');
}

export function sourceOptionsSummary(job: ImportJob): string[] {
  if (job.source.type === 'legacy_homebox_csv') {
    return ['CSV file', 'Photos are not included in Homebox CSV exports'];
  }
  const options = ['Connected directly to Homebox'];
  if (job.source.imageImport === 'disabled') {
    options.push('Photo import disabled');
  }
  if (job.source.allowPrivateNetwork) {
    options.push('Allowed local/private network address');
  }
  if (job.source.allowInsecureTLS) {
    options.push('Allowed self-signed certificate');
  }
  return options;
}

export function actorSummary(job: ImportJob, currentPrincipal?: Principal): string {
  if (!job.actorId) return '';
  if (job.actor?.email) return t('web.importWorkspacePresentation.preparedBy', { email: String(job.actor.email) });
  if (currentPrincipal?.id === job.actorId && currentPrincipal.email) return t('web.importWorkspacePresentation.preparedBy', { email: String(currentPrincipal.email) });
  if (!job.actorId.includes('@') && job.actorId.length > 24) return t('web.importWorkspacePresentation.preparedBy2', { value: String(compactIdentifier(job.actorId)) });
  return t('web.importWorkspacePresentation.preparedBy3', { actorId: String(job.actorId) });
}

export function compactSourceURL(value?: string): string {
  if (!value) return '';
  try {
    const parsed = new URL(value);
    return parsed.host || value;
  } catch {
    return value.replace(/^https?:\/\//, '').replace(/\/api\/v\d+\/?$/, '');
  }
}

export function compactIdentifier(value: string): string {
  if (value.length <= 24) return value;
  return `${value.slice(0, 12)}...${value.slice(-9)}`;
}

export function historyCountSummary(job: ImportJob): string {
  if (job.status === 'previewed') {
    return countParts([
      [job.counts.locations, 'location'],
      [job.counts.assets, 'asset'],
      [job.counts.attachments, 'photoFile']
    ]);
  }
  if (job.status === 'running' || job.status === 'cancel_requested') {
    return progressSummary(job);
  }
  if (job.status === 'cancelled_discarded') {
    return countParts([
      [job.counts.recordsDiscarded, 'recordDiscarded'],
      [job.counts.sourceLinksDiscarded, 'sourceLinkRemoved']
    ]);
  }
  return countParts([
    [job.counts.fieldsCreated, 'fieldCreated'],
    [job.counts.locationsCreated, 'locationCreated'],
    [job.counts.assetsCreated, 'assetCreated'],
    [job.counts.attachmentsCreated, 'photoFileImported'],
    [job.counts.assetsSkipped + job.counts.attachmentsSkipped, 'skipped']
  ]);
}

export function issueCountSummary(job: ImportJob): string {
  const errorCount = reportedErrorCount(job);
  const warningCount = reportedWarningCount(job);
  if (errorCount === 0 && warningCount === 0) return t('web.importWorkspacePresentation.noIssues');
  return countParts([
    [errorCount, 'blockingIssue'],
    [warningCount, 'warning']
  ]);
}

export function issueTotalCount(job: ImportJob): number {
  return reportedErrorCount(job) + reportedWarningCount(job);
}

export function reportedErrorCount(job: ImportJob): number {
  const messageCount = uniqueImportMessages(allJobMessages(job)).filter((message) => message.severity === 'error').length;
  return Math.max(job.counts.errors, messageCount);
}

export function reportedWarningCount(job: ImportJob): number {
  const messageCount = uniqueImportMessages(allJobMessages(job)).filter((message) => message.severity === 'warning').length;
  return Math.max(job.counts.warnings, messageCount);
}

export function changedRecordSummary(job: ImportJob): string {
  if (job.status === 'previewed') {
    return countParts([
      [job.counts.locations, 'plannedLocation'],
      [job.counts.assets, 'plannedAsset'],
      [job.counts.attachments, 'plannedPhotoFile']
    ]);
  }
  if (job.status === 'cancelled_discarded') {
    return countParts([
      [job.counts.recordsDiscarded, 'recordDiscarded'],
      [job.counts.sourceLinksDiscarded, 'sourceLinkRemoved']
    ]);
  }
  return countParts([
    [job.counts.locationsCreated, 'locationSaved'],
    [job.counts.assetsCreated, 'assetSaved'],
    [job.counts.attachmentsCreated, 'photoFileSaved']
  ]);
}

export function countParts(parts: Array<[number, ImportCountMetric]>): string {
  const labels = parts.filter(([count]) => count > 0).map(([count, metric]) => t(`import.count.${metric}`, { count }));
  return labels.length > 0 ? labels.join(' · ') : t('web.importWorkspacePresentation.noRecordsChanged');
}

function allJobMessages(job: ImportJob): ImportJob['messages'] {
  return job.messages.length > 0 ? job.messages : job.preview.messages;
}

export function previewCountCells(job: ImportJob): CountCell[] {
  return [
    countCell(job.counts.fields, 'field'),
    countCell(job.counts.locations, 'location'),
    countCell(job.counts.assets, 'asset'),
    countCell(job.counts.attachments, 'photoFile'),
    countCell(job.counts.fieldsExisting + job.counts.assetsSkipped + job.counts.attachmentsSkipped, 'duplicateSkip', true),
    countCell(job.counts.warnings, 'warning', true),
    countCell(job.counts.errors, 'blockingIssue', job.counts.errors === 0)
  ];
}

export function resultCountCells(job: ImportJob): CountCell[] {
  return [
    countCell(job.counts.fieldsCreated, 'fieldCreated'),
    countCell(job.counts.fieldsExisting, 'fieldReused', true),
    countCell(job.counts.locationsCreated, 'locationCreated'),
    countCell(job.counts.assetsCreated, 'assetCreated'),
    countCell(job.counts.attachmentsCreated, 'photoFileImported'),
    countCell(job.counts.assetsSkipped, 'assetSkipped', true),
    countCell(job.counts.attachmentsSkipped, 'photoFileSkipped', true),
    countCell(job.counts.warnings, 'warning', true),
    countCell(job.counts.errors, 'blockingIssue', job.counts.errors === 0),
    countCell(job.counts.recordsDiscarded, 'recordDiscarded', job.counts.recordsDiscarded === 0)
  ];
}

export function visiblePreviewCountCells(job: ImportJob): CountCell[] {
  const cells = previewCountCells(job);
  return [
    ...cells.slice(0, 4),
    ...cells.slice(4, 6).filter((cell) => cell.value > 0),
    cells[6]
  ];
}

export function visibleCountCells(cells: CountCell[]): CountCell[] {
  const visible = cells.filter((cell) => cell.value > 0 || cell.metric === 'blockingIssue');
  return visible.length > 0 ? visible : cells.slice(0, 4);
}

export function visiblePreviewMessages(job: ImportJob): ImportJob['messages'] {
  const messages = job.preview.messages.length > 0 ? job.preview.messages : job.messages;
  return uniqueImportMessages(messages).slice(0, 8);
}

export function uniqueImportMessages(messages: ImportMessage[]): ImportMessage[] {
  const seen = new Set<string>();
  const unique: ImportMessage[] = [];
  for (const message of messages) {
    const key = [
      message.severity,
      message.code,
      message.summary,
      message.detail ?? '',
      message.sourceName ?? '',
      message.sourceId ?? ''
    ].join('\u001f');
    if (seen.has(key)) continue;
    seen.add(key);
    unique.push(message);
  }
  return unique;
}

export function previewReadinessTitle(job: ImportJob, previewStale: boolean): string {
  if (previewStale) return t('web.importWorkspacePresentation.previewNeedsToBeRefreshed');
  if (job.counts.errors > 0) return t('web.importWorkspacePresentation.fixBlockingIssuesBeforeImporting');
  return t('web.importWorkspacePresentation.readyToStart');
}

export function previewReadinessDescription(job: ImportJob, previewStale: boolean): string {
  if (previewStale) return t('web.importWorkspacePresentation.theSourceSettingsChangedAfterThisPreviewConfirmThe');
  if (job.counts.errors > 0) return t('web.importWorkspacePresentation.nothingHasBeenSavedReviewTheBlockingMessagesBelow');
  if (job.counts.warnings > 0) return t('web.importWorkspacePresentation.nothingHasBeenSavedWarningsAreShownBelowSo');
  return t('web.importWorkspacePresentation.nothingHasBeenSavedStartTheImportWhenThis');
}

export function previewReadinessBadge(job: ImportJob, previewStale: boolean): string {
  if (previewStale) return t('web.importWorkspacePresentation.rePreviewRequired');
  if (job.counts.errors > 0) return t('web.importWorkspacePresentation.blocking', { errors: String(job.counts.errors) });
  if (job.counts.warnings > 0) return t('web.importWorkspacePresentation.warnings', { warnings: String(job.counts.warnings) });
  return t('web.importWorkspacePresentation.ready');
}

export function jobTimeLabel(label: string, value?: string): string {
  if (!value) return '';
  return `${label} ${shortDateTime(value)}`;
}

export function shortDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return localization.date(date, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  });
}

export function statusVariant(job: ImportJob): 'default' | 'secondary' | 'destructive' {
  if (job.status === 'failed' || job.status === 'discard_failed') return 'destructive';
  if (job.status === 'running' || job.status === 'succeeded') return 'default';
  return 'secondary';
}

export function resourceLabel(resource: ImportJob['resources'][number]): string {
  if (resource.displayName?.trim()) return resource.displayName.trim();
  if (resource.resourceType === 'attachment') return t('web.importWorkspacePresentation.importedPhotoFile');
  if (resource.sourceEntityType === 'asset' && resource.sourceEntityId.startsWith('location:')) return t('web.importWorkspacePresentation.importedLocation');
  return t('web.importWorkspacePresentation.importedAsset');
}

export function resourceDiagnosticLabel(resource: ImportJob['resources'][number]): string {
  return t('web.importWorkspacePresentation.source', { sourceEntityType: String(resource.sourceEntityType), sourceEntityId: String(resource.sourceEntityId) });
}

export function sourceSnapshotDescription(job: ImportJob): string {
  if (job.source.type === 'legacy_homebox_csv') return t('web.importWorkspacePresentation.cSVSnapshotCheckedForThisPreview');
  return t('web.importWorkspacePresentation.homeboxSourceCheckedForThisPreview');
}

export function previewLocationContext(item: { parentSourceId?: string; archived: boolean }): string {
  return t(`import.preview.location.${item.parentSourceId ? 'nested' : 'root'}.${item.archived ? 'archived' : 'active'}`);
}

export function previewAssetContext(item: { kind: string; parentSourceId?: string; archived: boolean }): string {
  const kind = item.kind === 'item' || item.kind === 'container' || item.kind === 'location' ? item.kind : 'unknown';
  return t(`import.preview.asset.${kind}.${item.parentSourceId ? 'nested' : 'root'}.${item.archived ? 'archived' : 'active'}`, { kind: item.kind });
}

export function fileSizeLabel(bytes: number): string {
  if (bytes <= 0) return t('web.importWorkspacePresentation.sizeUnknown');
  if (bytes < 1024) return `${localization.number(bytes)} B`;
  if (bytes < 1024 * 1024) return t('web.importWorkspacePresentation.kB', { value: localization.number(Math.round(bytes / 1024)) });
  return t('web.importWorkspacePresentation.mB', { value: localization.number(bytes / (1024 * 1024), { minimumFractionDigits: 1, maximumFractionDigits: 1 }) });
}

function countCell(value: number, metric: ImportCountMetric, muted = false): CountCell {
  return { value, metric, label: t(`import.label.${metric}`, { count: value }), muted };
}

export function ledgerChangeSummary(job: ImportJob): string {
  if (!isTerminal(job) || job.status === 'cancelled_discarded') return historyCountSummary(job);
  const skipped = job.counts.assetsSkipped + job.counts.attachmentsSkipped;
  const saved = changedRecordSummary(job);
  if (skipped === 0) return saved;
  const skippedLabel = t('import.history.skipped', { count: skipped });
  if (job.counts.locationsCreated + job.counts.assetsCreated + job.counts.attachmentsCreated === 0) return skippedLabel;
  return `${saved} · ${skippedLabel}`;
  }
