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
