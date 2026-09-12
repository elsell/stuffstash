import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import type { Asset } from '$lib/domain/inventory';
export type ExpirationMode = 'soon' | 'expired' | 'all';
export interface ExpirationFilter { kind?: 'item' | 'container' | 'location'; checkoutState?: 'any' | 'available' | 'checked_out'; mode: ExpirationMode; query?: string; typeId?: string; tagIds?: string[]; locationId?: string; fromDate?: string; throughDate?: string; }
export interface ExpirationItem extends Asset { ancestorPath: { id: string; title: string }[]; }
export interface ExpirationPage { items: ExpirationItem[]; counts: { soon: number; expired: number; all: number }; timezone: string; hasMore: boolean; nextCursor: string | null; }
export interface ExpirationChoices { types: { id: string; title: string }[]; tags: { id: string; title: string }[]; locations: { id: string; title: string }[]; }
export interface ExpirationRepository {
 list(tenantId: string, inventoryId: string, filter: ExpirationFilter, options?: { cursor?: string; limit?: number; signal?: AbortSignal }): Promise<ExpirationPage>;
 choices(tenantId: string, inventoryId: string, signal?: AbortSignal): Promise<ExpirationChoices>;
}
export const expirationWorkspaceContext = Symbol('expirationWorkspace');
export interface ExpirationWorkspace { repository: ExpirationRepository; observer?: WorkspaceObserver; positions?: Map<string, {scrollY:number;assetId:string}>; cache?: Map<string, ExpirationPage>; revision?: () => unknown; }
