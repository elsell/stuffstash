import { mount, unmount, tick } from 'svelte';
import { expect, it, vi } from 'vitest';
import NotificationInbox from './NotificationInbox.svelte';
import { StuffStashNotificationRepository } from '$lib/adapters/api/stuffStashNotificationRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
it('renders an unread month-only alert and resolves its current item before navigating', async () => {
  const requests: string[] = [];
  const opened: string[] = [];
  const notice = { id: 'notice', assetId: 'item', title: 'Tylenol', parentAssetId: 'bin', customAssetTypeId: 'medicine', expirationDate: '2026-10', expirationPrecision: 'month', milestone: 'upcoming', createdAt: '2026-09-10T12:00:00Z' };
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async (input, init) => {
    const request = new Request(input, init); requests.push(`${request.method} ${request.url}`);
    if (request.url.endsWith('/read')) return Response.json({ data: {}, meta: {} });
    if (request.url.endsWith('/notice')) return Response.json({ data: { ...notice, assetId: 'current-item' }, meta: {} });
    return Response.json({ data: [notice], meta: { pagination: { limit: 30, nextCursor: null, hasMore: false } } });
  });
  const component = mount(NotificationInbox, { target: document.body, props: { tenantId: 'tenant', inventoryId: 'inventory', repository, observer: new InMemoryWorkspaceObserver(), onOpenAsset: (id) => { opened.push(id); } } });
  try {
    const alert = () => Array.from(document.querySelectorAll('button')).find((button) => button.textContent?.includes('Tylenol'));
    await vi.waitFor(() => expect(alert()).toBeDefined());
    expect(alert()?.textContent).toContain('Unread');
    expect(alert()?.textContent).toContain('October 2026');
    alert()!.click();
    await vi.waitFor(() => expect(opened).toEqual(['current-item']));
    expect(requests.at(-1)).toContain('/notice/read');
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
it('shows errors separately from an empty inbox and retries the unread filter', async () => {
  let fail = true;
  const requests: string[] = [];
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async (input, init) => {
    const request = new Request(input, init); requests.push(request.url);
    return fail ? Response.json({ error: { code: 'unavailable', message: 'Unavailable' } }, { status: 503 })
      : Response.json({ data: [], meta: { pagination: { limit: 30, nextCursor: null, hasMore: false } } });
  });
  const component = mount(NotificationInbox, { target: document.body, props: { tenantId: 'tenant', inventoryId: 'inventory', repository, observer: new InMemoryWorkspaceObserver(), onOpenAsset() {} } });
  try {
    await vi.waitFor(() => expect(document.querySelector('[role="alert"]')).not.toBeNull());
    expect(document.body.textContent).not.toContain('No expiration notifications');
    fail = false;
    Array.from(document.querySelectorAll('button')).find((button) => button.textContent === 'Unread')!.click();
    await tick();
    await vi.waitFor(() => expect(document.body.textContent).toContain('No unread notifications.'));
    expect(requests.at(-1)).toContain('unreadOnly=true');
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
it('keeps continuation available after opening the last loaded unread notification', async () => {
  const notice = { id: 'notice', assetId: 'item', title: 'Bottle', parentAssetId: 'bin', customAssetTypeId: 'medicine', expirationDate: '2026-10-01', expirationPrecision: 'day', milestone: 'upcoming', createdAt: '2026-09-10T12:00:00Z' };
  let read = false;
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async (input, init) => {
    const request = new Request(input, init);
    if (request.url.endsWith('/read')) { read = true; return Response.json({ data: {}, meta: {} }); }
    if (request.url.endsWith('/notice')) return Response.json({ data: notice, meta: {} });
    return Response.json({ data: [notice], meta: { pagination: { limit: 1, nextCursor: 'next', hasMore: true } } });
  });
  const component = mount(NotificationInbox, { target: document.body, props: { tenantId: 'tenant', inventoryId: 'inventory', repository, observer: new InMemoryWorkspaceObserver(), onOpenAsset() {} } });
  try {
    await vi.waitFor(() => expect(document.body.textContent).toContain('Bottle'));
    Array.from(document.querySelectorAll('button')).find((button) => button.textContent === 'Unread')!.click();
    await tick();
    await vi.waitFor(() => expect(document.body.textContent).toContain('Bottle'));
    Array.from(document.querySelectorAll('button')).find((button) => button.textContent?.includes('Bottle'))!.click();
    await vi.waitFor(() => expect(read).toBe(true));
    await vi.waitFor(() => expect(document.body.textContent).not.toContain('Bottle'));
    expect(document.body.textContent).not.toContain('No unread notifications');
    expect(document.body.textContent).toContain('Load more');
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
