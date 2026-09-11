import { expect, it } from 'vitest';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationRepository } from './ApiNotificationRepository';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
it('maps expiration precision and keeps empty continued pages', async () => {
  let empty = false;
  const api = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async () => Response.json({ data: empty ? [] : [{ id: 'notice', assetId: 'item', title: 'Bottle', parentAssetId: 'bin', customAssetTypeId: 'medicine', expirationDate: '2026-10', expirationPrecision: 'month', milestone: 'upcoming', createdAt: '2026-09-10T12:00:00Z' }], meta: { pagination: { limit: 30, nextCursor: 'next', hasMore: true } } }) });
  const repository = new ApiNotificationRepository(api);
  expect((await repository.listInbox('tenant', 'inventory')).items[0].expiration).toEqual({ date: '2026-10', precision: 'month' });
  empty = true;
  await expect(repository.listInbox('tenant', 'inventory')).resolves.toMatchObject({ items: [], pagination: { nextCursor: 'next', hasMore: true } });
});
it.each([[401,'authentication-required'],[403,'permission-denied'],[404,'not-found'],[409,'conflict'],[422,'invalid'],[503,'unavailable']] as const)('sanitizes notification API status %s', async (status, kind) => {
  const api = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async () => Response.json({ error: { code: 'failure', message: 'private backend details' } }, { status }) });
  const result = await new ApiNotificationRepository(api).getPreferences('tenant', 'inventory').catch((error: unknown) => error);
  expect(result).toBeInstanceOf(NotificationFailure);
  expect(result).toMatchObject({ kind });
  expect((result as Error).message).not.toContain('private');
});
it('keeps cancellation distinct from a visible notification failure', async () => {
  const controller = new AbortController();
  const api = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async () => {
    controller.abort();
    return Response.json({ data: { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'UTC', pushEnabled: false, overrides: [] }, meta: {} });
  }});
  const error = await new ApiNotificationRepository(api).getPreferences('tenant', 'inventory', controller.signal).catch((caught: unknown) => caught);
  expect(error).not.toBeInstanceOf(NotificationFailure);
  expect(error).toMatchObject({ name: 'AbortError' });
});
it('preserves current parent kinds and partial paths from the inbox API',async()=>{
 const parentTrail=[{assetId:'closet',title:'Hall closet',kind:'location'},{assetId:'bin',title:'Bin 8',kind:'container'}];
 const api=new StuffStashClient({baseUrl:'https://api.test',tokenProvider:()=> 'token',fetch:async()=>Response.json({data:{id:'notice',assetId:'item',title:'Bottle',parentAssetId:'bin',customAssetTypeId:'medicine',expirationDate:'2028-02',expirationPrecision:'month',milestone:'upcoming',createdAt:'2028-01-01T00:00:00Z',parentTrail,parentTrailIncomplete:true},meta:{}})});
 const notification=await new ApiNotificationRepository(api).getNotification('tenant','inventory','notice');
 expect(notification.parentTrail).toEqual(parentTrail);expect(notification.parentTrailIncomplete).toBe(true);
});
