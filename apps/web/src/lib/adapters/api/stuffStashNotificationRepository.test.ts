import { expect, it } from 'vitest';
import { StuffStashNotificationRepository } from './stuffStashNotificationRepository';

it('maps inbox dates with their precision and retains empty continuation pages', async () => {
  let empty = false;
  const notice = { id: 'notice', assetId: 'bottle', title: 'Pain relief', parentAssetId: 'bin', customAssetTypeId: 'medicine', expirationDate: '2026-10', expirationPrecision: 'month', milestone: 'upcoming', createdAt: '2026-09-10T12:00:00Z' };
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async () => Response.json({ data: empty ? [] : [notice], meta: { pagination: { limit: 30, nextCursor: 'next', hasMore: true } } }));
  const page = await repository.listInbox('tenant', 'inventory');
  expect(page.items[0]).toEqual({ id: 'notice', assetId: 'bottle', title: 'Pain relief', parentAssetId: 'bin', customAssetTypeId: 'medicine', expiration: { date: '2026-10', precision: 'month' }, milestone: 'upcoming', createdAt: notice.createdAt, readAt: undefined });
  empty = true;
  await expect(repository.listInbox('tenant', 'inventory', { cursor: 'next' })).resolves.toEqual({ items: [], pagination: { limit: 30, nextCursor: 'next', hasMore: true } });
});

it('preserves personal preferences and full type overrides', async () => {
  const policy = { enabled: false, upcoming: true, expired: false, advanceDays: 14 };
  const preferences = { revision: 4, defaults: policy, timezone: 'America/New_York', pushEnabled: false, overrides: [{ customAssetTypeId: 'medicine', settings: { ...policy, enabled: true } }] };
  const repository = new StuffStashNotificationRepository('https://api.test', () => 'token', async () => Response.json({ data: preferences, meta: {} }));
  await expect(repository.getPreferences('tenant', 'inventory')).resolves.toEqual(preferences);
});
