import { expect, it } from 'vitest';
import { countUnreadNotifications, markAllNotificationsRead } from './notificationInboxBatch';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';

it('accumulates all count contributions including empty continued pages', async () => {
  const cursors: Array<string | undefined> = [];
  const repository = { async countUnreadPage(_tenant: string, _inventory: string, cursor?: string) {
    cursors.push(cursor);
    return cursor === 'last' ? { count: 3, nextCursor: null } : { count: 0, nextCursor: 'last' };
  } };
  await expect(countUnreadNotifications(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory')).resolves.toBe(3);
  expect(cursors).toEqual([undefined, 'last']);
});
it('continues mark-all to completion and rejects broken completion signals', async () => {
  let calls = 0;
  const repository = { async markAllReadPage() { calls++; return { complete: calls === 2, nextCursor: calls === 2 ? null : 'next' }; } };
  await markAllNotificationsRead(repository, new InMemoryWorkspaceObserver(), 'tenant', 'inventory');
  expect(calls).toBe(2);
  await expect(markAllNotificationsRead({ async markAllReadPage() { return { complete: false, nextCursor: null }; } }, new InMemoryWorkspaceObserver(), 'tenant', 'inventory')).rejects.toThrow('completed');
});
it('rejects repeated count cursors and cancelled results', async () => {
  await expect(countUnreadNotifications({ async countUnreadPage() { return { count: 1, nextCursor: 'same' }; } }, new InMemoryWorkspaceObserver(), 'tenant', 'inventory')).rejects.toThrow('completed');
  const controller = new AbortController();
  await expect(countUnreadNotifications({ async countUnreadPage() { controller.abort(); return { count: 1, nextCursor: null }; } }, new InMemoryWorkspaceObserver(), 'tenant', 'inventory', controller.signal)).rejects.toThrow();
});
