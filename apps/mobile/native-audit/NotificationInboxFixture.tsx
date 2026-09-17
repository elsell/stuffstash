import { useState } from 'react';
import { ScrollView, Text } from 'react-native';
import { Stack, useLocalSearchParams, useRouter, type Href } from 'expo-router';
import { NotificationInboxQueries } from '../src/application/notifications/NotificationInboxQueries';
import { NotificationFailure } from '../src/application/notifications/NotificationFailure';
import type { ExpirationNotification } from '../src/domain/notifications/Notification';
import { NotificationInboxScreen } from '../src/ui/screens/NotificationInboxScreen';
import { useAppearancePalette } from '../src/ui/theme/AppearanceContext';

const initialRows: ExpirationNotification[] = [
  { id: 'medicine', parentAssetId: 'box', customAssetTypeId: 'medicine', assetId: 'medicine-item', title: 'Household medicine with a long descriptive label',
    expiration: { date: '2027-02', precision: 'month' }, milestone: 'expired', createdAt: '2027-03-01T00:00:00Z',
    parentTrail: [{ assetId: 'closet', title: 'Hall medicine closet', kind: 'location' }, { assetId: 'box', title: 'Cold and cough supplies', kind: 'container' }] },
  { id: 'batteries', parentAssetId: '', customAssetTypeId: 'supplies', assetId: 'battery-item', title: 'Emergency batteries',
    expiration: { date: '2027-04-15', precision: 'day' }, milestone: 'upcoming', createdAt: '2027-03-01T00:00:00Z', readAt: '2027-03-02T00:00:00Z' }
];

export function NotificationInboxFixture() {
  const { scenario } = useLocalSearchParams<{ scenario?: string }>();
  return <InboxScenario key={scenario === 'denied' ? 'denied' : 'normal'} denied={scenario === 'denied'} />;
}
function InboxScenario({ denied }: { readonly denied: boolean }) {
  const router = useRouter();
  const [queries] = useState(() => {
    let rows = initialRows.map(row => ({ ...row })); let rejectNextMutation = denied;
    const mutate = (id: string | undefined, read: boolean) => {
      if (rejectNextMutation) { rejectNextMutation = false; throw new NotificationFailure('permission-denied'); }
      rows = rows.map(row => !id || row.id === id ? { ...row, readAt: read ? '2027-03-03T00:00:00Z' : undefined } : row);
    };
    return new NotificationInboxQueries({
      async listInbox(_tenant, _inventory, options) { return { items: rows.filter(row => !options?.unreadOnly || !row.readAt), pagination: { limit: 20, hasMore: false, nextCursor: null } }; },
      async getNotification(_tenant, _inventory, id) { const row = rows.find(entry => entry.id === id); if (!row) throw new NotificationFailure('not-found'); return row; },
      async markRead(_tenant, _inventory, id) { mutate(id, true); },
      async markUnread(_tenant, _inventory, id) { mutate(id, false); },
      async markAllReadPage() { mutate(undefined, true); return { complete: true, nextCursor: null }; },
      async countUnreadPage() { return { count: rows.filter(row => !row.readAt).length, nextCursor: null }; }
    }, { record() {} });
  });
  const open = (target: string) => router.push({ pathname: '/audit-notification-target', params: { target } } as Href);
  return <NotificationInboxScreen tenantId="audit-tenant" inventoryId="audit-inventory" queries={queries}
    onOpenAsset={open} onChanged={() => {}} onSettings={() => open('Reminder settings')} />;
}
export function NotificationTargetFixture() {
  const { target } = useLocalSearchParams<{ target: string }>();
  const colors = useAppearancePalette();
  return <><Stack.Screen options={{ title: 'Notification destination' }} /><ScrollView contentInsetAdjustmentBehavior="automatic" style={{ backgroundColor: colors.background }}>
    <Text style={{ color: colors.text, padding: 24 }}>Resolved target: {target}</Text>
  </ScrollView></>;
}
