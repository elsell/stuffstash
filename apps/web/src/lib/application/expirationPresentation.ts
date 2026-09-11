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
 if (!context.trackingEnabled) return 'Expiration tracking disabled';
 if (context.state === 'upcoming') return 'Expiring soon';
 if (context.state === 'expired') return 'Expired';
 return undefined;
}
