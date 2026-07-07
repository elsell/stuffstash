import { expect, test } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

test('desktop import surface scans like durable job history', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Desktop import UX coverage runs on the desktop project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');

  await expect(page.getByRole('heading', { name: 'Imports', exact: true })).toBeVisible();
  await expect(page.getByText('In progress')).toBeVisible();
  await expect(page.getByText('Importing photos and files')).toBeVisible();
  await expect(page.getByText('1 field created')).toBeVisible();
  const completedRow = page.locator('.history-row').filter({ hasText: '1 field created' });
  await expect(completedRow.getByText('Completed', { exact: true })).toBeVisible();
  await expect(completedRow.getByText('Started Jul 6, 2026')).toBeVisible();

  const discardedRow = page.locator('.history-row').filter({ hasText: 'Partial progress discarded' });
  await expect(discardedRow).toBeVisible();
  await discardedRow.getByRole('button', { name: 'Details' }).click();
  await expect(page.getByText('Records created by this job were discarded. Audit history remains.')).toBeVisible();
  await expect(page.locator('a.resource-link')).toHaveCount(0);
});

test('mobile import setup keeps one-column flow and subordinate connection options', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile import UX coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');

  await expect(page.getByRole('heading', { name: 'Imports', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'New import' }).first().click();
  await expect(page.getByText('Choose import method')).toBeVisible();
  await expect(page.getByRole('link', { name: /Connect to Homebox/ })).toBeVisible();
  await expect(page.getByRole('link', { name: /Upload Homebox CSV/ })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)).toBe(true);

  await page.getByRole('link', { name: /Connect to Homebox/ }).click();
  await expect(page.getByText('Connect to Homebox')).toBeVisible();
  const connectionOptions = page.locator('details').filter({ hasText: 'Connection options' });
  await expect(connectionOptions).toHaveJSProperty('open', false);
  await connectionOptions.locator('summary').click();
  await expect(page.getByLabel('Allow private-network Homebox URL')).toBeVisible();
  await expect(page.getByLabel('Allow self-signed TLS certificate')).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)).toBe(true);
});
