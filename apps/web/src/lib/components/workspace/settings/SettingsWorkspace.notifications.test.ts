import { mount, unmount } from 'svelte';
import { writable, fromStore } from 'svelte/store';
import { expect, it, vi } from 'vitest';
import SettingsWorkspace from './SettingsWorkspace.svelte';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import { StuffStashNotificationRepository } from '$lib/adapters/api/stuffStashNotificationRepository';
import { notificationWorkspaceContext } from '$lib/ports/notificationWorkspace';
import { parseWorkspaceRoute } from '$lib/application/workspaceRoute';
import type { Inventory, Tenant } from '$lib/domain/inventory';

it('renders personal settings for viewers and replaces drafts when inventory changes', async () => {
  const principal = { id: 'viewer', email: 'viewer@example.test' };
  const tenant: Tenant = { id: 'home', name: 'Home', access: { relationship: 'viewer', permissions: ['view'] } };
  const first: Inventory = { id: 'first', tenantId: 'home', name: 'First', access: { relationship: 'viewer', permissions: ['view'] } };
  const second: Inventory = { ...first, id: 'second', name: 'Second' };
  const inventoryStore = writable(first);
  const selected = fromStore(inventoryStore);
  const requests: string[] = [];
  const notificationRepository = new StuffStashNotificationRepository('https://api.test', () => 'token', async (input, init) => {
    const request = new Request(input, init); requests.push(request.url);
    return Response.json({ data: { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: request.url.includes('/second/') ? 60 : 30 }, timezone: 'UTC', pushEnabled: false, overrides: [] }, meta: {} });
  });
  const repository = new SeededInventoryRepository({ principal, tenants: [tenant], inventories: [first, second], assets: [], customAssetTypes: [], customFieldDefinitions: [] });
  const component = mount(SettingsWorkspace, { target: document.body, context: new Map([[notificationWorkspaceContext, { apiIdentity: 'https://api.test', repository: notificationRepository }]]), props: {
    principal, tenant, get inventory() { return selected.current; }, route: parseWorkspaceRoute(new URL('https://app.test/settings/tenants/home/inventories/first/notifications')),
    repository, observer: { record() {} }, currentAssetTypes: [], currentFields: [], onNavigate() {}, onSchemaChange() {}, onTagsChange() {}, async onPermissionDenied() {}
  } });
  try {
    const days = () => document.querySelector<HTMLInputElement>('input[type="number"]');
    await vi.waitFor(() => expect(days()?.value).toBe('30'));
    days()!.value = '7'; days()!.dispatchEvent(new Event('input', { bubbles: true }));
    inventoryStore.set(second);
    await vi.waitFor(() => expect(days()?.value).toBe('60'));
    expect(requests.some((url) => url.includes('/second/notification-preferences/initialize'))).toBe(true);
  } finally { await unmount(component); document.body.innerHTML = ''; }
});
