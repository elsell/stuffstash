import { expect, it } from 'vitest';
import { NotificationInboxQueries } from './NotificationInboxQueries';
import type { InboxOptions } from './NotificationRepository';
const observer = { record() {} };
it('counts every page and follows sparse inbox pages with explicit scope', async () => {
  const calls: string[][] = [];
  const repository = {
    async countUnreadPage(tenant: string, inventory: string, cursor?: string) { calls.push([tenant, inventory, cursor ?? '']); return { count: cursor ? 2 : 0, nextCursor: cursor ? null : 'next' }; },
    async listInbox(_tenant: string, _inventory: string, options?: InboxOptions) { return { items: [], pagination: { limit: 30, hasMore: !options?.cursor, nextCursor: options?.cursor ? null : 'next' } }; },
    async getNotification() { throw new Error('Unavailable'); }, async markUnread() { throw new Error('Unread not exercised by this fixture'); }, async markRead() {}, async markAllReadPage() { return { complete: true, nextCursor: null }; }
  };
  const queries = new NotificationInboxQueries(repository, observer);
  expect(await queries.count('tenant', 'inventory')).toBe(2);
  expect(calls).toEqual([['tenant','inventory',''],['tenant','inventory','next']]);
  expect(await queries.list('tenant','inventory')).toMatchObject({ items: [], pagination: { hasMore: false } });
  await expect(queries.open('tenant','inventory','missing')).rejects.toThrow('Unavailable');
});
it('rejects cancelled counts using the native cancellation helper', async () => {
  const controller = new AbortController();
  const repository = {
    async countUnreadPage() { controller.abort(); return { count: 2, nextCursor: null }; },
    async listInbox() { throw new Error('unused'); }, async getNotification() { throw new Error('unused'); }, async markUnread() { throw new Error('Unread not exercised by this fixture'); }, async markRead() {}, async markAllReadPage() { return { complete: false, nextCursor: null }; }
  };
  const queries = new NotificationInboxQueries(repository, observer);
  await expect(queries.count('tenant','inventory', { signal: controller.signal })).rejects.toMatchObject({ name: 'AbortError' });
  await expect(queries.markAllRead('tenant','inventory')).rejects.toThrow('completed');
});
it('finishes all read pages and returns only a revalidated asset target', async () => {
  const actions: string[] = [];
  const repository = {
    async countUnreadPage() { return { count: 0, nextCursor: null }; },
    async listInbox() { return { items: [], pagination: { limit: 30, nextCursor: null, hasMore: false } }; },
    async markAllReadPage(_tenant: string, _inventory: string, cursor?: string) { actions.push(cursor ?? 'first'); return { complete: !!cursor, nextCursor: cursor ? null : 'last' }; },
    async getNotification() { actions.push('resolve'); return { id: 'notice', assetId: 'current-item', title: 'Bottle', parentAssetId: 'bin', customAssetTypeId: 'medicine', expiration: { date: '2026-10', precision: 'month' as const }, milestone: 'upcoming' as const, createdAt: '2026-09-10T12:00:00Z' }; },
    async markUnread() { throw new Error('Unread not exercised by this fixture'); }, async markRead() { actions.push('read'); }
  };
  const queries = new NotificationInboxQueries(repository, observer);
  await queries.markAllRead('tenant','inventory');
  expect(actions).toEqual(['first','last']);
  expect(await queries.open('tenant','inventory','notice')).toBe('current-item');
  expect(actions.slice(-2)).toEqual(['resolve','read']);
});
