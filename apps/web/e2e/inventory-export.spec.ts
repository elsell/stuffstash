import { expect, test } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

const settingsPath = '/settings/tenants/tenant-home/inventories/inventory-household';
const exportURL = 'http://127.0.0.1:18080/tenants/tenant-home/inventories/inventory-household/export*';
test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

for (const format of ['json', 'csv'] as const) {
  test(`downloads the ${format} inventory file from scoped settings`, async ({ page }, testInfo) => {
    const body = format === 'json' ? '{"schemaVersion":1,"assets":[{"title":"Camping tent"}]}' : 'id,title\n1,Camping tent\n';
    await page.route(exportURL, async route => {
      expect(route.request().headers().authorization).toMatch(/^Bearer /);
      expect(new URL(route.request().url()).searchParams.get('format')).toBe(format);
      await route.fulfill({ status: 200, contentType: format === 'json' ? 'application/json' : 'text/csv', body });
    });
    await page.goto(settingsPath);
    await page.getByRole('button', { name: 'Export inventory', exact: true }).click();
    const downloaded = page.waitForEvent('download');
    await page.getByRole('menuitem', { name: format === 'json' ? /JSON —/ : /CSV —/ }).click();
    const file = await downloaded;
    expect(file.suggestedFilename()).toBe(`stuff-stash-inventory.${format}`);
    const path = await file.path(); expect(path).not.toBeNull();
    expect(await readFile(path!, 'utf8')).toBe(body);
    await expect(page.getByRole('status').filter({ hasText: 'Inventory download started.' })).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath(`export-${format}.png`), fullPage: true });
  });
}

test('keeps a failed export retryable without downloading an error envelope', async ({ page }) => {
  let requests = 0;
  const downloads: string[] = [];
  page.on('download', file => downloads.push(file.suggestedFilename()));
  await page.route(exportURL, async route => {
    requests++;
    await route.fulfill({ status: 403, contentType: 'application/json', body: JSON.stringify({ error: { code: 'forbidden', message: 'Forbidden.' }, meta: {} }) });
  });
  await page.goto(settingsPath);
  await page.getByRole('button', { name: 'Export inventory', exact: true }).click();
  await page.getByRole('menuitem', { name: /JSON —/ }).click();
  await expect(page.getByRole('alert')).toContainText('You no longer have access');
  expect(downloads).toEqual([]);
  await page.getByRole('button', { name: 'Retry export' }).click();
  await expect.poll(() => requests).toBe(2);
  expect(downloads).toEqual([]);
});
