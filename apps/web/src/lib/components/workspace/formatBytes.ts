import { localization, t } from '$lib/presentation/localization';

export function formatBytes(sizeBytes: number): string {
  if (sizeBytes < 1024) return t('media.bytes', { size: localization.number(sizeBytes) });
  if (sizeBytes < 1024 * 1024) return t('media.kilobytes', { size: localization.number(Math.round(sizeBytes / 1024)) });
  return t('media.megabytes', { size: localization.number(sizeBytes / 1024 / 1024, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) });
}
