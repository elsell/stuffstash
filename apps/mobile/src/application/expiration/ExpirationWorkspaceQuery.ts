import { toAssetCardViewModel } from '../assets/AssetViewModels';
import { assertReadActive } from '../shared/ReadRequest';
import type { ExpirationFilter, ExpirationObserver, ExpirationRepository } from './ExpirationRepository';

export class ExpirationWorkspaceQuery {
 constructor(private readonly repository: ExpirationRepository, private readonly observer: ExpirationObserver) {}
 async list(tenantId: string, inventoryId: string, filter: ExpirationFilter, signal?: AbortSignal) {
  try {
   assertReadActive(signal);
   const page = await this.repository.list(tenantId, inventoryId, filter, signal);
   assertReadActive(signal);
   if (page.hasMore && (!page.nextCursor || page.nextCursor === filter.cursor)) throw new Error('Expiration results could not be continued. Refresh and try again.');
   this.observer.record({ operation: 'list', outcome: 'succeeded' });
   return { ...page, items: page.items.map(toAssetCardViewModel) };
  } catch (error) { this.observer.record({ operation: 'list', outcome: 'failed' }); throw error; }
 }
 async home(tenantId: string, inventoryId: string, signal?: AbortSignal) {
  try {
   const [expired, soon] = await Promise.all([
    this.list(tenantId, inventoryId, { mode: 'expired', limit: 2 }, signal),
    this.list(tenantId, inventoryId, { mode: 'soon', limit: 2 }, signal)
   ]);
   assertReadActive(signal);
   const items = [expired.items[0], soon.items[0], expired.items[1], soon.items[1]].filter((item): item is NonNullable<typeof item> => !!item).slice(0, 3);
   this.observer.record({ operation: 'home', outcome: 'succeeded' });
   return { items, counts: expired.counts, timezone: expired.timezone };
  } catch (error) { this.observer.record({ operation: 'home', outcome: 'failed' }); throw error; }
 }
}
