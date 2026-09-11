import type { AssetExpiration } from './AssetSummary';

export function isAssetExpiration(value: unknown): value is AssetExpiration {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false;
  const candidate = value as Partial<AssetExpiration>;
  if (typeof candidate.date !== 'string' || (candidate.precision !== 'day' && candidate.precision !== 'month')) return false;
  const pattern = candidate.precision === 'month' ? /^\d{4}-\d{2}$/ : /^\d{4}-\d{2}-\d{2}$/;
  if (!pattern.test(candidate.date)) return false;
  const [year, month, day] = candidate.date.split('-').map(Number);
  if (year < 1 || month < 1 || month > 12) return false;
  if (candidate.precision === 'month') return true;
  const leap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const days = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
  return day >= 1 && day <= days[month - 1];
}
