import { mount, unmount } from 'svelte';
import { expect, it, vi } from 'vitest';
import NotificationBell from './NotificationBell.svelte';
import { StuffStashNotificationRepository } from '$lib/adapters/api/stuffStashNotificationRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
it('registers the inventory before loading a complete unread badge', async () => {
  const requests: string[] = [];
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async (input, init) => {
    const request = new Request(input, init); requests.push(request.url);
    if (request.url.endsWith('/initialize')) return Response.json({ data: { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'UTC', pushEnabled: false, overrides: [] }, meta: {} });
    return Response.json({ data: { count: 3 }, meta: { pagination: { limit: 100, nextCursor: null, hasMore: false } } });
  });
  const component = mount(NotificationBell, { target: document.body, props: { tenantId: 'tenant', inventoryId: 'inventory', repository, observer: new InMemoryWorkspaceObserver(), onOpenAsset() {}, onOpenSettings() {} } });
  try {
    await vi.waitFor(() => expect(document.querySelector('[aria-label="Notifications, 3 unread"]')).not.toBeNull());
    expect(requests[0]).toContain('/notification-preferences/initialize');
    expect(requests[1]).toContain('/notifications/unread-count');
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
