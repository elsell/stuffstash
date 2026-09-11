import type {AssetExpirationContext} from '../../domain/assets/AssetSummary';
import type { AssetExpiration } from '../../domain/assets/AssetSummary';
import { isAssetExpiration } from '../../domain/assets/AssetExpiration';

export function formatAssetExpiration(value: AssetExpiration, locale?: string): string {
  const originalDate = value.date;
  if (!isAssetExpiration(value)) return originalDate;
  const calendarDate = value.precision === 'month' ? `${value.date}-01` : value.date;
  return new Intl.DateTimeFormat(locale, { timeZone: 'UTC', year: 'numeric', month: 'long', ...(value.precision === 'day' ? { day: 'numeric' as const } : {}) }).format(new Date(`${calendarDate}T00:00:00Z`));
}

export function formatExpirationChange(value?: AssetExpiration, cleared?: boolean): string | undefined {
  return cleared ? 'Remove expiration date' : value ? `Expires ${formatAssetExpiration(value)}` : undefined;
}

export function expirationStatusLabel(context?: AssetExpirationContext): string | undefined {
 if (!context) return undefined;
 if (!context.trackingEnabled) return 'Expiration tracking disabled';
 if (context.state === 'upcoming') return 'Expiring soon';
 if (context.state === 'expired') return 'Expired';
 return undefined;
}
