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
  expect(await hasHorizontalOverflow(page.locator('.import-workspace'))).toBe(false);
  await expect(page.getByText('In progress')).toBeVisible();
  await expect(page.getByText('Importing photos and files')).toBeVisible();
  await expect(page.getByText('Prepared by owner@example.com').first()).toBeVisible();
  await expect(page.getByText('stuff.jsksell.com').first()).toBeVisible();
  await expect(page.getByText('/api/v1')).toHaveCount(0);
  await expect(page.getByRole('button', { name: /Warnings/ })).toBeVisible();
  await expect(page.getByRole('button', { name: /Action required/ })).toHaveCount(0);
  const warningRow = warningImportRow(page);
  await expect(warningRow.getByText('Completed with warnings.')).toHaveCount(0);
  await expect(warningRow.getByText('Completed')).toBeVisible();
  await expect(warningRow.locator('[data-slot="badge"]').filter({ hasText: 'Warnings' })).toBeVisible();
  await expect(warningRow.getByRole('cell', { name: /Jul 6, 2026, 7:15 AM/ })).toBeVisible();
  expect(await hasHorizontalOverflow(warningRow)).toBe(false);

  await warningRow.getByRole('button', { name: /review details/i }).click();
  await expect(page.getByText('Asset appears to have already been imported')).toBeVisible();
  await expect(page.getByText('Prepared by owner@example.com')).toBeVisible();
  await expect(page.getByText('Already linked to an earlier import')).toBeVisible();
  await expect(page.getByText('Source ID source-wardrobe')).toBeVisible();
  await expect(page.getByText('Source ID source-baby-hats')).toBeVisible();
  expect(await hasHorizontalOverflow(page.locator('.import-detail-content'))).toBe(false);
  await page.evaluate(() => {
    (window as Window & { __importNavigationSentinel?: string }).__importNavigationSentinel = 'same-document';
  });
  await page.getByRole('button', { name: /More/ }).click();
  await page.getByRole('menuitem', { name: /Open inventory activity/ }).click();
  await expect(page).toHaveURL(/\/settings\/activity/);
  expect(await importNavigationSentinel(page)).toBe('same-document');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');
  await warningRow.getByRole('button', { name: /review details/i }).click();
  await page.getByRole('tab', { name: 'Records' }).click();
  const importedAssetRow = page.getByRole('row').filter({ hasText: 'Imported asset' });
  await expect(importedAssetRow).toBeVisible();
  await page.evaluate(() => {
    (window as Window & { __importNavigationSentinel?: string }).__importNavigationSentinel = 'same-document';
  });
  await importedAssetRow.getByRole('link', { name: /Open Imported asset/ }).click();
  await expect(page).toHaveURL(/\/assets\/asset-tomato$/);
  expect(await importNavigationSentinel(page)).toBe('same-document');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');
  await warningRow.getByRole('button', { name: /review details/i }).click();
  await page.getByRole('button', { name: 'Back to history' }).click();

  const discardedRow = page.locator('.history-ledger .history-row').filter({ hasText: 'Partial progress discarded' });
  await expect(discardedRow).toBeVisible();
  await discardedRow.getByRole('button', { name: /details/i }).click();
  await page.getByRole('tab', { name: 'Records' }).click();
  await expect(page.getByText('Records created by this job were discarded. Audit history remains.')).toBeVisible();
  await expect(page.locator('a.resource-link')).toHaveCount(0);

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');
  await page.getByRole('button', { name: 'New import' }).first().click();
  await expect(page.getByText('Choose import method')).toBeVisible();
  await expectImportStepperGeometry(page);
});

test('mobile import setup keeps one-column flow and subordinate connection options', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile import UX coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import');

  await expect(page.getByRole('heading', { name: 'Imports', exact: true })).toBeVisible();
  expect(await hasHorizontalOverflow(page.locator('.import-workspace'))).toBe(false);
  const warningRow = warningImportRow(page);
  await warningRow.getByRole('button', { name: /review details/i }).click();
  await expect(page.getByText('Source ID source-wardrobe')).toBeVisible();
  expect(await hasHorizontalOverflow(page.locator('.import-detail-content'))).toBe(false);
  await page.getByRole('button', { name: 'Back to history' }).click();

  await page.getByRole('button', { name: 'New import' }).first().click();
  await expect(page.getByText('Choose import method')).toBeVisible();
  await expectImportStepperGeometry(page);
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

  await page.getByRole('textbox', { name: 'Homebox URL' }).fill('stuff.jsksell.com');
  await page.getByRole('textbox', { name: 'Email' }).fill('codex@jsksell.com');
  await page.getByRole('textbox', { name: 'Password' }).fill('asldfj3290f!');
  const actionRow = page.locator('.import-source-setup-content .action-row');
  await actionRow.scrollIntoViewIfNeeded();
  await expect(page.getByRole('button', { name: 'Confirm connection' })).toBeEnabled();
  await expect(actionRow).toBeInViewport();
  const [actionBox, mobileNavBox] = await Promise.all([
    actionRow.boundingBox(),
    page.locator('.mobile-nav').boundingBox()
  ]);
  expect(actionBox && mobileNavBox ? actionBox.y + actionBox.height <= mobileNavBox.y - 8 : false).toBe(true);
});

async function hasHorizontalOverflow(locator: import('@playwright/test').Locator): Promise<boolean> {
  return locator.evaluate((element) => element.scrollWidth > element.clientWidth + 1);
}

async function importNavigationSentinel(page: import('@playwright/test').Page): Promise<string | undefined> {
  return page.evaluate(() => (window as Window & { __importNavigationSentinel?: string }).__importNavigationSentinel);
}

function warningImportRow(page: import('@playwright/test').Page): import('@playwright/test').Locator {
  return page
    .locator('.history-ledger .history-row')
    .filter({ hasText: 'Warnings' })
    .filter({ hasText: '1 location saved · 2 assets saved · 1 photo/file saved' });
}

async function expectImportStepperGeometry(page: import('@playwright/test').Page): Promise<void> {
  const stepper = page.getByRole('list', { name: 'Import progress' });
  await expect(stepper).toBeVisible();
  expect(await hasHorizontalOverflow(stepper)).toBe(false);

  await expect(stepper.getByText('Source', { exact: true })).toBeVisible();
  await expect(stepper.getByText('Connect', { exact: true })).toBeVisible();
  await expect(stepper.getByText('Preview', { exact: true })).toBeVisible();
  await expect(stepper.getByText('Run', { exact: true })).toBeVisible();
  await expect(stepper.getByText('Current', { exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: /Source, current step/ })).toBeVisible();
  await expect(page.getByRole('button', { name: /Go to Preview/ })).toHaveCount(0);

  const labelBoxes = await Promise.all(
    ['Source', 'Connect', 'Preview', 'Run'].map(async (label) => {
      const box = await stepper.getByText(label, { exact: true }).boundingBox();
      expect(box).toBeTruthy();
      return box!;
    })
  );
  for (let index = 1; index < labelBoxes.length; index += 1) {
    expect(labelBoxes[index].x).toBeGreaterThan(labelBoxes[index - 1].x + labelBoxes[index - 1].width);
  }
}
