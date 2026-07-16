import { expect, test, type Locator, type Page } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

test('mobile Settings and Import controls retain 44-pixel touch targets', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Touch-target geometry runs at the mobile viewport.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/access');
  await expect(page.getByRole('heading', { name: 'Sharing' })).toBeVisible();
  await expectMinimumTarget(page.locator('#invite-email'));
  await expectMinimumTarget(page.getByRole('button', { name: 'Create invite' }));
  await expectMinimumTarget(page.getByRole('group', { name: 'Invitation access level' }).getByRole('button').first());
  await expectMinimumTarget(page.getByRole('group', { name: 'Invitation status' }).getByRole('link').first());

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');
  await expect(page.getByRole('heading', { name: 'Imports', exact: true })).toBeVisible();
  await expectMinimumTarget(page.getByRole('button', { name: 'Refresh' }));
  await expectMinimumTarget(page.getByRole('button', { name: 'New import' }));
  await expectMinimumTarget(page.getByLabel('Import history filters').getByRole('button').first());
  await expectMinimumTarget(page.getByRole('button', { name: /View details for/ }).first());
  await expectMinimumTarget(page.getByRole('button', { name: /Remove from history/ }).first());

  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)).toBe(true);
});

async function expectMinimumTarget(locator: Locator): Promise<void> {
  await expect(locator).toBeVisible();
  const box = await locator.boundingBox();
  expect(box, `Missing layout box for ${await locator.getAttribute('aria-label') ?? await locator.textContent()}`).not.toBeNull();
  expect(box?.width ?? 0).toBeGreaterThanOrEqual(44);
  expect(box?.height ?? 0).toBeGreaterThanOrEqual(44);
}
