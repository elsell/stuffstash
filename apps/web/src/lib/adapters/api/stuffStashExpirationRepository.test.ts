import { expect, it } from 'vitest';
import { StuffStashExpirationRepository } from './stuffStashExpirationRepository';
const item = { id: 'item', tenantId: 'tenant', inventoryId: 'inventory', kind: 'item', lifecycleState: 'active', title: 'Solution', description: '', expiration: { date: '2028-02', precision: 'month' }, expirationContext: { state: 'current', trackingEnabled: false, advanceDays: 7, timezone: 'UTC' }, ancestorPath: [{ id: 'cabinet', title: 'Cabinet' }], tags: [], customFields: {}, createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' };
function repository(value: unknown) { return new StuffStashExpirationRepository('https://api.test', () => 'access', async () => Response.json({ data: { items: [value], counts: { soon: 0, expired: 0, all: 1 }, timezone: 'UTC' }, meta: { pagination: { limit: 30, hasMore: false, nextCursor: null } } })); }
it('maps month precision, disabled tracking and placement into web domain', async () => {
 const page = await repository(item).list('tenant', 'inventory', { mode: 'all' });
 expect(page.items[0].expiration).toEqual(item.expiration); expect(page.items[0].expirationContext?.trackingEnabled).toBe(false); expect(page.items[0].ancestorPath).toEqual([{ id: 'cabinet', title: 'Cabinet' }]);
});
it.each([{ ...item, tenantId: 'foreign' }, { ...item, inventoryId: 'foreign' }, { ...item, lifecycleState: 'archived' }, { ...item, expiration: null }])('rejects wrong scope or invalid expiry records', async value => {
 await expect(repository(value).list('tenant', 'inventory', { mode: 'all' })).rejects.toThrow();
});
