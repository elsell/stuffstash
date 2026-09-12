import type { AssetSummary } from '../../domain/assets/AssetSummary';
export type ExpirationMode = 'soon' | 'expired' | 'all';
export type ExpirationFilter = {
 readonly kind?: 'item' | 'container' | 'location';
 readonly checkoutState?: 'any' | 'available' | 'checked_out';
 readonly mode: ExpirationMode;
 readonly query?: string;
 readonly typeId?: string;
 readonly tagIds?: readonly string[];
 readonly locationId?: string;
 readonly fromDate?: string;
 readonly throughDate?: string;
 readonly cursor?: string;
 readonly limit?: number;
};
export type ExpirationCounts = { readonly soon: number; readonly expired: number; readonly all: number };
export type ExpirationPage = { readonly items: readonly AssetSummary[]; readonly counts: ExpirationCounts; readonly timezone: string; readonly hasMore: boolean; readonly nextCursor: string | null };
export interface ExpirationRepository { list(tenantId: string, inventoryId: string, filter: ExpirationFilter, signal?: AbortSignal): Promise<ExpirationPage>; }
export type ExpirationEvent = { readonly operation: 'list' | 'home'; readonly outcome: 'succeeded' | 'failed' };
export interface ExpirationObserver { record(event: ExpirationEvent): void; }
