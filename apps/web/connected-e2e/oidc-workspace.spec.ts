import { verifyPrintingJourney, verifyPrintingIsolation } from './printing-journey';
import { verifyEditAccessibility } from './edit-accessibility';
import { addArchivePhoto } from './archive-media';
import { verifyArchiveJourney } from './archive-journey';
import { readFile } from 'node:fs/promises';
import { expect, test, type Page, type APIRequestContext } from '@playwright/test';

async function signIn(page: Page, email: string) {
  await page.goto('/');
  await page.getByRole('button', { name: 'Continue to sign in' }).click();
  await expect(page).toHaveURL(/http:\/\/dex:5556\/dex\//);
  // Dex's password connector may first offer the local account method.
  const localLogin = page.getByRole('link', { name: /email/i });
  await expect(page.locator('input[name="login"]').or(localLogin).first()).toBeVisible();
  if (await localLogin.isVisible()) await localLogin.click();
  await page.locator('input[name="login"]').fill(email);
  await page.locator('input[name="password"]').fill('password');
  await page.getByRole('button', { name: /sign in|log in/i }).click();
  await expect(page).toHaveURL(/^http:\/\/localhost:5173\/(?!callback)/);
}

async function sessionToken(page: Page): Promise<string> {
  return page.evaluate(() => JSON.parse(sessionStorage.getItem('stuffstash.oidc.session') ?? '{}').idToken ?? '');
}

async function inventoryStatus(request: APIRequestContext, url: string, token?: string): Promise<number> {
  try {
    const response = await request.get(url, token ? { headers: { Authorization: `Bearer ${token}` } } : {});
    return response.status();
  } catch {
    // Playwright request errors may carry Authorization headers in their call log.
    throw new Error('Connected inventory request failed before returning an HTTP status.');
  }
}

async function exportedAuditFormats(request: APIRequestContext, url: string, token: string): Promise<string[]> {
  try {
    const formats: string[] = [];
    let cursor: string | undefined;
    const visited = new Set<string>();
    do {
      const query = new URLSearchParams({ limit: '100', ...(cursor ? { cursor } : {}) });
      const response = await request.get(`${url}/audit-records?${query}`, { headers: { Authorization: `Bearer ${token}` } });
      if (response.status() !== 200) throw new Error('Audit history unavailable.');
      const body = await response.json() as { data: { action: string; metadata: { format?: string } }[]; meta: { pagination?: { nextCursor?: string } } };
      formats.push(...body.data.filter(record => record.action === 'inventory.exported').map(record => record.metadata.format ?? ''));
      if (formats.includes('json') && formats.includes('csv')) return formats;
      cursor = body.meta.pagination?.nextCursor;
      if (cursor && visited.has(cursor)) throw new Error('Audit cursor repeated.');
      if (cursor) visited.add(cursor);
    } while (cursor && visited.size < 20);
    return formats;
  } catch {
    throw new Error('Could not verify persisted inventory export history.');
  }
}

test('real OIDC workspace, item creation and exports preserve principal isolation', async ({ page, browser, request }, testInfo) => {
  test.setTimeout(240_000);
  await signIn(page, 'owner@example.com');
  await expect(page.getByRole('heading', { name: 'Home', exact: true })).toBeVisible();
  await page.getByRole('link', { name: /^Browse\b/ }).click();
  await expect(page).toHaveURL(/\/tenants\/[^/]+\/inventories\/[^/?#]+/);
  const scope = new URL(page.url()).pathname.match(/\/tenants\/([^/]+)\/inventories\/([^/]+)/)!;
  const inventoryURL = `http://localhost:8080/tenants/${scope[1]}/inventories/${scope[2]}`;
  const ownerToken = await sessionToken(page);
  expect(Boolean(ownerToken)).toBe(true);
  expect(await inventoryStatus(request, inventoryURL, ownerToken)).toBe(200);
  expect(await inventoryStatus(request, inventoryURL)).toBe(401);
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Browse', exact: true })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath('connected-browse.png') });

  const labelResolutionURL = await verifyPrintingJourney(page, request, inventoryURL, ownerToken, testInfo);

  const originalTitle = 'Connected export, flashlight';
  const itemTitle = 'Connected edited, flashlight';
  await page.goto(`/tenants/${scope[1]}/inventories/${scope[2]}/add/item`);
  await expect(page.getByRole('dialog', { name: 'Add item' })).toBeVisible();
  await page.getByLabel('Item name').fill(originalTitle);
  await page.getByRole('button', { name: 'Save item' }).click();
  await expect(page.getByRole('heading', { name: originalTitle, exact: true })).toBeVisible();
  const assetPath = new URL(page.url()).pathname;
  expect(assetPath).toMatch(/\/assets\/[^/]+$/);
  const assetURL = `http://localhost:8080${assetPath}`;
  await page.goto(`${assetPath}/edit`);
  const edit = page.getByRole('dialog', { name: 'Edit asset' });
  await expect(edit).toBeVisible();
  await edit.getByLabel('Name', { exact: true }).fill(itemTitle);
  await edit.getByRole('button', { name: 'Save', exact: true }).click();
  await expect(page.getByRole('heading', { name: itemTitle, exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole('heading', { name: itemTitle, exact: true })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath('connected-edited-asset.png') });
  expect(await inventoryStatus(request, assetURL)).toBe(401);
  await verifyEditAccessibility(page, assetPath, itemTitle, testInfo);
  await page.goto(`/settings/tenants/${scope[1]}/inventories/${scope[2]}`);
  for (const format of ['json', 'csv'] as const) {
    await page.getByRole('button', { name: 'Export inventory', exact: true }).click();
    const downloaded = page.waitForEvent('download', { timeout: 20_000 }).catch(() => null);
    const exported = page.waitForResponse(response => response.url() === `${inventoryURL}/export?format=${format}`, { timeout: 20_000 });
    await page.getByRole('menuitem', { name: format === 'json' ? /JSON —/ : /CSV —/ }).click();
    const response = await exported;
    await page.screenshot({ path: testInfo.outputPath(`connected-export-${format}-response.png`) });
    expect(response.status(), `Authenticated ${format} export HTTP status`).toBe(200);
    const file = await downloaded;
    expect(file, `Browser must accept the ${format} export download`).not.toBeNull();
    if (!file) throw new Error('Browser did not accept the inventory download.');
    expect(file.suggestedFilename()).toBe(`stuff-stash-inventory.${format}`);
    expect(await file.failure()).toBeNull();
    const filePath = await file.path();
    expect(filePath).not.toBeNull();
    const contents = await readFile(filePath!, 'utf8');
    if (format === 'json') {
      const document = JSON.parse(contents);
      expect(document.schemaVersion).toBe(1);
      expect(document.assets).toEqual(expect.arrayContaining([expect.objectContaining({ title: itemTitle })]));
    } else {
      expect(contents.split(/\r?\n/, 1)[0]).toBe('id,title,description,kind,parentAssetId,customAssetTypeId,lifecycleState,createdAt,updatedAt,expirationDate,expirationPrecision,tagIds,customFields,currentCheckout,attachments');
      expect(contents).toContain(`,"${itemTitle}",`);
    }
    await expect(page.getByRole('status').filter({ hasText: 'Inventory download started.' })).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath(`connected-export-${format}.png`) });
    await file.delete();
    expect(await inventoryStatus(request, `${inventoryURL}/export?format=${format}`)).toBe(401);
  }

  expect(await exportedAuditFormats(request, inventoryURL, ownerToken)).toEqual(expect.arrayContaining(['json', 'csv']));

  const photo = await addArchivePhoto(request, assetURL, ownerToken, await page.screenshot());
  await verifyArchiveJourney(page, scope[1], scope[2], itemTitle, testInfo, request, ownerToken, photo);

  const otherContext = await browser.newContext({ baseURL: 'http://localhost:5173' });
  try {
    const otherPage = await otherContext.newPage();
    await signIn(otherPage, 'viewer@example.com');
    await expect(otherPage.getByRole('heading', { name: 'Home', exact: true })).toBeVisible();
    const otherToken = await sessionToken(otherPage);
    expect(Boolean(otherToken)).toBe(true);
    await verifyPrintingIsolation(request, labelResolutionURL, otherToken);
    expect([403, 404]).toContain(await inventoryStatus(request, inventoryURL, otherToken));
    expect([403, 404]).toContain(await inventoryStatus(request, assetURL, otherToken));
    for (const format of ['json', 'csv']) {
      expect([403, 404]).toContain(await inventoryStatus(request, `${inventoryURL}/export?format=${format}`, otherToken));
    }
  } finally { await otherContext.close(); }
});
