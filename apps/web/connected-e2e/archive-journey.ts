import { expect, type Page, type TestInfo } from '@playwright/test';
import { readFile, writeFile } from 'node:fs/promises';

/** Real API/job/ZIP flow using the portable browser download fallback. */
async function runArchiveJourney(page: Page, tenantId: string, inventoryId: string, title: string, info: TestInfo) {
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
  await writeFile(info.outputPath('connected-archive-fixture.zip'), bytes);
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

export async function verifyArchiveJourney(page: Page, tenantId: string, inventoryId: string, title: string, info: TestInfo) {
  const states: { status: number; state?: string; failure?: string; error?: string }[] = [];
  const pending: Promise<void>[] = [];
  const observe = (response: import('@playwright/test').Response) => {
    if (!response.url().includes('/archive-') || response.url().includes('/content')) return;
    pending.push((async () => {
      try {
        const body = await response.json();
        const jobs = Array.isArray(body.data) ? body.data : [body.data];
        for (const job of jobs) states.push({ status: response.status(), state: job?.state, failure: job?.failure, error: body.error?.code });
      } catch { states.push({ status: response.status() }); }
    })());
  };
  page.on('response', observe);
  try { await runArchiveJourney(page, tenantId, inventoryId, title, info); }
  catch (error) { await page.screenshot({ path: info.outputPath('connected-archive-failure.png'), fullPage: true }); throw error; }
  finally {
    page.off('response', observe); await Promise.all(pending);
    await writeFile(info.outputPath('connected-archive-state.json'), JSON.stringify(states));
  }
}
