import type { ExpirationFilter } from '$lib/ports/expirationRepository';
export function parseExpirationFilter(search: URLSearchParams): ExpirationFilter {
 const mode = search.get('expiration');
 const kind=search.get('kind');const checkout=search.get('availability');
 return { kind:kind==='item'||kind==='container'||kind==='location'?kind:undefined,checkoutState:checkout==='available'||checkout==='checked_out'?checkout:undefined,mode: mode === 'soon' || mode === 'expired' ? mode : 'all', query: search.get('q') || undefined, typeId: search.get('type') || undefined, tagIds: [...new Set(search.getAll('tag'))], locationId: search.get('location') || undefined, fromDate: search.get('from') || undefined, throughDate: search.get('through') || undefined };
}
export function writeExpirationFilter(search: URLSearchParams, filter: ExpirationFilter): void {
 search.set('expiration', filter.mode);
 for (const [key, value] of [['kind',filter.kind],['availability',filter.checkoutState],['q', filter.query], ['type', filter.typeId], ['location', filter.locationId], ['from', filter.fromDate], ['through', filter.throughDate]]) if (value) search.set(key!, value);
 for (const id of filter.tagIds ?? []) search.append('tag', id);
}
