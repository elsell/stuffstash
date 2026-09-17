import React from 'react';
import { expect, it } from 'vitest';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationRepository } from '../../adapters/notifications/ApiNotificationRepository';
import { NotificationInboxQueries } from '../../application/notifications/NotificationInboxQueries';
import { MobileRenderHarness } from '../../test-support/render';
import { NotificationInboxScreen } from './NotificationInboxScreen';

const operations = ['refresh', 'open', 'read-state', 'mark-all'] as const;
const cases = [401, 403].flatMap(status => operations.map(operation => ({ status, operation })));

it.each(cases)('clears denied inbox data after $status during $operation and recovers only after a successful read', async ({ status, operation }) => {
  const h = new MobileRenderHarness(); let responseStatus = 200;
  const opened: string[] = [];
  const api = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'synthetic-token', fetch: async () => {
    if (responseStatus !== 200) return Response.json({ error: { code: 'failure', message: 'Private transport detail' } }, { status: responseStatus });
    return Response.json({ data: [{ id: 'notice', assetId: 'item', title: 'Private medicine', expirationDate: '2027-02', expirationPrecision: 'month', milestone: 'upcoming', createdAt: '2027-01-01T00:00:00Z' }], meta: { pagination: { limit: 20, hasMore: true, nextCursor: 'next' } } });
  } });
  const queries = new NotificationInboxQueries(new ApiNotificationRepository(api), { record() {} });
  const settle = () => h.run(() => new Promise(resolve => setTimeout(resolve, 10)));
  try {
    await h.render(<NotificationInboxScreen tenantId="tenant" inventoryId="inventory" queries={queries} onOpenAsset={id => opened.push(id)} onChanged={() => {}} onSettings={() => {}} />);
    await settle();
    expect(h.byLabel('Open Private medicine')).toBeDefined();
    expect(h.byLabel('Load more notifications')).toBeDefined();
    responseStatus = 503;
    const ordinaryRefresh = h.byType('ScrollView')!.props.refreshControl as React.ReactElement<{ onRefresh: () => void }>;
    await h.run(() => ordinaryRefresh.props.onRefresh()); await settle();
    expect(h.byLabel('Open Private medicine')).toBeDefined();
    responseStatus = status;
    if (operation === 'refresh') {
      const refresh = h.byType('ScrollView')!.props.refreshControl as React.ReactElement<{ onRefresh: () => void }>;
      await h.run(() => refresh.props.onRefresh());
    } else {
      await h.press(h.byLabel(operation === 'open' ? 'Open Private medicine' : operation === 'read-state' ? 'Mark Private medicine read' : 'Mark all read'));
    }
    await settle();
    expect(h.allText().join(' ')).not.toContain('Private medicine');
    expect(h.byLabel('Load more notifications')).toBeUndefined();
    expect(h.allText().join(' ')).not.toContain('No notifications yet');
    expect(h.allText().join(' ')).not.toContain('Private transport detail');
    expect(opened).toEqual([]);
    responseStatus = 503;
    await h.press(h.byLabel('Retry notifications')); await settle();
    expect(h.allText().join(' ')).not.toContain('Private medicine');
    responseStatus = 200;
    await h.press(h.byLabel('Retry notifications')); await settle();
    expect(h.byLabel('Open Private medicine')).toBeDefined();
  } finally { await h.unmount(); }
});
