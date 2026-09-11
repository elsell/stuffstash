import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';
import type { InboxOptions, NotificationRepository } from './NotificationRepository';
import type { NotificationEvent, NotificationObservability } from './NotificationObservability';
type InboxRepository = Pick<NotificationRepository, 'listInbox' | 'countUnreadPage' | 'getNotification' | 'markRead' | 'markUnread' | 'markAllReadPage'>;
const maximumPages = 100;
function incomplete(): Error { return new Error('The notification operation could not be completed. Try again.'); }

export class NotificationInboxQueries {
  constructor(private readonly repository: InboxRepository, private readonly observer: NotificationObservability) {}
  list(tenantId: string, inventoryId: string, options: InboxOptions = {}) {
    return this.observe('list', async () => {
      let cursor = options.cursor;
      const seen = new Set<string>(cursor ? [cursor] : []);
      for (let index = 0; index < maximumPages; index++) {
        assertReadActive(options.signal);
        const page = await this.repository.listInbox(tenantId, inventoryId, { ...options, cursor });
        assertReadActive(options.signal);
        const next = page.pagination.nextCursor;
        if (page.pagination.hasMore && (!next || seen.has(next))) throw incomplete();
        if (page.items.length || !page.pagination.hasMore) return page;
        seen.add(next!); cursor = next!;
      }
      throw incomplete();
    });
  }
  count(tenantId: string, inventoryId: string, request: ReadRequest = {}): Promise<number> {
    return this.observe('count', async () => {
      let count = 0;
      for await (const page of this.pages((cursor) => this.repository.countUnreadPage(tenantId, inventoryId, cursor, request.signal), request)) {
        if (!Number.isSafeInteger(page.count) || page.count < 0 || !Number.isSafeInteger(count + page.count)) throw incomplete();
        count += page.count;
      }
      return count;
    });
  }
  markAllRead(tenantId: string, inventoryId: string, request: ReadRequest = {}): Promise<void> {
    return this.observe('mark-all', async () => {
      for await (const page of this.pages((cursor) => this.repository.markAllReadPage(tenantId, inventoryId, cursor, request.signal), request)) {
        if (page.complete !== (page.nextCursor === null)) throw incomplete();
      }
    });
  }
  open(tenantId: string, inventoryId: string, notificationId: string, request: ReadRequest = {}): Promise<string> {
    return this.observe('open', async () => {
      assertReadActive(request.signal);
      const current = await this.repository.getNotification(tenantId, inventoryId, notificationId, request.signal);
      assertReadActive(request.signal);
      await this.repository.markRead(tenantId, inventoryId, current.id, request.signal);
      assertReadActive(request.signal);
      return current.assetId;
    });
  }
  setRead(tenantId: string, inventoryId: string, notificationId: string, read: boolean, request: ReadRequest = {}): Promise<void> {
    return this.observe('read-state', async () => {
      assertReadActive(request.signal);
      if (read) await this.repository.markRead(tenantId, inventoryId, notificationId, request.signal);
      else await this.repository.markUnread(tenantId, inventoryId, notificationId, request.signal);
      assertReadActive(request.signal);
    });
  }
  private async *pages<T extends { nextCursor: string | null }>(load: (cursor?: string) => Promise<T>, request: ReadRequest) {
    let cursor: string | undefined;
    const seen = new Set<string>();
    for (let index = 0; index < maximumPages; index++) {
      assertReadActive(request.signal);
      const page = await load(cursor);
      assertReadActive(request.signal);
      if (page.nextCursor !== null && (!page.nextCursor || seen.has(page.nextCursor))) throw incomplete();
      yield page;
      if (page.nextCursor === null) return;
      seen.add(page.nextCursor); cursor = page.nextCursor;
    }
    throw incomplete();
  }
  private async observe<T>(operation: NotificationEvent['operation'], run: () => Promise<T>): Promise<T> {
    try { const result = await run(); this.observer.record({ operation, outcome: 'succeeded' }); return result; }
    catch (error) { this.observer.record({ operation, outcome: 'failed' }); throw error; }
  }
}
