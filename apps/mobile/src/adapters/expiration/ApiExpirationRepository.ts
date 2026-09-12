import type { StuffStashClient, ExpirationWorkspaceAsset } from '@stuff-stash/api-client';
import { assetId, type AssetSummary } from '../../domain/assets/AssetSummary';
import type { ExpirationFilter, ExpirationRepository } from '../../application/expiration/ExpirationRepository';
import { assertReadActive } from '../../application/shared/ReadRequest';
import { isAssetExpiration } from '../../domain/assets/AssetExpiration';

export class ApiExpirationRepository implements ExpirationRepository {
 constructor(private readonly client: Pick<StuffStashClient, 'expiration' | 'assetAttachmentThumbnailReference'>) {}
 async list(tenantId: string, inventoryId: string, filter: ExpirationFilter, signal?: AbortSignal) {
  const page = await this.client.expiration.list(tenantId, inventoryId, { kind:filter.kind,checkoutState:filter.checkoutState,mode: filter.mode, q: filter.query, customAssetTypeId: filter.typeId, tagIds: filter.tagIds ? [...filter.tagIds] : undefined, locationId: filter.locationId, fromDate: filter.fromDate, throughDate: filter.throughDate, cursor: filter.cursor, limit: filter.limit, signal });
  assertReadActive(signal);
  const items = await Promise.all(page.items.map(async value => {
   if (value.tenantId !== tenantId || value.inventoryId !== inventoryId) throw new Error('Expiration inventory changed. Refresh and try again.');
   const item = mapExpirationAsset(value);
   if (!value.primaryPhoto) return item;
   const photo = await this.client.assetAttachmentThumbnailReference(tenantId, inventoryId, value.id, value.primaryPhoto.id, 'small');
   return { ...item, hasPhoto: true, photo: { uri: photo.uri, headers: photo.headers } };
  }));
  assertReadActive(signal);
  return { items, counts: page.counts, timezone: page.timezone, nextCursor: page.pagination.nextCursor, hasMore: page.pagination.hasMore };
 }
}
function mapExpirationAsset(value: ExpirationWorkspaceAsset): AssetSummary {
 if (value.lifecycleState !== 'active' || !isAssetExpiration(value.expiration) || !value.expirationContext || !['item', 'container', 'location'].includes(value.kind)) throw new Error('Expiration item could not be read. Try refreshing.');
 const context = value.expirationContext;
 if (context.state !== 'current' && context.state !== 'upcoming' && context.state !== 'expired') throw new Error('Expiration status could not be read.');
 return { currentCheckout:value.currentCheckout ? {id:value.currentCheckout.id,state:value.currentCheckout.state,checkedOutAt:value.currentCheckout.checkedOutAt,checkedOutByPrincipalId:value.currentCheckout.checkedOutByPrincipalId}:undefined,id: assetId(value.id), title: value.title, kind: value.kind as AssetSummary['kind'], lifecycleState: 'active', description: value.description, locationLabel: (value.ancestorPath ?? []).at(-1)?.title ?? 'Inventory root', locationTrail: (value.ancestorPath ?? []).map(parent => parent.title), parentLocationTrail: (value.ancestorPath ?? []).map(parent => ({ id: assetId(parent.id), title: parent.title })), updatedAtLabel: '', hasPhoto: false, expiration: value.expiration, expirationContext: { state: context.state, trackingEnabled: context.trackingEnabled, advanceDays: context.advanceDays, timezone: context.timezone }, customAssetTypeId: value.customAssetTypeId, tags: value.tags?.map(tag => ({ id: tag.id, key: tag.key, displayName: tag.displayName, color: tag.color })) };
}
