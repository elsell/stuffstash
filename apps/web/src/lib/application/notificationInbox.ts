import type { InboxOptions, NotificationPage, NotificationRepository } from '$lib/ports/notificationRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
const maximumSparsePages = 100;

export async function loadVisibleNotificationPage(
  repository: Pick<NotificationRepository, 'listInbox'>, observer: WorkspaceObserver,
  tenantId: string, inventoryId: string, options: InboxOptions = {}
): Promise<NotificationPage> {
  observer.record('workspace.notification_inbox_load_started');
  try {
    let cursor = options.cursor;
    const seen = new Set<string>(cursor ? [cursor] : []);
    for (let index = 0; index < maximumSparsePages; index++) {
      options.signal?.throwIfAborted();
      const page = await repository.listInbox(tenantId, inventoryId, { ...options, cursor });
      options.signal?.throwIfAborted();
      const next = page.pagination.nextCursor;
      if (page.pagination.hasMore && (!next || seen.has(next))) throw new Error('Notifications could not be fully loaded.');
      if (page.items.length || !page.pagination.hasMore) {
        observer.record('workspace.notification_inbox_loaded', { count: page.items.length });
        return page;
      }
      seen.add(next!);
      cursor = next!;
    }
    throw new Error('Notifications could not be fully loaded.');
  } catch (error) {
    observer.record('workspace.notification_inbox_load_failed');
    throw error;
  }
}

export async function openNotification(
  repository: Pick<NotificationRepository, 'getNotification' | 'markRead'>, observer: WorkspaceObserver,
  tenantId: string, inventoryId: string, notificationId: string, signal?: AbortSignal
): Promise<string> {
  observer.record('workspace.notification_open_started');
  try {
    signal?.throwIfAborted();
    const current = await repository.getNotification(tenantId, inventoryId, notificationId, signal);
    signal?.throwIfAborted();
    await repository.markRead(tenantId, inventoryId, current.id, signal);
    signal?.throwIfAborted();
    observer.record('workspace.notification_opened');
    return current.assetId;
  } catch (error) {
    observer.record('workspace.notification_open_failed');
    throw error;
  }
}
