import type { ImportJob } from '$lib/domain/inventory';

export type CountCell = { value: number; label: string; muted?: boolean };

export function statusLabel(job: ImportJob): string {
  switch (job.status) {
    case 'previewed':
      return 'Ready';
    case 'running':
      return 'In progress';
    case 'succeeded':
      return 'Completed';
    case 'failed':
      return 'Failed';
    case 'cancel_requested':
      return 'Cancelling';
    case 'cancelled_kept':
      return 'Kept';
    case 'cancelled_discarded':
      return 'Discarded';
    case 'discard_failed':
      return 'Discard failed';
    default:
      return String(job.status).replaceAll('_', ' ');
  }
}

export function jobNeedsAttention(job: ImportJob): boolean {
  return job.status === 'failed' || job.status === 'discard_failed' || job.counts.errors > 0 || job.counts.warnings > 0;
}

export function statusSentence(job: ImportJob): string {
  switch (job.status) {
    case 'previewed':
      return 'Ready for your review.';
    case 'running':
      return 'Import is running in the background.';
    case 'cancel_requested':
      return 'Cancellation is waiting for a safe stopping point.';
    case 'succeeded':
      return job.counts.warnings > 0 ? 'Completed with warnings.' : 'Completed successfully.';
    case 'failed':
      return 'Import failed before it could finish.';
    case 'cancelled_kept':
      return 'Cancelled. Partial progress was kept.';
    case 'cancelled_discarded':
      return 'Cancelled. Partial progress was discarded.';
    case 'discard_failed':
      return 'Cancellation cleanup needs attention.';
    default:
      return statusLabel(job);
  }
}

export function phaseLabel(job: ImportJob): string {
  const phase = job.progress.phase || job.status;
  switch (phase) {
    case 'ready':
      return 'Ready';
    case 'reading_source':
      return 'Reading source';
    case 'creating_fields':
    case 'fields':
      return 'Creating fields';
    case 'creating_locations':
    case 'locations':
      return 'Creating locations';
    case 'creating_assets':
    case 'assets':
      return 'Creating assets';
    case 'importing_attachments':
    case 'attachments':
      return 'Importing photos and files';
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
  return 'Total not known yet';
}

export function progressBarLabel(job: ImportJob): string {
  if (progressKnown(job)) {
    return `Import progress ${progressPercent(job)} percent`;
  }
  return `Import progress for ${phaseLabel(job)}; total not known yet`;
}

export function progressBarStyle(job: ImportJob): string | undefined {
  return progressKnown(job) ? `width: ${progressPercent(job)}%` : undefined;
}

export function progressTimeline(job: ImportJob): ImportJob['progressHistory'] {
  return job.progressHistory.length > 0 ? job.progressHistory : [job.progress];
}

export function sourceDescription(job: ImportJob): string {
  const parts = [job.source.type === 'legacy_homebox_csv' ? 'CSV upload' : job.source.baseUrl || 'Live connection'];
  if (job.source.version) parts.push(job.source.version);
  if (job.source.imageImport === 'enabled') parts.push('photos on');
  if (job.source.imageImport === 'disabled') parts.push('photos off');
  if (job.source.imageImport === 'unavailable') parts.push('photos unavailable');
  return parts.join(' · ');
}

export function sourceOptionsSummary(job: ImportJob): string[] {
  if (job.source.type === 'legacy_homebox_csv') {
    return ['CSV import', 'Photos unavailable from CSV'];
  }
  const options = [job.source.imageImport === 'disabled' ? 'Photos off' : 'Photos on'];
  if (job.source.allowPrivateNetwork) {
    options.push('Private-network URLs allowed');
  }
  if (job.source.allowInsecureTLS) {
    options.push('Self-signed TLS allowed');
  }
  if (!job.source.allowPrivateNetwork && !job.source.allowInsecureTLS) {
    options.push('Standard network protections');
  }
  return options;
}

export function actorSummary(job: ImportJob): string {
  if (!job.actorId) return '';
  return `Prepared by ${job.actorId}`;
}

export function historyCountSummary(job: ImportJob): string {
  if (job.status === 'previewed') {
    return countParts([
      [job.counts.locations, 'location', 'locations'],
      [job.counts.assets, 'asset', 'assets'],
      [job.counts.attachments, 'photo/file', 'photos/files']
    ]);
  }
  if (job.status === 'running' || job.status === 'cancel_requested') {
    return progressSummary(job);
  }
  if (job.status === 'cancelled_discarded') {
    return countParts([
      [job.counts.recordsDiscarded, 'record discarded', 'records discarded'],
      [job.counts.sourceLinksDiscarded, 'source link removed', 'source links removed']
    ]);
  }
  return countParts([
    [job.counts.fieldsCreated, 'field created', 'fields created'],
    [job.counts.locationsCreated, 'location created', 'locations created'],
    [job.counts.assetsCreated, 'asset created', 'assets created'],
    [job.counts.attachmentsCreated, 'photo/file imported', 'photos/files imported'],
    [job.counts.assetsSkipped + job.counts.attachmentsSkipped, 'skipped', 'skipped']
  ]);
}

export function countParts(parts: Array<[number, string, string]>): string {
  const labels = parts.filter(([count]) => count > 0).map(([count, singular, plural]) => `${count} ${count === 1 ? singular : plural}`);
  return labels.length > 0 ? labels.join(' · ') : 'No records changed';
}

export function previewCountCells(job: ImportJob): CountCell[] {
  return [
    countCell(job.counts.fields, 'field', 'fields'),
    countCell(job.counts.locations, 'location', 'locations'),
    countCell(job.counts.assets, 'asset', 'assets'),
    countCell(job.counts.attachments, 'photo/file', 'photos/files'),
    countCell(job.counts.fieldsExisting + job.counts.assetsSkipped + job.counts.attachmentsSkipped, 'duplicate/skip', 'duplicates/skips', true),
    countCell(job.counts.warnings, 'warning', 'warnings', true),
    countCell(job.counts.errors, 'blocking issue', 'blocking issues', job.counts.errors === 0)
  ];
}

export function resultCountCells(job: ImportJob): CountCell[] {
  return [
    countCell(job.counts.fieldsCreated, 'field created', 'fields created'),
    countCell(job.counts.fieldsExisting, 'field reused', 'fields reused', true),
    countCell(job.counts.locationsCreated, 'location created', 'locations created'),
    countCell(job.counts.assetsCreated, 'asset created', 'assets created'),
    countCell(job.counts.attachmentsCreated, 'photo/file imported', 'photos/files imported'),
    countCell(job.counts.assetsSkipped, 'asset skipped', 'assets skipped', true),
    countCell(job.counts.attachmentsSkipped, 'photo/file skipped', 'photos/files skipped', true),
    countCell(job.counts.warnings, 'warning', 'warnings', true),
    countCell(job.counts.errors, 'blocking issue', 'blocking issues', job.counts.errors === 0),
    countCell(job.counts.recordsDiscarded, 'record discarded', 'records discarded', job.counts.recordsDiscarded === 0)
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
  const visible = cells.filter((cell) => cell.value > 0 || cell.label.startsWith('blocking'));
  return visible.length > 0 ? visible : cells.slice(0, 4);
}

export function visiblePreviewMessages(job: ImportJob): ImportJob['messages'] {
  return job.preview.messages.length > 0 ? job.preview.messages : job.messages.slice(0, 8);
}

export function previewReadinessTitle(job: ImportJob, previewStale: boolean): string {
  if (previewStale) return 'Preview needs to be refreshed';
  if (job.counts.errors > 0) return 'Fix blocking issues before importing';
  return 'Ready to start';
}

export function previewReadinessDescription(job: ImportJob, previewStale: boolean): string {
  if (previewStale) return 'The source settings changed after this preview. Confirm the source again before starting.';
  if (job.counts.errors > 0) return 'Nothing has been saved. Review the blocking messages below and preview again after fixing the source.';
  if (job.counts.warnings > 0) return 'Nothing has been saved. Warnings are shown below so you can decide whether to continue.';
  return 'Nothing has been saved. Start the import when this plan looks right.';
}

export function previewReadinessBadge(job: ImportJob, previewStale: boolean): string {
  if (previewStale) return 'Re-preview required';
  if (job.counts.errors > 0) return `${job.counts.errors} blocking`;
  if (job.counts.warnings > 0) return `${job.counts.warnings} warnings`;
  return 'Ready';
}

export function jobTimeLabel(label: string, value?: string): string {
  if (!value) return '';
  return `${label} ${shortDateTime(value)}`;
}

export function shortDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  }).format(date);
}

export function statusVariant(job: ImportJob): 'default' | 'secondary' | 'destructive' {
  if (job.status === 'failed' || job.status === 'discard_failed') return 'destructive';
  if (job.status === 'running' || job.status === 'succeeded') return 'default';
  return 'secondary';
}

export function resourceLabel(resource: ImportJob['resources'][number]): string {
  if (resource.displayName?.trim()) return resource.displayName.trim();
  if (resource.resourceType === 'attachment') return 'Imported photo/file';
  if (resource.sourceEntityType === 'asset' && resource.sourceEntityId.startsWith('location:')) return 'Imported location';
  return 'Imported asset';
}

export function resourceDiagnosticLabel(resource: ImportJob['resources'][number]): string {
  return `Source ${resource.sourceEntityType}: ${resource.sourceEntityId}`;
}

export function sourceSnapshotDescription(job: ImportJob): string {
  if (job.source.type === 'legacy_homebox_csv') return 'CSV snapshot checked for this preview.';
  return 'Homebox source checked for this preview.';
}

export function previewLocationContext(item: { parentSourceId?: string; archived: boolean }): string {
  return `${item.parentSourceId ? 'inside another imported record' : 'top level'}${item.archived ? ' · archived source skipped' : ''}`;
}

export function previewAssetContext(item: { kind: string; parentSourceId?: string; archived: boolean }): string {
  return `${item.kind}${item.parentSourceId ? ' · inside another imported record' : ''}${item.archived ? ' · archived source skipped' : ''}`;
}

export function fileSizeLabel(bytes: number): string {
  if (bytes <= 0) return 'size unknown';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function countCell(value: number, singular: string, plural: string, muted = false): CountCell {
  return { value, label: value === 1 ? singular : plural, muted };
}
