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

export function expirationDateLabel(value: AssetExpiration, context?: AssetExpirationContext, now = new Date(), locale?: string): string {
 const formatted = formatAssetExpiration(value, locale);
 const precision = value.precision === 'month' ? `${formatted} (end of month)` : formatted;
 if (!context) return `Expiration: ${precision}`;
 if (!context.trackingEnabled) return `Expiration tracking disabled: ${precision}`;
 const parts = new Intl.DateTimeFormat('en-US', { timeZone: context.timezone, year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(now);
 const part = (name: string) => parts.find(part => part.type === name)?.value ?? '';
 const today = `${part('year')}-${part('month')}-${part('day')}`;
 const last = value.precision === 'month' ? new Date(Date.UTC(Number(value.date.slice(0,4)), Number(value.date.slice(5,7)), 0)).toISOString().slice(0,10) : value.date;
 return `${last === today ? 'Expires today' : expirationStatusLabel(context) ?? 'Expiration'}: ${precision}`;
}
