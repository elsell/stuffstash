import { expect, it } from 'vitest';
import { OpenPushNotification } from '../../application/notifications/OpenPushNotification';
import { ExpoPushNotificationResponses, type NativePushResponse } from './ExpoPushNotificationResponses';

const routing = { serverId: 'https://api.test', principalId: 'user', tenantId: 'tenant', inventoryId: 'inventory', notificationId: 'notice' };
const directResponse = (payload: unknown, type = 'push', data: unknown = null): NativePushResponse => ({
  actionIdentifier: 'default',
  notification: { request: { identifier: 'delivery', content: { data }, trigger: { type, payload } } }
});

async function openResponse(response: NativePushResponse, launch: boolean) {
  const calls: string[] = [];
  let receive!: (response: NativePushResponse) => void;
  let finish!: () => void;
  const settled = new Promise<void>(resolve => { finish = resolve; });
  const command = new OpenPushNotification('https://api.test', {
    async getCurrentPrincipal() { return { id: 'user' }; }
  }, {
    async open(tenant, inventory, notice) { calls.push(`${tenant}:${inventory}:${notice}`); return 'authorized-item'; }
  }, {
    async execute(id) { calls.push(`select:${id}`); return { selectedInventoryId: id }; }
  });
  const source = new ExpoPushNotificationResponses({
    async getLastNotificationResponseAsync() { return launch ? response : null; },
    addNotificationResponseReceivedListener(listener) { receive = listener; return { remove() {} }; }
  }, 'default');
  let failure: unknown;
  const stop = source.subscribe(async payload => {
    try { calls.push(await command.execute(payload)); }
    catch (error) { failure = error; }
    finally { finish(); }
  }, () => { throw new Error('Unexpected response source failure'); });
  if (!launch) receive(response);
  await settled;
  stop();
  return { calls, failure };
}

it.each([true, false])('opens direct APNs routing through account validation on launch=%s', async launch => {
  const result = await openResponse(directResponse({ aps: { alert: { title: 'Inventory reminder' } }, ...routing, assetId: 'forged' }), launch);
  expect(result.failure).toBeUndefined();
  expect(result.calls).toEqual(['tenant:inventory:notice', 'select:inventory', 'authorized-item']);
});

it('preserves ordinary content data instead of merging raw routing fields', async () => {
  expect((await openResponse(directResponse({ ...routing, principalId: 'other' }, 'push', routing), false)).failure).toBeUndefined();
  const result = await openResponse(directResponse(routing, 'push', {}), false);
  expect(result.calls).toEqual([]);
  expect(result.failure).toMatchObject({ kind: 'invalid-payload' });
});

it.each([
  directResponse(routing, 'calendar'),
  directResponse({ ...routing, serverId: 'https://other.test' }),
  directResponse({ ...routing, principalId: 'other' }),
  directResponse({ ...routing, notificationId: ['notice'] }),
  directResponse({ ...routing, inventoryId: '' })
])('rejects untrusted or malformed native hints before inventory access', async response => {
  const result = await openResponse(response, false);
  expect(result.failure).toBeDefined();
  expect(result.calls).toEqual([]);
});
