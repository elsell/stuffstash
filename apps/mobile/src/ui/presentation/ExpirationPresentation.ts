import type { AssetExpiration } from '../../domain/assets/AssetSummary';
import { isAssetExpiration } from '../../domain/assets/AssetExpiration';

export function formatAssetExpiration(value: AssetExpiration, locale?: string): string {
  const originalDate = value.date;
  if (!isAssetExpiration(value)) return originalDate;
  const calendarDate = value.precision === 'month' ? `${value.date}-01` : value.date;
  return new Intl.DateTimeFormat(locale, { timeZone: 'UTC', year: 'numeric', month: 'long', ...(value.precision === 'day' ? { day: 'numeric' as const } : {}) }).format(new Date(`${calendarDate}T00:00:00Z`));
}
