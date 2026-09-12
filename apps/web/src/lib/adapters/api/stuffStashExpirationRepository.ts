import { StuffStashClient, type TokenProvider, type ExpirationWorkspaceAsset, type Page } from '@stuff-stash/api-client';
import type { ExpirationRepository, ExpirationFilter, ExpirationItem } from '$lib/ports/expirationRepository';
export class StuffStashExpirationRepository implements ExpirationRepository {
 private readonly client: StuffStashClient;
 constructor(baseUrl: string, tokenProvider: TokenProvider, fetchImpl?: typeof fetch) { this.client = new StuffStashClient({ baseUrl, tokenProvider, fetch: fetchImpl }); }
 async list(tenantId: string, inventoryId: string, filter: ExpirationFilter, options: { cursor?: string; limit?: number; signal?: AbortSignal } = {}) {
  const page = await this.client.expiration.list(tenantId, inventoryId, { ...options, kind:filter.kind,checkoutState:filter.checkoutState,mode: filter.mode, q: filter.query || undefined, customAssetTypeId: filter.typeId || undefined, tagIds: filter.tagIds?.length ? filter.tagIds : undefined, locationId: filter.locationId || undefined, fromDate: filter.fromDate || undefined, throughDate: filter.throughDate || undefined });
  return { items: page.items.map(item => mapItem(item, tenantId, inventoryId)), counts: { ...page.counts }, timezone: page.timezone, hasMore: page.pagination.hasMore, nextCursor: page.pagination.nextCursor };
 }
 async choices(tenantId: string, inventoryId: string, signal?: AbortSignal) {
  const [tenantTypes, inventoryTypes, tags, assets] = await Promise.all([
   collect(cursor => this.client.listTenantCustomAssetTypes(tenantId, 100, cursor, 'active', signal), signal),
   collect(cursor => this.client.listInventoryCustomAssetTypes(tenantId, inventoryId, 100, cursor, 'active', signal), signal),
   collect(cursor => this.client.listAssetTags(tenantId, inventoryId, 100, cursor, signal), signal),
   collect(cursor => this.client.listAssets(tenantId, inventoryId, 100, cursor, 'active', 'id_asc', signal), signal)
  ]);
  return { types: [...tenantTypes, ...inventoryTypes].map(value => ({ id: value.id, title: value.displayName })), tags: tags.map(value => ({ id: value.id, title: value.displayName })), locations: assets.filter(value => value.kind === 'location' || value.kind === 'container').map(value => ({ id: value.id, title: value.title })).sort((a, b) => a.title.localeCompare(b.title)) };
 }
}
function mapItem(item: ExpirationWorkspaceAsset, tenantId: string, inventoryId: string): ExpirationItem {
 if (item.tenantId !== tenantId || item.inventoryId !== inventoryId || item.lifecycleState !== 'active' || !item.expiration || !item.expirationContext || !['item','container','location'].includes(item.kind)) throw new Error('Expiration results are unavailable. Refresh and try again.');
 return { currentCheckout:item.currentCheckout ? {id:item.currentCheckout.id,state:'open',checkedOutAt:item.currentCheckout.checkedOutAt,checkedOutByPrincipalId:item.currentCheckout.checkedOutByPrincipalId}:undefined,id: item.id, tenantId, inventoryId, kind: item.kind as ExpirationItem['kind'], lifecycleState: 'active', title: item.title, description: item.description, parentAssetId: item.parentAssetId ?? null, customAssetTypeId: item.customAssetTypeId, expiration: { ...item.expiration }, expirationContext: { ...item.expirationContext }, primaryPhotoId: item.primaryPhoto?.id, tags: (item.tags ?? []).map(tag => ({ id: tag.id, key: tag.key, displayName: tag.displayName, color: tag.color })), ancestorPath: (item.ancestorPath ?? []).map(parent => ({ id: parent.id, title: parent.title })) };
}
async function collect<T>(load: (cursor?: string) => Promise<Page<T>>, signal?: AbortSignal): Promise<T[]> {
 const items: T[] = []; const seen = new Set<string>(); let cursor: string | undefined;
 for (;;) {
  signal?.throwIfAborted(); const page = await load(cursor); signal?.throwIfAborted(); items.push(...page.items);
  if (!page.pagination.hasMore) return items;
  const next = page.pagination.nextCursor;
  if (!next || seen.has(next)) throw new Error('Filter choices are incomplete. Try again.');
  seen.add(next); cursor = next;
 }
}
