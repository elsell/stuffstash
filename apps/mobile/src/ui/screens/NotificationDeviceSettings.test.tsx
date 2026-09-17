import React from 'react';
import { expect, it } from 'vitest';
import { StuffStashClient } from '@stuff-stash/api-client';
import { ApiNotificationRepository } from '../../adapters/notifications/ApiNotificationRepository';
import { NotificationPreferencesSession } from '../../application/notifications/NotificationPreferencesSession';
import { MobileRenderHarness } from '../../test-support/render';
import { deviceSettingsFake, setAppStateForTest } from '../../test-support/react-native';
import { setScreenFocused } from '../../test-support/navigation';
import { NotificationSettingsScreen } from './NotificationSettingsScreen';

const failureMessage = 'Device settings could not be opened. Open Settings on your device and choose Stuff Stash to change notification access, or try again.';
it.each(['current', 'returned', 'background'].flatMap(context => (['enabled', 'denied'] as const).map(outcome => ({ context, outcome }))))('owns device Settings failure and retry in $context after $outcome', async ({ context, outcome }) => {
  const h = new MobileRenderHarness(); deviceSettingsFake.reset();
  const preferences = { revision: 1, defaults: { enabled: true, upcoming: true, expired: true, advanceDays: 30 }, timezone: 'UTC', pushEnabled: true, overrides: [] };
  let writes = 0;
  let setups = 0;
  const settingsLabel = outcome === 'enabled' ? 'Open device settings' : 'Open notification system settings';
  const repository = new ApiNotificationRepository(new StuffStashClient({ baseUrl: 'https://api.test', tokenProvider: () => 'token', fetch: async (input, init) => {
    if (new Request(input, init).method !== 'GET') writes++;
    return Response.json({ data: preferences, meta: {} });
  } }));
  const session = new NotificationPreferencesSession(repository, { record() {} }, 'tenant', 'inventory');
  let reject!: (error: Error) => void;
  deviceSettingsFake.completion = new Promise<void>((_, fail) => { reject = fail; });
  try {
    await h.render(<NotificationSettingsScreen tenantId="tenant" inventoryId="inventory" session={session}
      assetTypesQuery={{ async execute() { return []; } }} pushSession={{ async enable() { setups++; return outcome; } }} onBack={() => {}} onNavigate={() => {}} />);
    await h.settle();
    await h.press(h.byLabel('Set up alerts on this device')); await h.settle();
    const writesBeforeLaunch = writes;
    await h.press(h.byLabel(settingsLabel));
    expect(h.byLabel(settingsLabel)?.props.disabled).toBe(true);
    await h.press(h.byLabel(settingsLabel));
    expect(deviceSettingsFake.attempts).toBe(1);
    if (context === 'returned') { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); await h.settle(); }
    if (context === 'background') { await h.run(() => setAppStateForTest('background')); await h.run(() => setAppStateForTest('active')); }
    let finishNew!: () => void;
    if (context === 'returned') {
      deviceSettingsFake.completion = new Promise<void>(resolve => { finishNew = resolve; });
      await h.press(h.byLabel('Set up alerts on this device')); await h.settle();
      await h.press(h.byLabel(settingsLabel));
    }
    await h.run(() => reject(new Error('private native launch detail'))); await h.settle();
    expect(h.allText()).not.toContain('private native launch detail');
    if (context === 'current') expect(h.byText(failureMessage)).toBeDefined();
    else expect(h.byText(failureMessage)).toBeUndefined();
    if (context === 'returned') {
      expect(h.byLabel(settingsLabel)?.props.disabled).toBe(true);
      await h.run(() => finishNew()); await h.settle();
    }
    deviceSettingsFake.completion = undefined;
    if (context === 'background') { await h.press(h.byLabel('Set up alerts on this device')); await h.settle(); }
    await h.press(h.byLabel(settingsLabel)); await h.settle();
    expect(deviceSettingsFake.attempts).toBe(context === 'returned' ? 3 : 2);
    expect(h.byText(failureMessage)).toBeUndefined();
    expect(writes).toBe(writesBeforeLaunch);
    expect(setups).toBe(context === 'current' ? 1 : 2);
  } finally { await h.unmount(); setScreenFocused(true); setAppStateForTest('active'); deviceSettingsFake.reset(); }
});
