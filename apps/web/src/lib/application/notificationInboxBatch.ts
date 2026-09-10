import type { NotificationRepository } from '$lib/ports/notificationRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
const maximumBatchPages = 100;
const incomplete = () => new Error('The notification operation could not be completed. Try again.');

async function* pages<T extends { nextCursor: string | null }>(load: (cursor?: string) => Promise<T>, signal?: AbortSignal): AsyncGenerator<T> {
  let cursor: string | undefined;
  const seen = new Set<string>();
  for (let index = 0; index < maximumBatchPages; index++) {
    signal?.throwIfAborted();
    const page = await load(cursor);
    signal?.throwIfAborted();
    if (page.nextCursor !== null && (!page.nextCursor || seen.has(page.nextCursor))) throw incomplete();
    yield page;
    if (page.nextCursor === null) return;
    seen.add(page.nextCursor); cursor = page.nextCursor;
  }
  throw incomplete();
}
export async function countUnreadNotifications(repository: Pick<NotificationRepository, 'countUnreadPage'>, observer: WorkspaceObserver, tenantId: string, inventoryId: string, signal?: AbortSignal): Promise<number> {
  try {
    let count = 0;
    for await (const page of pages((cursor) => repository.countUnreadPage(tenantId, inventoryId, cursor, signal), signal)) {
      if (!Number.isSafeInteger(page.count) || page.count < 0 || !Number.isSafeInteger(count + page.count)) throw incomplete();
      count += page.count;
    }
    observer.record('workspace.notification_count_loaded');
    return count;
  } catch (error) { observer.record('workspace.notification_count_failed'); throw error; }
}
export async function markAllNotificationsRead(repository: Pick<NotificationRepository, 'markAllReadPage'>, observer: WorkspaceObserver, tenantId: string, inventoryId: string, signal?: AbortSignal): Promise<void> {
  try {
    for await (const page of pages((cursor) => repository.markAllReadPage(tenantId, inventoryId, cursor, signal), signal)) {
      if (page.complete !== (page.nextCursor === null)) throw incomplete();
    }
    observer.record('workspace.notifications_marked_read');
  } catch (error) { observer.record('workspace.notifications_mark_read_failed'); throw error; }
}
