import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationRepository } from '../../adapters/notifications/ApiNotificationRepository';
import { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import { NotificationSettingsScreen } from './NotificationSettingsScreen';
it('loads personal settings and saves a changed threshold without changing timezone', async () => {
  let preferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'America/New_York', pushEnabled: false, overrides: [] };
  const repository = new ApiNotificationRepository(new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async (input, init) => {
    const request = new Request(input, init);
    if (request.method === 'PUT') preferences = { ...preferences, ...await request.json(), revision: preferences.revision + 1 };
    return Response.json({ data: preferences, meta: {} });
  }}));
  const session = new NotificationPreferencesSession(repository, { record() {} }, 'tenant','inventory');
  await session.initialize('America/New_York');
  const harness = new MobileRenderHarness();
  try {
    await harness.render(<NotificationSettingsScreen tenantId="tenant" inventoryId="inventory" session={session} assetTypesQuery={{ async execute() { return []; } }} />);
    await harness.settle();
    expect(harness.byLabel('Timezone')?.props.value).toBe('America/New_York');
    expect(harness.byLabel('Days before expiration')).toBeDefined();
    await harness.changeText(harness.byLabel('Days before expiration'), '7');
    await harness.press(harness.byLabel('Save reminders'));
    expect(preferences.defaults.advanceDays).toBe(7);
    expect(preferences.timezone).toBe('America/New_York');
  } finally { await harness.unmount(); }
});
