import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import type { ExpirationFilter, ExpirationPage, ExpirationRepository } from '$lib/ports/expirationRepository';
export interface ExpirationListState { page?: ExpirationPage; loading: boolean; appending: boolean; error: string; }
export class ExpirationList {
 state: ExpirationListState = { loading: false, appending: false, error: '' };
 private controller?: AbortController;
 private key = '';
 constructor(private readonly repository: ExpirationRepository, private readonly changed: (state: ExpirationListState) => void, private readonly cache = new Map<string, ExpirationPage>(), private readonly observer?: WorkspaceObserver) {}
 dispose() { this.controller?.abort(); }
 async load(tenantId: string, inventoryId: string, filter: ExpirationFilter, append = false): Promise<void> {
  const key = JSON.stringify([tenantId, inventoryId, filter]);
  this.controller?.abort(); const controller = new AbortController(); this.controller = controller;
  if (key !== this.key) { this.key = key; this.state = { page: this.cache.get(key), loading: false, appending: false, error: '' }; }
  const previous = this.state.page;
  this.update({ ...this.state, loading: !append, appending: append, error: '' });
  try {
   let cursor = append ? previous?.nextCursor ?? undefined : undefined;
   let items = append ? previous?.items ?? [] : [];
   const target = append ? items.length + 1 : Math.max(previous?.items.length ?? 0, 1);
   const seen = new Set<string>(); let page: ExpirationPage;
   do {
    page = await this.repository.list(tenantId, inventoryId, filter, { cursor, limit: 30, signal: controller.signal });
    if (controller.signal.aborted) return;
    if (page.hasMore && (!page.nextCursor || page.nextCursor === cursor || seen.has(page.nextCursor))) throw new Error('Incomplete expiration results.');
    if (page.nextCursor) seen.add(page.nextCursor);
    items = [...new Map([...items, ...page.items].map(item => [item.id, item])).values()]; cursor = page.nextCursor ?? undefined;
   } while (!append && page.hasMore && items.length < target);
   this.observer?.record('workspace.expiration_loaded');
   const complete = { ...page, items };
   this.cache.set(key, complete);
   if (this.cache.size > 20) this.cache.delete(this.cache.keys().next().value!);
   this.update({ page: complete, loading: false, appending: false, error: '' });
  } catch (error) {
   if (controller.signal.aborted) return;
   this.observer?.record('workspace.expiration_load_failed');
   const denied = typeof error === 'object' && error !== null && 'status' in error && [401,403,404].includes(Number(error.status));
   if (denied) {
    const scopePrefix = JSON.stringify([tenantId,inventoryId]).slice(0,-1) + ',';
    for (const cachedKey of this.cache.keys()) if (cachedKey.startsWith(scopePrefix)) this.cache.delete(cachedKey);
   }
   this.update({ page: denied ? undefined : this.state.page, loading: false, appending: false, error: denied ? 'This inventory is no longer available.' : append ? 'More expiration dates could not be loaded. Try again.' : 'Expiration could not be refreshed. Try again.' });
  }
 }
 private update(state: ExpirationListState) { this.state = state; this.changed(state); }
}
