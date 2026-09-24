import { useState } from 'react';
import { useRouter } from 'expo-router';
import { NotificationInboxQueries } from '../src/application/notifications/NotificationInboxQueries';
import type { ExpirationNotification } from '../src/domain/notifications/Notification';
import { NotificationInboxScreen } from '../src/ui/screens/NotificationInboxScreen';
import { assetDetailHref } from '../src/ui/screens/AssetDetailNavigation';

/** Real inbox/application behavior; local paginated repository, no remote writes. */
export function NotificationJourneyFixture() {
  const router = useRouter();
  const [queries] = useState(() => {
    let rows: ExpirationNotification[] = Array.from({ length: 24 }, (_, index) => ({
      id: `notice-${index + 1}`, assetId: `filter-item-${index + 1}`,
      title: index === 23 ? 'Camping item 24 with a long descriptive name' : `Camping item ${String(index + 1).padStart(2, '0')}`,
      parentAssetId: '', customAssetTypeId: '',
      expiration: { date: '2027-04-15', precision: 'day' },
      milestone: 'upcoming', createdAt: '2027-03-01T00:00:00Z'
    }));
    const setRead = (id: string | undefined, read: boolean) => {
      rows = rows.map(row => id === undefined || row.id === id
        ? { ...row, readAt: read ? '2027-03-03T00:00:00Z' : undefined } : row);
    };
    return new NotificationInboxQueries({
      async listInbox(_tenant, _inventory, options) {
        const selected = rows.filter(row => !options?.unreadOnly || !row.readAt);
        const offset = Number(options?.cursor ?? 0); const limit = options?.limit ?? 20;
        const items = selected.slice(offset, offset + limit);
        const hasMore = offset + items.length < selected.length;
        return { items, pagination: { limit, hasMore, nextCursor: hasMore ? String(offset + items.length) : null } };
      },
      async getNotification(_tenant, _inventory, id) {
        const row = rows.find(row => row.id === id);
        if (!row) throw new Error('Unknown fixture notification');
        return row;
      },
      async markRead(_tenant, _inventory, id) { setRead(id, true); },
      async markUnread(_tenant, _inventory, id) { setRead(id, false); },
      async markAllReadPage() { setRead(undefined, true); return { complete: true, nextCursor: null }; },
      async countUnreadPage() { return { count: rows.filter(row => !row.readAt).length, nextCursor: null }; }
    }, { record() {} });
  });
  return <NotificationInboxScreen tenantId="audit-tenant" inventoryId="audit-inventory" queries={queries}
    onOpenAsset={id => router.push(assetDetailHref(id))} onChanged={() => {}}
    onSettings={() => router.push('/settings/inventory/notifications')} />;
}
