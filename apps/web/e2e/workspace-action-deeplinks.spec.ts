import { expect, test, type Locator } from '@playwright/test';
import { installAuthenticatedWorkspace, lastAssetPatch, resetWorkspaceApiState } from './workspace-fixture';

test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

test('asset edit action can be opened, saved, and closed from a direct URL', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Action deep-link coverage runs on desktop.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato/edit');

  const editPanel = page.getByRole('dialog', { name: 'Edit asset' });
  const editHeading = editPanel.getByRole('heading', { name: 'Edit asset' });
  await expect(editPanel).toBeVisible();
  await expect(async () => {
    expect(await elementIsInViewportAndUnoccluded(editHeading)).toBe(true);
  }).toPass();
  await expect(editPanel.getByLabel('Name')).toHaveValue('Tomato fertilizer');

  await editPanel.getByLabel('Name').fill('Tomato fertilizer granules');
  await page.getByRole('button', { name: 'Save' }).click();

  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato');
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer granules' })).toBeVisible();
  expect(lastAssetPatch(page)).toMatchObject({ assetId: 'asset-tomato', title: 'Tomato fertilizer granules' });

  await page.reload();
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer granules' })).toBeVisible();
});

test('mobile asset edit direct URL lands on the edit panel', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile direct edit coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato/edit');

  const editPanel = page.getByRole('dialog', { name: 'Edit asset' });
  const editHeading = editPanel.getByRole('heading', { name: 'Edit asset' });
  await expect(editPanel).toBeVisible();
  await expect(async () => {
    expect(await elementIsInViewportAndUnoccluded(editHeading)).toBe(true);
  }).toPass();
  await expect(editPanel.getByLabel('Name')).toHaveValue('Tomato fertilizer');
});

test('asset move action direct URL opens the shared searchable parent picker', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Action deep-link coverage runs on desktop.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato/move');

  await expect(page.getByRole('heading', { name: 'Move asset' })).toBeVisible();
  await expect(page.getByRole('group', { name: 'Move target current destination' })).toBeVisible();
  await expect(page.getByRole('group', { name: 'Move target suggested destinations' })).toBeVisible();
  await expect(page.getByText('Suggested destinations', { exact: true })).toBeVisible();

  await page.getByLabel('Find parent').fill('Garage');
  await expect(page.getByRole('group', { name: 'Move target search results' })).toBeVisible();
  await page.getByRole('group', { name: 'Move target root destination' }).getByRole('button', { name: 'Inventory root' }).click();
  await page.getByRole('button', { name: 'Move' }).click();

  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato');
  expect(lastAssetPatch(page)).toMatchObject({ assetId: 'asset-tomato', parentAssetId: null });

  await page.reload();
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();
  await expect(page.getByText('inventory root')).toBeVisible();
});

test('unavailable action deep links normalize back to asset detail', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Action deep-link coverage runs on desktop.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato/restore');

  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato');
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Restore asset' })).toHaveCount(0);
});

async function elementIsInViewportAndUnoccluded(locator: Locator): Promise<boolean> {
  return locator.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const target = document.elementFromPoint(rect.left + rect.width / 2, rect.top + rect.height / 2);
    return rect.top >= 0 && rect.bottom <= window.innerHeight && target instanceof Node && element.contains(target);
  });
}
