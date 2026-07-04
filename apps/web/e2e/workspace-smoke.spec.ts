import { expect, test, type Locator } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState, signedUploadPuts, thumbnailRequestPaths } from './workspace-fixture';

test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

test('desktop shell loads the authenticated tenant and compact inventory switcher', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Desktop shell coverage runs on the desktop project.');

  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Home' })).toBeVisible();
  await expect(page.getByRole('button', { name: /Household/ })).toContainText('Home');
  await expect(page.getByRole('navigation', { name: 'Inventory destinations' }).getByText('Search')).toHaveCount(0);
  await expect(page.getByRole('link', { name: /Open location Garage/ })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add', exact: true })).toBeEnabled();

  await page.getByRole('button', { name: /Household/ }).click();
  await expect(page.getByRole('dialog', { name: 'Inventory context' })).toBeVisible();
  await expect(page.getByLabel('Inventories').getByRole('link', { name: /Household/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household'
  );
  await page.getByRole('button', { name: 'Switch tenant' }).click();
  await expect(page.getByLabel('Tenants').getByRole('button', { name: /Cabin/ })).toBeVisible();
  await page.getByLabel('Tenants').getByRole('button', { name: /Cabin/ }).click();
  await expect(page).toHaveURL('/tenants/tenant-cabin/inventories/inventory-cabin');
  await expect(page.getByRole('button', { name: /Cabin Gear/ })).toContainText('Cabin');
});

test('mobile shell opens context and add flows without desktop-only controls', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile shell coverage runs on the mobile project.');

  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Home' })).toBeVisible();
  await expect(page.getByLabel('Search this inventory')).toBeHidden();
  await page.getByRole('button', { name: /Household/ }).click();
  const contextDialog = page.getByRole('dialog', { name: 'Inventory context' });
  await expect(contextDialog).toBeVisible();
  const contextDialogLayer = await dialogTopLayerInfo(contextDialog);
  expect(contextDialogLayer.ownsPoint, JSON.stringify(contextDialogLayer)).toBe(true);
  await expect(page.getByLabel('Inventories').getByRole('link', { name: /Household/ })).toBeVisible();
  await page.keyboard.press('Escape');

  await page.getByRole('link', { name: 'Add asset' }).click();
  const addDialog = page.getByRole('dialog', { name: 'Add item' });
  await expect(addDialog).toBeVisible();
  await expect(page.getByLabel('Item name')).toBeVisible();
  await expect(page.getByLabel('Find parent')).toBeVisible();
  await page.setViewportSize({ width: 393, height: 520 });
  await addDialog.getByRole('switch', { name: 'Create a parent first' }).click();
  await expect(page.getByLabel('Parent name')).toBeVisible();
  await expect(addDialog).toHaveCSS('display', 'flex');
  await expect(addDialog.locator('.tray-actions')).toHaveCSS('position', 'sticky');

  const dialogBox = await addDialog.boundingBox();
  const actionsBox = await addDialog.locator('.tray-actions').boundingBox();
  const viewport = page.viewportSize();
  expect(dialogBox?.y).toBeLessThanOrEqual(10);
  expect(dialogBox && viewport ? dialogBox.y + dialogBox.height : 0).toBeGreaterThanOrEqual((viewport?.height ?? 0) - 2);
  expect(dialogBox && viewport ? dialogBox.y + dialogBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual((viewport?.height ?? 0) + 2);
  expect(actionsBox && viewport ? actionsBox.y + actionsBox.height : 0).toBeGreaterThanOrEqual((viewport?.height ?? 0) - 2);
  expect(actionsBox && viewport ? actionsBox.y + actionsBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual((viewport?.height ?? 0) + 2);
});

test('add flow saves items with and without selected photo previews', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Add smoke uses the desktop create menu.');

  await page.goto('/');

  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await page.locator('#header-add-menu').getByRole('link', { name: 'Item', exact: true }).click();
  await page.getByLabel('Item name').fill('Cordless drill');
  await page.getByRole('button', { name: 'Save item' }).click();
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-cordless-drill');
  await expect(page.getByRole('heading', { name: 'Cordless drill' })).toBeVisible();

  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await page.locator('#header-add-menu').getByRole('link', { name: 'Item', exact: true }).click();
  await page.getByLabel('Item name').fill('Photo tape');
  await page.locator('#asset-photos').setInputFiles({
    name: 'front.png',
    mimeType: 'image/png',
    buffer: tinyPng()
  });
  await expect(page.locator('.photo-preview img[alt="front.png"]')).toBeVisible();
  await expect(page.getByLabel('Photo actions').getByText('1 photo')).toBeVisible();
  await page.getByRole('button', { name: 'Save item' }).click();
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-photo-tape');
  await expect(page.getByRole('heading', { name: 'Photo tape' })).toBeVisible();
  await expect(page.locator('.asset-photo-panel img[alt="Photo tape"]')).toBeVisible();
  expect(signedUploadPuts(page)).toBe(1);

  await page.goto('/tenants/tenant-home/inventories/inventory-household/search');
  await page.getByLabel('Search query').fill('Photo tape');
  const photoTapeThumbnailRequestsBeforeSearch = thumbnailRequestPaths(page).filter((path) =>
    path.includes('/assets/asset-photo-tape/attachments/attachment-photo/thumbnail')
  ).length;
  await page.getByRole('button', { name: 'Run search' }).click();
  await expect(page.locator('.asset-list').getByRole('link', { name: /Photo tape/ })).toBeVisible();
  const photoTapeThumbnail = page.locator('.asset-list img[alt="Photo tape"]');
  await expect(photoTapeThumbnail).toBeVisible();
  expect(
    thumbnailRequestPaths(page).filter((path) => path.includes('/assets/asset-photo-tape/attachments/attachment-photo/thumbnail')).length
  ).toBeGreaterThan(photoTapeThumbnailRequestsBeforeSearch);
});

test('viewer inventory disables desktop add affordances', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Viewer denied smoke runs on desktop.');

  await page.goto('/');

  await page.getByRole('button', { name: /Household/ }).click();
  await page.getByRole('button', { name: 'Switch tenant' }).click();
  await page.getByLabel('Tenants').getByRole('button', { name: /Cabin/ }).click();

  await expect(page).toHaveURL('/tenants/tenant-cabin/inventories/inventory-cabin');
  await expect(page.getByRole('heading', { name: 'Home' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add', exact: true })).toBeDisabled();
  await expect(page.getByRole('link', { name: 'Add location' })).toBeDisabled();
});

test('search entry shows autocomplete and image-bearing results', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Search smoke uses the desktop header search.');

  await page.goto('/');

  await page.getByLabel('Search this inventory').fill('Tomato');
  await expect(page.getByLabel('Search suggestions').getByRole('link', { name: 'Open Tomato fertilizer' })).toBeVisible();
  await expect(page.getByLabel('Search suggestions').locator('img[alt="Tomato fertilizer"]')).toBeVisible();
  await page.getByRole('button', { name: 'Run search' }).click();

  await expect(page.getByRole('heading', { name: 'Search' })).toBeVisible();
  await expect(page.locator('.asset-list').getByRole('link', { name: /Tomato fertilizer/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato'
  );
  await expect(page.locator('.asset-list img[alt="Tomato fertilizer"]')).toBeVisible();
});

test('location navigation opens asset detail and returns to the location list', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Location detail smoke runs on desktop first.');

  await page.goto('/');

  await page.getByRole('link', { name: 'Open location Garage' }).click();
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
  await expect(page.getByRole('heading', { name: 'Garage' })).toBeVisible();

  await page.getByRole('link', { name: /Tomato fertilizer/ }).click();
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato');
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();

  await page.getByRole('link', { name: /Back/ }).click();
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
  await expect(page.getByRole('heading', { name: 'Garage' })).toBeVisible();
});

function tinyPng(): Buffer {
  return Buffer.from(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=',
    'base64'
  );
}

async function dialogTopLayerInfo(dialog: Locator): Promise<{ ownsPoint: boolean; rect: string; topElement: string }> {
  return dialog.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const target = document.elementFromPoint(rect.left + rect.width / 2, rect.top + Math.min(32, rect.height / 2));
    return {
      ownsPoint: target instanceof Node && element.contains(target),
      rect: `${Math.round(rect.left)},${Math.round(rect.top)},${Math.round(rect.width)},${Math.round(rect.height)}`,
      topElement: target instanceof HTMLElement ? `${target.tagName}.${target.className}` : String(target)
    };
  });
}
