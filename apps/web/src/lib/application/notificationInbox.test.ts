import { expect, it } from 'vitest';
import { loadVisibleNotificationPage, openNotification } from './notificationInbox';
import type { ExpirationNotification } from '$lib/domain/notification';
import type { InboxOptions, NotificationPage } from '$lib/ports/notificationRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
const notice: ExpirationNotification = { id: 'notice', assetId: 'current-asset', title: 'Bottle', parentAssetId: 'bin', customAssetTypeId: 'medicine', expiration: { date: '2026-10', precision: 'month' }, milestone: 'upcoming', createdAt: '2026-09-10T12:00:00Z' };
class InboxFake {
  requests: InboxOptions[] = [];
  pages: NotificationPage[] = [];
  read = false;
  unavailable = false;
  async listInbox(_tenant: string, _inventory: string, options: InboxOptions = {}) {
    this.requests.push(options);
    const page = this.pages.shift();
    if (!page) throw new Error('No page');
    return page;
  }
  async getNotification() { if (this.unavailable) throw new Error('Unavailable'); return notice; }
  async markRead() { this.read = true; }
}
it('continues sparse inbox pages with the same unread filter and cancellation signal', async () => {
  const repository = new InboxFake();
  repository.pages = [
    { items: [], pagination: { limit: 30, nextCursor: 'next', hasMore: true } },
    { items: [notice], pagination: { limit: 30, nextCursor: null, hasMore: false } }
  ];
  const signal = new AbortController().signal;
  const result = await loadVisibleNotificationPage(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory', { unreadOnly: true, signal });
  expect(result.items).toEqual([notice]);
  expect(repository.requests[1]).toMatchObject({ cursor: 'next', unreadOnly: true, signal });
});
it('rejects broken pagination instead of claiming an empty inbox', async () => {
  const repository = new InboxFake();
  repository.pages = [{ items: [], pagination: { limit: 30, nextCursor: 'same', hasMore: true } }];
  await expect(loadVisibleNotificationPage(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory', { cursor: 'same' })).rejects.toThrow('fully loaded');
});
it('resolves current asset access before marking read or returning a navigation target', async () => {
  const repository = new InboxFake();
  const observer = new InMemoryWorkspaceObserver();
  repository.unavailable = true;
  await expect(openNotification(repository, observer, 'tenant', 'inventory', 'notice')).rejects.toThrow('Unavailable');
  expect(repository.read).toBe(false);
  repository.unavailable = false;
  await expect(openNotification(repository, observer, 'tenant', 'inventory', 'notice')).resolves.toBe('current-asset');
  expect(repository.read).toBe(true);
});
it('does not publish a page after cancellation even if the repository completes', async () => {
  const controller = new AbortController();
  const repository = { async listInbox() {
    controller.abort();
    return { items: [notice], pagination: { limit: 30, nextCursor: null, hasMore: false } };
  } };
  await expect(loadVisibleNotificationPage(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory', { signal: controller.signal })).rejects.toThrow();
});
it('returns an empty page only after exhaustion and bounds sparse traversal', async () => {
  const repository = new InboxFake();
  repository.pages = [{ items: [], pagination: { limit: 30, nextCursor: null, hasMore: false } }];
  await expect(loadVisibleNotificationPage(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory')).resolves.toMatchObject({ items: [], pagination: { hasMore: false } });
  let page = 0;
  const endless = { async listInbox() { return { items: [], pagination: { limit: 30, nextCursor: String(++page), hasMore: true } }; } };
  await expect(loadVisibleNotificationPage(endless, new InMemoryWorkspaceObserver(), 'tenant', 'inventory')).rejects.toThrow('fully loaded');
  expect(page).toBe(100);
});
