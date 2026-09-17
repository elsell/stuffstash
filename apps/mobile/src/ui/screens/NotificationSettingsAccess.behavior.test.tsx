import React from 'react';
import { expect, it } from 'vitest';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationRepository } from '../../adapters/notifications/ApiNotificationRepository';
import { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import { CustomizationFailure } from '../../application/customization/CustomizationErrors';
import { MobileRenderHarness } from '../../test-support/render';
import { NotificationSettingsScreen } from './NotificationSettingsScreen';

const cases: { status: number; operation: 'refresh' | 'save' | 'types' }[] = [
  ...[401, 403].flatMap(status => (['refresh', 'save'] as const).map(operation => ({ status, operation }))),
  { status: 403, operation: 'types' },
];
it.each(cases)('retires cached settings after $status during $operation until an authorized reload', async ({ status, operation }) => {
  const h = new MobileRenderHarness(); let responseStatus = 200; let denyTypes = false; let writes = 0;
  let preferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'America/New_York', pushEnabled: false, overrides: [] };
  const client = new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'synthetic-token', fetch: async (input, init) => {
    const request = new Request(input, init);
    if (request.method === 'PUT') writes++;
    if (responseStatus !== 200) return Response.json({ error: { code: 'failure', message: 'Private transport detail' } }, { status: responseStatus });
    if (request.method === 'PUT') preferences = { ...preferences, ...await request.json(), revision: preferences.revision + 1 };
    return Response.json({ data: preferences, meta: {} });
  } });
  const session = new NotificationPreferencesSession(new ApiNotificationRepository(client), { record() {} }, 'tenant', 'inventory');
  const refresh = async () => {
    const control = h.byType('ScrollView')!.props.refreshControl as React.ReactElement<{ onRefresh: () => void }>;
    await h.run(() => control.props.onRefresh()); await h.settle();
  };
  try {
    await h.render(<NotificationSettingsScreen tenantId="tenant" inventoryId="inventory" session={session} assetTypesQuery={{ async execute() {
      if (denyTypes) throw new CustomizationFailure('permission-denied');
      return [];
    } }} onNavigate={() => {}} onBack={() => {}} />);
    await h.settle();
    expect(h.byLabel('Default reminders')).toBeDefined();
    const retainedChange = h.byLabel('Default reminders')!.props.onValueChange;
    responseStatus = 503; await refresh();
    expect(h.byLabel('Default reminders')).toBeDefined();
    responseStatus = status;
    if (operation === 'save') { await h.run(() => retainedChange(false)); await h.settle(); }
    else { denyTypes = operation === 'types'; await refresh(); }
    expect(h.byLabel('Default reminders')).toBeUndefined();
    expect(h.byLabel('Time zone')).toBeUndefined();
    expect(h.allText().join(' ')).not.toContain('Private transport detail');
    const deniedWrites = writes;
    await h.run(() => retainedChange(true)); await h.settle();
    expect(writes).toBe(deniedWrites);
    denyTypes = false; responseStatus = 503;
    await h.press(h.byLabel('Retry loading reminders')); await h.settle();
    expect(h.byLabel('Default reminders')).toBeUndefined();
    responseStatus = 200;
    await h.press(h.byLabel('Retry loading reminders')); await h.settle();
    expect(h.byLabel('Default reminders')).toBeDefined();
    await h.run(() => h.byLabel('Default reminders')!.props.onValueChange(false)); await h.settle();
    expect(writes).toBe(deniedWrites + 1);
    expect(preferences.defaults.enabled).toBe(false);
  } finally { await h.unmount(); }
});
