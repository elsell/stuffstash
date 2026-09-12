import type { ExpirationFilter, ExpirationMode } from '../../application/expiration/ExpirationRepository';
import { isAssetExpiration } from '../../domain/assets/AssetExpiration';
type Param = string | string[] | undefined;
export type ExpirationRouteParams = { kind?: Param; checkoutState?: Param; tenantId?: Param; inventoryId?: Param; mode?: Param; query?: Param; typeId?: Param; tagIds?: Param; locationId?: Param; fromDate?: Param; throughDate?: Param };
export function parseExpirationRoute(params: ExpirationRouteParams) {
 const scalar = (value: Param) => typeof value === 'string' ? value : undefined;
 const mode = scalar(params.mode);
 const kind = scalar(params.kind); const checkout = scalar(params.checkoutState);
 const date = (value:Param) => {const text=scalar(value);return text && isAssetExpiration({date:text,precision:'day'}) ? text : undefined;};
 return { tenantId: scalar(params.tenantId), inventoryId: scalar(params.inventoryId), filter: { kind:kind === 'item' || kind === 'container' || kind === 'location' ? kind : undefined,checkoutState:checkout === 'checked_out' || checkout === 'available' ? checkout : undefined,mode: mode === 'soon' || mode === 'expired' ? mode : 'all' as ExpirationMode, query: scalar(params.query), typeId: scalar(params.typeId), tagIds: Array.isArray(params.tagIds) ? params.tagIds : params.tagIds ? [params.tagIds] : [], locationId: scalar(params.locationId), fromDate: date(params.fromDate), throughDate: date(params.throughDate) } satisfies ExpirationFilter };
}
export function expirationRouteParams(tenantId: string, inventoryId: string, filter: ExpirationFilter) {
 return { tenantId, inventoryId, kind:filter.kind ?? '',checkoutState:filter.checkoutState ?? '',mode: filter.mode, query: filter.query ?? '', typeId: filter.typeId ?? '', tagIds: [...(filter.tagIds ?? [])], locationId: filter.locationId ?? '', fromDate: filter.fromDate ?? '', throughDate: filter.throughDate ?? '' };
}
