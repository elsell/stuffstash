import { t } from '$lib/presentation/localization';
import type {AssetExpirationContext} from '$lib/domain/inventory';
import type {AssetExpiration} from '$lib/domain/inventory';
import {validExpirationInput} from '$lib/domain/expiration';
export function formatAssetExpiration(value:AssetExpiration,locale?:string):string {
 if(!value.date || !validExpirationInput(value.date,value.precision))return value.date;
 const date=value.precision==='month'?`${value.date}-01`:value.date;
 return new Intl.DateTimeFormat(locale,{year:'numeric',month:'long',...(value.precision==='day'?{day:'numeric' as const}:{}),timeZone:'UTC'}).format(new Date(`${date}T00:00:00Z`));
}

export function expirationStatusLabel(context?: AssetExpirationContext): string | undefined {
 if (!context) return undefined;
 if (!context.trackingEnabled) return t('web.expirationPresentation.expirationTrackingDisabled');
 if (context.state === 'upcoming') return t('web.expirationPresentation.expiringSoon');
 if (context.state === 'expired') return t('web.expirationPresentation.expired');
 return undefined;
}

export function expirationDateLabel(value: AssetExpiration, context?: AssetExpirationContext, now = new Date(), locale?: string): string {
 const formatted = formatAssetExpiration(value, locale);
 const label = (state: 'default' | 'disabled' | 'today' | 'upcoming' | 'expired') => t(`web.expiration.date.${state}.${value.precision}`, {date: formatted});
 if (!context) return label('default');
 if (!context.trackingEnabled) return label('disabled');
 const parts = new Intl.DateTimeFormat('en-US', { timeZone: context.timezone, year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(now);
 const part = (name: string) => parts.find(part => part.type === name)?.value ?? '';
 const today = `${part('year')}-${part('month')}-${part('day')}`;
 const last = value.precision === 'month' ? new Date(Date.UTC(Number(value.date.slice(0,4)), Number(value.date.slice(5,7)), 0)).toISOString().slice(0,10) : value.date;
 return label(last === today ? 'today' : context.state === 'upcoming' || context.state === 'expired' ? context.state : 'default');
}
