import { expect, test, type Locator } from '@playwright/test';
import {
  apiRequestPaths,
  installAuthenticatedWorkspace,
  resetWorkspaceApiState,
  signedUploadPuts,
  thumbnailRequestPaths
} from './workspace-fixture';

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
  expect(await clippedTextCount(page.locator('.nav-button small'))).toBe(0);

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
  await page.getByLabel('Find parent').fill('g');
  const lastParentResult = addDialog.locator('.parent-picker-results .parent-target-button').last();
  await expect(lastParentResult).toBeVisible();
  await lastParentResult.scrollIntoViewIfNeeded();
  const selectedParentName = await lastParentResult.locator('strong').textContent();
  const lastParentResultBox = await lastParentResult.boundingBox();
  const actionsBoxAfterParentScroll = await addDialog.locator('.tray-actions').boundingBox();
  const addDialogBoxAfterParentScroll = await addDialog.boundingBox();
  expect(
    lastParentResultBox && actionsBoxAfterParentScroll
      ? lastParentResultBox.y + lastParentResultBox.height
      : Number.POSITIVE_INFINITY
  ).toBeLessThanOrEqual((actionsBoxAfterParentScroll?.y ?? 0) - 8);
  expect(
    addDialogBoxAfterParentScroll && lastParentResultBox
      ? addDialogBoxAfterParentScroll.y + addDialogBoxAfterParentScroll.height - (lastParentResultBox.y + lastParentResultBox.height)
      : 0
  ).toBeGreaterThanOrEqual(88);
  await lastParentResult.click();
  await expect(addDialog.locator('.add-summary-destination strong')).toHaveText(selectedParentName ?? '');
  await expect(addDialog.locator('.tray-actions')).toHaveCSS('position', 'static');
  await page.getByLabel('Find parent').fill('');
  await addDialog.getByRole('switch', { name: 'Create a parent first' }).click();
  await expect(page.getByLabel('Parent name')).toBeVisible();
  await expect(addDialog).toHaveCSS('display', 'flex');
  await expect(addDialog.locator('.tray-actions')).toHaveCSS('position', 'static');

  const dialogBox = await addDialog.boundingBox();
  const bodyBox = await addDialog.locator('.add-tray-body').boundingBox();
  const actionsBox = await addDialog.locator('.tray-actions').boundingBox();
  const viewport = page.viewportSize();
  await expectCompactAddSummary(addDialog);
  expect(dialogBox?.y).toBeLessThanOrEqual(10);
  expect(dialogBox && viewport ? dialogBox.y + dialogBox.height : 0).toBeGreaterThanOrEqual((viewport?.height ?? 0) - 2);
  expect(dialogBox && viewport ? dialogBox.y + dialogBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual((viewport?.height ?? 0) + 2);
  expect(bodyBox && actionsBox ? bodyBox.y + bodyBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual((actionsBox?.y ?? 0) + 1);
  expect(actionsBox && viewport ? actionsBox.y + actionsBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual((viewport?.height ?? 0) + 2);
  expect(actionsBox && viewport ? viewport.height - (actionsBox.y + actionsBox.height) : Number.POSITIVE_INFINITY).toBeLessThanOrEqual(32);

  await page.setViewportSize({ width: 320, height: 520 });
  await expectCompactAddSummary(addDialog);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)).toBe(true);
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

test('add location deep link saves to the canonical focused location route', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Add-location smoke uses the desktop create menu.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/add/location');

  const dialog = page.getByRole('dialog', { name: 'Add location' });
  await expect(dialog).toBeVisible();
  await expect(page.getByLabel('Location name')).toHaveAttribute('placeholder', 'Garage shelf');
  await page.getByLabel('Location name').fill('Garage shelf');
  await page.getByRole('button', { name: 'Save location' }).click();

  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/locations/asset-garage-shelf');
  await expect(page.getByRole('heading', { name: 'Garage shelf' })).toBeVisible();
  await expect(page.getByRole('link', { name: /Back/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/locations'
  );
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

test('settings and import deep links render route-backed sections', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Settings and import deep links run on desktop first.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=pending');
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=pending');
  await expect(page.getByRole('heading', { name: 'Sharing' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Pending', exact: true })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=pending'
  );
  await expect(page.getByRole('link', { name: 'Pending', exact: true })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByLabel('Invitations').getByText('friend@example.test')).toBeVisible();
  await expect(page.getByLabel('Direct grants').getByText('oidc_OuQU94grMoaZ8cly6ZUUpXUVhloLanDNZ')).toBeVisible();
  expect(await hasHorizontalOverflow(page.locator('.settings-panel').filter({ hasText: 'Sharing' }).first())).toBe(false);
  expect(apiRequestPaths(page)).toContain(
    'GET /tenants/tenant-home/inventories/inventory-household/access-invitations?limit=50&status=pending'
  );

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/fields');
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/settings/fields');
  await expect(page.getByRole('heading', { name: 'Custom fields' })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: /Fields/ })).toHaveAttribute(
    'aria-current',
    'page'
  );

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/activity?auditScope=tenant');
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/settings/activity?auditScope=tenant');
  await expect(page.getByRole('heading', { name: 'Activity' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Tenant', exact: true })).toHaveAttribute('aria-current', 'page');
  await expect(page.locator('.audit-row').first().getByText('Asset created')).toBeVisible();
  await expect(page.locator('.audit-row').first().locator('[data-slot="badge"]')).toHaveText('API');
  expect(apiRequestPaths(page)).toContain('GET /tenants/tenant-home/audit-records?limit=50');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox-csv');
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox-csv');
  await expect(page.getByRole('heading', { name: 'Import', exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: 'CSV', exact: true })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByLabel('CSV file')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Connect', exact: true })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox'
  );
});

test('mobile import source actions clear the bottom nav', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile import layout coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox');

  await expect(page.getByRole('heading', { name: 'Import', exact: true })).toBeVisible();
  const sourcePanel = page.locator('.import-source-panel');
  await sourcePanel.getByRole('button', { name: 'Preview' }).scrollIntoViewIfNeeded();
  await page.evaluate(() => { window.scrollBy(0, 96); });
  const actionsBox = await sourcePanel.locator('.heading-actions').boundingBox();
  const navBox = await page.locator('.mobile-nav').boundingBox();
  const lastOptionBox = await sourcePanel.locator('.binary-option').filter({ hasText: 'Private network address' }).boundingBox();
  expect(lastOptionBox && actionsBox ? lastOptionBox.y + lastOptionBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual(
    (actionsBox?.y ?? 0) - 8
  );
  expect(actionsBox && navBox ? actionsBox.y + actionsBox.height : Number.POSITIVE_INFINITY).toBeLessThanOrEqual((navBox?.y ?? 0) - 8);

  await page.setViewportSize({ width: 820, height: 650 });
  await page.goto('/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox');
  await expect(page.locator('.mobile-nav')).toBeVisible();
  await expect(page.locator('.import-layout')).toHaveCSS('grid-template-columns', '792px');
});

test('mobile long settings and import pages keep final content above bottom chrome', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile bottom clearance coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=pending');
  await expect(page.getByRole('heading', { name: 'Sharing' })).toBeVisible();
  await scrollToEnd(page.locator('.settings-panel').last());
  expect(await clearsBottomChrome(page.locator('.settings-panel').last(), page.locator('.mobile-nav'))).toBe(true);

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/activity?auditScope=tenant');
  await expect(page.getByRole('heading', { name: 'Activity' })).toBeVisible();
  await scrollToEnd(page.locator('.settings-panel').last());
  expect(await clearsBottomChrome(page.locator('.settings-panel').last(), page.locator('.mobile-nav'))).toBe(true);

  await page.goto('/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox');
  await expect(page.getByRole('heading', { name: 'Import', exact: true })).toBeVisible();
  await scrollToEnd(page.locator('.import-source-panel .heading-actions'));
  expect(await clearsBottomChrome(page.locator('.import-source-panel .heading-actions'), page.locator('.mobile-nav'))).toBe(true);
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

async function hasHorizontalOverflow(locator: Locator): Promise<boolean> {
  return locator.evaluate((element) => element.scrollWidth > element.clientWidth + 1);
}

async function clearsBottomChrome(content: Locator, chrome: Locator): Promise<boolean> {
  const [contentBox, chromeBox] = await Promise.all([content.boundingBox(), chrome.boundingBox()]);
  if (!contentBox || !chromeBox) {
    return false;
  }
  return contentBox.y + contentBox.height <= chromeBox.y - 8;
}

async function scrollToEnd(locator: Locator): Promise<void> {
  await locator.evaluate((element) => element.scrollIntoView({ block: 'end', inline: 'nearest' }));
}

async function clippedTextCount(locator: Locator): Promise<number> {
  return locator.evaluateAll((elements) =>
    elements.filter((element) => element.scrollWidth > element.clientWidth + 1 || element.scrollHeight > element.clientHeight + 1).length
  );
}

async function expectCompactAddSummary(addDialog: Locator): Promise<void> {
  const summary = addDialog.locator('.add-summary');
  const summaryMetrics = await summary.evaluate((element) => {
    const box = element.getBoundingClientRect();
    const childBoxes = Array.from(element.querySelectorAll(':scope > div')).map((child) => child.getBoundingClientRect());
    const firstY = childBoxes[0]?.y ?? 0;
    return {
      height: box.height,
      width: box.width,
      scrollWidth: element.scrollWidth,
      sameRow: childBoxes.every((childBox) => Math.abs(childBox.y - firstY) <= 1),
      destinationWidth: childBoxes[1]?.width ?? 0,
      typeWidth: childBoxes[0]?.width ?? 0,
      photoWidth: childBoxes[2]?.width ?? 0
    };
  });
  expect(summaryMetrics.height).toBeLessThanOrEqual(62);
  expect(summaryMetrics.scrollWidth).toBeLessThanOrEqual(summaryMetrics.width + 1);
  expect(summaryMetrics.sameRow).toBe(true);
  expect(summaryMetrics.destinationWidth).toBeGreaterThan(summaryMetrics.typeWidth);
  expect(summaryMetrics.destinationWidth).toBeGreaterThan(summaryMetrics.photoWidth);
  const hiddenSummary = summary.locator('.visually-hidden');
  await expect(hiddenSummary).toHaveAttribute('aria-live', 'polite');
  await expect(hiddenSummary).toHaveAttribute('aria-atomic', 'true');
  const hiddenSummaryText = await hiddenSummary.evaluate((element) =>
    (element.textContent ?? '').replace(/\s+/g, ' ').trim()
  );
  expect(hiddenSummaryText).toMatch(/^Type: .+ Parent: .+ Photos: .+$/);
  await expect(summary.locator('.add-summary-destination strong')).not.toHaveText('');
}
