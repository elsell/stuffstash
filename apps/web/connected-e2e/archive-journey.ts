import { expect, type Page, type TestInfo } from '@playwright/test';
import { readFile } from 'node:fs/promises';

/** Real API/job/ZIP flow using the portable browser download fallback. */
export async function verifyArchiveJourney(page: Page, tenantId: string, inventoryId: string, title: string, info: TestInfo) {
  // Playwright cannot operate the OS file-save picker. Exercise the supported
  // bounded download fallback, without intercepting any archive HTTP operation.
  await page.addInitScript(() => { Object.defineProperty(window, 'showSaveFilePicker', { value: undefined, configurable: true }); });
  await page.goto(`/settings/tenants/${tenantId}/inventories/${inventoryId}`);
  const exportTask = page.getByRole('region', { name: 'Export archive', exact: true });
  await exportTask.getByRole('button', { name: 'Create archive', exact: true }).click();
  const downloadButton = exportTask.getByRole('button', { name: 'Download', exact: true }).first();
  await expect(downloadButton).toBeVisible({ timeout: 45_000 });
  const downloaded = page.waitForEvent('download');
  await downloadButton.click();
  const archive = await downloaded;
  const path = await archive.path();
  expect(path).not.toBeNull();
  const bytes = await readFile(path!);
  expect(bytes.subarray(0, 4).equals(Buffer.from([0x50, 0x4b, 0x03, 0x04]))).toBe(true);
  await page.screenshot({ path: info.outputPath('connected-archive-export.png'), fullPage: true });
  try {
    await page.goto(`/settings/tenants/${tenantId}`);
    const restoreTask = page.getByRole('region', { name: 'Restore inventory', exact: true });
    await restoreTask.getByLabel('Choose archive', { exact: true }).setInputFiles(path!);
    await restoreTask.getByRole('button', { name: 'Upload and validate', exact: true }).click();
    await expect(restoreTask.getByRole('button', { name: 'Review restore', exact: true })).toBeVisible({ timeout: 45_000 });
    // Validation alone must not expose a destination inventory to open.
    await expect(restoreTask.getByRole('link', { name: 'Open inventory', exact: true })).toHaveCount(0);
    await restoreTask.getByRole('button', { name: 'Review restore', exact: true }).click();
    await restoreTask.getByLabel('Inventory name', { exact: true }).fill('Connected restored inventory');
    await expect(restoreTask).toContainText('This creates a new inventory');
    await page.screenshot({ path: info.outputPath('connected-archive-review.png'), fullPage: true });
    await restoreTask.getByRole('button', { name: 'Restore inventory', exact: true }).click();
    const open = restoreTask.getByRole('link', { name: 'Open inventory', exact: true });
    await expect(open).toBeVisible({ timeout: 45_000 });
    const destination = await open.getAttribute('href');
    expect(destination).not.toContain(`/inventories/${inventoryId}`);
    expect(destination).not.toBeNull();
    const destinationURL = new URL(destination!, page.url()).href;
    await open.click();
    await expect(page).toHaveURL(destinationURL);
    await expect(page.getByText(title, { exact: true }).first()).toBeVisible();
    await page.reload();
    await expect(page).toHaveURL(destinationURL);
    await expect(page.getByText(title, { exact: true }).first()).toBeVisible();
    await page.screenshot({ path: info.outputPath('connected-archive-restored.png'), fullPage: true });
  } finally { await archive.delete(); }
}
