import { expect, it } from 'vitest';
import { StuffStashClient, StuffStashAPIError } from './stuffStashClient';

it('preserves personal false settings and revisions using authenticated scoped requests', async () => {
  const requests: Request[] = [];
  const policy = { enabled: false, upcoming: false, expired: true, advanceDays: 0 };
  const preferences = { revision: 7, defaults: policy, timezone: 'America/New_York', pushEnabled: false, overrides: [] };
  const client = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: async () => 'token', fetch: async (input, init) => {
    requests.push(new Request(input, init)); return Response.json({ data: preferences, meta: {} });
  }});
  await expect(client.notifications.updatePreferences('tenant', 'inventory', preferences)).resolves.toEqual(preferences);
  const { overrides: _overrides, ...update } = preferences;
  expect(await requests[0].json()).toEqual(update);
  expect(requests[0].headers.get('Authorization')).toBe('Bearer token');
  await client.notifications.removeTypeOverride('tenant', 'inventory', 'medicine', 7);
  expect(requests[1].method).toBe('DELETE');
  expect(requests[1].url).toBe('https://api.test/tenants/tenant/inventories/inventory/notification-preferences/types/medicine?revision=7');
});
it('retains empty inbox continuation pages and surfaces API failures', async () => {
  let failure = false;
  const client = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => null, fetch: async () => failure
    ? Response.json({ error: { code: 'forbidden', message: 'Denied' } }, { status: 403 })
    : Response.json({ data: [], meta: { pagination: { limit: 30, nextCursor: 'next', hasMore: true } } }) });
  await expect(client.notifications.listInbox('tenant', 'inventory', { unreadOnly: true })).resolves.toMatchObject({ items: [], pagination: { nextCursor: 'next', hasMore: true } });
  failure = true;
  await expect(client.notifications.markRead('tenant', 'inventory', 'notice')).rejects.toBeInstanceOf(StuffStashAPIError);
});

it('initializes preferences and writes a complete type override without dropping disabled flags', async () => {
  const requests: Request[] = [];
  const policy = { enabled: true, upcoming: false, expired: false, advanceDays: 90 };
  const client = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async (input, init) => {
    requests.push(new Request(input, init));
    return Response.json({ data: { revision: requests.length, defaults: policy, timezone: 'America/New_York', pushEnabled: false, overrides: [] }, meta: {} });
  }});
  await client.notifications.initializePreferences('tenant', 'inventory', 'America/New_York');
  await client.notifications.setTypeOverride('tenant', 'inventory', 'medicine', 1, policy);
  expect(requests[0].method).toBe('POST');
  expect(await requests[0].json()).toEqual({ timezone: 'America/New_York' });
  expect(requests[1].method).toBe('PUT');
  expect(await requests[1].json()).toEqual({ revision: 1, settings: policy });
});
it('preserves count contributions and mark-all continuation without claiming completion', async () => {
  const client = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async (input, init) => {
    const request = new Request(input, init);
    return Response.json({ data: request.method === 'GET' ? { count: 2 } : { complete: false }, meta: { pagination: { limit: 100, nextCursor: 'next', hasMore: true } } });
  }});
  await expect(client.notifications.countUnreadPage('tenant', 'inventory')).resolves.toEqual({ count: 2, nextCursor: 'next' });
  await expect(client.notifications.markAllReadPage('tenant', 'inventory')).resolves.toEqual({ complete: false, nextCursor: 'next' });
});
it('registers, looks up and revokes a scoped device without losing revisions or false state', async () => {
 const requests: Request[] = [];
 const metadata = { id: 'device', installationId: 'phone', transport: 'apns', revision: 1, active: false };
 const client = new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'access',fetch:async(input,init)=>{requests.push(new Request(input,init));return Response.json({data:metadata,meta:{}});}});
 const input = { installationId:'phone',transport:'apns' as const,token:'abab',revision:0 };
 await expect(client.notifications.registerDevice('tenant','inventory',input)).resolves.toEqual(metadata);
 expect(await requests[0].json()).toEqual(input);
 expect(requests[0].headers.get('Authorization')).toBe('Bearer access');
 await client.notifications.getDeviceByInstallation('tenant','inventory','phone');
 await client.notifications.revokeDevice('tenant','inventory','device',1);
 expect(requests[1].url).toBe('https://api.test/tenants/tenant/inventories/inventory/notification-devices/by-installation/phone');
 expect(requests[2].method).toBe('DELETE');
 expect(requests[2].url).toBe('https://api.test/tenants/tenant/inventories/inventory/notification-devices/device?revision=1');
});

it('marks unread through the authenticated read subresource without sending another notification', async () => {
 const requests: Request[] = [];
 const client = new StuffStashClient({baseUrl: 'https://api.test', tokenProvider: () => 'access', fetch: async (input, init) => {
  requests.push(new Request(input, init)); return Response.json({data: {id: 'notice', read: false}, meta: {}});
 }});
 await client.notifications.markUnread('tenant', 'inventory', 'notice');
 expect(requests[0].method).toBe('DELETE');
 expect(requests[0].url).toBe('https://api.test/tenants/tenant/inventories/inventory/notifications/notice/read');
 expect(requests[0].headers.get('Authorization')).toBe('Bearer access');
});
