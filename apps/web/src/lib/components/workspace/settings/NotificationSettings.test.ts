import { mount, unmount } from 'svelte';
import { expect, it, vi } from 'vitest';
import NotificationSettings from './NotificationSettings.svelte';
import { StuffStashNotificationRepository } from '$lib/adapters/api/stuffStashNotificationRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
it('loads personal settings and saves an inventory policy through its scoped repository', async () => {
  let settings = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'America/New_York', pushEnabled: false, overrides: [] };
  const requests: Request[] = [];
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async (input, init) => {
    const request = new Request(input, init); requests.push(request);
    if (request.method === 'PUT') settings = { ...settings, ...await request.json(), revision: settings.revision + 1 };
    return Response.json({ data: settings, meta: {} });
  });
  const component = mount(NotificationSettings, { target: document.body, props: { tenantId: 'tenant', inventoryId: 'inventory', initialTimezone: 'UTC', repository, observer: new InMemoryWorkspaceObserver(), typeRepository: { async listInventoryCustomAssetTypes() { return { items: [], pagination: { limit: 30, hasMore: false, nextCursor: null } }; } } } });
  try {
    await vi.waitFor(() => expect(document.querySelector('input[type="number"]')).not.toBeNull());
    expect(requests[0].url).toContain('/tenants/tenant/inventories/inventory/notification-preferences/initialize');
    const input = document.querySelector<HTMLInputElement>('input[type="number"]')!;
    input.value = '7'; input.dispatchEvent(new Event('input', { bubbles: true }));
    const button = Array.from(document.querySelectorAll('button')).find((value) => value.textContent?.includes('Save reminders'))!;
    button.click();
    await vi.waitFor(() => expect(settings.defaults.advanceDays).toBe(7));
    expect(settings.timezone).toBe('America/New_York');
    expect(document.body.textContent).toContain('mobile app');
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
it('shows a retryable type-load failure instead of claiming no types support expiration', async () => {
  let fail = true;
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async () => Response.json({ data: { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'UTC', pushEnabled: false, overrides: [] }, meta: {} }));
  const component = mount(NotificationSettings, { target: document.body, props: {
    tenantId: 'tenant', inventoryId: 'inventory', initialTimezone: 'UTC', repository, observer: new InMemoryWorkspaceObserver(),
    typeRepository: { async listInventoryCustomAssetTypes() { if (fail) throw new Error('Temporary failure'); return { items: [], pagination: { limit: 30, hasMore: false, nextCursor: null } }; } }
  } });
  try {
    await vi.waitFor(() => expect(document.querySelector('[role="alert"]')).not.toBeNull());
    expect(document.body.textContent).not.toContain('Enable expiration tracking');
    fail = false;
    Array.from(document.querySelectorAll('button')).find((button) => button.textContent?.includes('Retry loading'))!.click();
    await vi.waitFor(() => expect(document.querySelector('input[type="number"]')).not.toBeNull());
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
