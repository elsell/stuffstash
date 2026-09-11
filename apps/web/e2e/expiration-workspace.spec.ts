import { expect, test } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

test('keeps read and unread distinct and makes personal reminder editing reachable', async ({page}, testInfo) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
  let read = false;
  let preferences = {revision: 1, defaults: {enabled: true, upcoming: true, expired: true, advanceDays: 30}, timezone: 'America/New_York', pushEnabled: false, overrides: []};
  const notification = () => ({id: 'notice', assetId: 'medicine', title: 'Tylenol', customAssetTypeId: 'medicine', expirationDate: '2026-10', expirationPrecision: 'month', milestone: 'upcoming', createdAt: '2026-09-11T12:00:00Z', readAt: read ? '2026-09-11T12:01:00Z' : null});
  await page.route('http://127.0.0.1:18080/**', async route => {
    const url = new URL(route.request().url());
    if (!url.pathname.includes('/notification')) return route.fallback();
    let data: unknown;
    if (url.pathname.includes('/notification-preferences')) {
      if (route.request().method() === 'PUT') preferences = {...preferences, ...route.request().postDataJSON(), revision: preferences.revision + 1};
      data = preferences;
    } else if (url.pathname.endsWith('/unread-count')) data = {count: read ? 0 : 1};
    else if (url.pathname.endsWith('/read')) {read = route.request().method() !== 'DELETE'; data = {id: 'notice', read};}
    else if (url.pathname.endsWith('/notice')) data = notification();
    else data = url.searchParams.get('unreadOnly') === 'true' && read ? [] : [notification()];
    await route.fulfill({json: {data, meta: {pagination: {limit: 30, hasMore: false, nextCursor: null}}}});
  });
  await page.goto('/');
  await page.getByRole('button', {name: 'Notifications, 1 unread', exact: true}).click();
  await expect(page.getByRole('heading', {name: 'Notifications', exact: true})).toHaveCount(1);
  await expect(page.getByRole('button', {name: 'Mark Tylenol read', exact: true})).toBeVisible();
  await page.screenshot({path: testInfo.outputPath('inbox-unread.png'), fullPage: true});
  await page.getByRole('button', {name: 'Mark Tylenol read', exact: true}).click();
  await expect(page.getByRole('button', {name: 'Mark Tylenol unread', exact: true})).toBeVisible();
  await page.screenshot({path: testInfo.outputPath('inbox-read.png'), fullPage: true});
  await page.getByRole('button', {name: 'Mark Tylenol unread', exact: true}).click();
  await expect(page.getByRole('button', {name: 'Mark Tylenol read', exact: true})).toBeVisible();
  await page.getByRole('button', {name: 'Notification settings', exact: true}).click();
  await expect(page.getByRole('heading', {name: 'Expiration reminders', exact: true})).toBeVisible();
  await page.getByRole('button', {name: /Before expiration/}).click();
  await page.getByLabel('Days before expiration', {exact: true}).fill('14');
  await page.getByRole('button', {name: 'Save reminders', exact: true}).click();
  await expect.poll(() => preferences.defaults.advanceDays).toBe(14);
  await page.screenshot({path: testInfo.outputPath('reminder-settings.png'), fullPage: true});
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});
