import { t } from '../../presentation/localization';
import type {AssetExpirationContext} from '../../domain/assets/AssetSummary';
import type { AssetExpiration } from '../../domain/assets/AssetSummary';
import { isAssetExpiration } from '../../domain/assets/AssetExpiration';

export function formatAssetExpiration(value: AssetExpiration, locale?: string): string {
  const originalDate = value.date;
  if (!isAssetExpiration(value)) return originalDate;
  const calendarDate = value.precision === 'month' ? `${value.date}-01` : value.date;
  const monthOnly = value.precision === 'month';
  const formatted = new Intl.DateTimeFormat(locale, {
    timeZone: 'UTC', year: 'numeric', month: 'long',
    ...(monthOnly ? { calendar: 'gregory' } : { day: 'numeric' as const })
  }).format(new Date(`${calendarDate}T00:00:00Z`));
  return monthOnly && hasAlternativeCalendar(locale) ? t('mobile.ExpirationPresentation.gregorian', { formatted: String(formatted) }) : formatted;
}

export function formatExpirationChange(value?: AssetExpiration, cleared?: boolean): string | undefined {
  return cleared ? t('mobile.ExpirationPresentation.removeExpirationDate') : value ? t('mobile.ExpirationPresentation.expires', { value: String(formatAssetExpiration(value)) }) : undefined;
}

export function expirationStatusLabel(context?: AssetExpirationContext): string | undefined {
 if (!context) return undefined;
 if (!context.trackingEnabled) return t('mobile.ExpirationPresentation.expirationTrackingDisabled');
 if (context.state === 'upcoming') return t('mobile.ExpirationPresentation.expiringSoon');
 if (context.state === 'expired') return t('mobile.ExpirationPresentation.expired');
 return undefined;
}

export function expirationDateLabel(value: AssetExpiration, context?: AssetExpirationContext, now = new Date(), locale?: string): string {
 const formatted = formatAssetExpiration(value, locale);
 const label = (state: 'default' | 'disabled' | 'today' | 'upcoming' | 'expired') =>
   t(`expiration.${state}.${value.precision}`, { date: formatted });
 if (!context) return label('default');
 if (!context.trackingEnabled) return label('disabled');
 const parts = new Intl.DateTimeFormat('en-US', { timeZone: context.timezone, year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(now);
 const part = (name: string) => parts.find(part => part.type === name)?.value ?? '';
 const today = `${part('year')}-${part('month')}-${part('day')}`;
 const last = value.precision === 'month' ? new Date(Date.UTC(Number(value.date.slice(0,4)), Number(value.date.slice(5,7)), 0)).toISOString().slice(0,10) : value.date;
 return label(last === today ? 'today' : context.state === 'current' ? 'default' : context.state);
}

/** Month values remain Gregorian periods, even when the device uses another calendar. */
export function expirationMonthOptions(locale?: string): readonly { value: string; label: string }[] {
  const formatter = new Intl.DateTimeFormat(locale, { calendar: 'gregory', month: 'long', timeZone: 'UTC' });
  return Array.from({ length: 12 }, (_, index) => ({
    value: String(index + 1),
    label: formatter.format(new Date(Date.UTC(2000, index, 1)))
  }));
}

export function expirationMonthCalendarNotice(locale?: string): string | undefined {
  return hasAlternativeCalendar(locale) ? t('mobile.ExpirationPresentation.monthAndYearUseTheGregorianCalendar') : undefined;
}

function hasAlternativeCalendar(locale?: string): boolean {
  return new Intl.DateTimeFormat(locale).resolvedOptions().calendar !== 'gregory';
}
