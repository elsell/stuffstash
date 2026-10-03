import { generateKeyPairSync } from 'node:crypto';
import { expect, type APIRequestContext, type Page, type TestInfo } from '@playwright/test';

// Transport exceptions can include bearer credentials in Playwright call logs.
async function api(request: APIRequestContext, method: 'get' | 'post', url: string, token?: string, data?: unknown, key?: string) {
  try {
    return await request[method](url, { data, headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(key ? { 'Idempotency-Key': key } : {})
    } });
  } catch { throw new Error('Connected printing request failed before returning a status.'); }
}
async function choose(page: Page, label: string, option: string | RegExp) {
  await page.getByRole('button', { name: label, exact: true }).click();
  await page.getByRole('option', { name: option }).click();
}

export async function verifyPrintingJourney(page: Page, request: APIRequestContext, inventoryURL: string, token: string, testInfo: TestInfo) {
  const scopePath = new URL(inventoryURL).pathname;
  const publicKey = generateKeyPairSync('ed25519').publicKey.export({ type: 'spki', format: 'der' }).subarray(-32).toString('base64');
  const started = await api(request, 'post', 'http://localhost:8080/print-connector-pairings', undefined, {
    name: 'Connected acceptance computer', publicKey,
    candidates: [{ id: 'offline-printer', name: 'Connected Brother', adapterId: 'brother-ql800', deviceId: 'acceptance-offline-device' }]
  });
  expect(started.status()).toBe(201);
  const pairing = (await started.json()).data;
  await page.goto(`/print-connectors/pair/${pairing.id}`);
  await page.getByRole('button', { name: 'Inventory', exact: true }).click();
  await page.getByRole('option').filter({ hasNotText: 'Choose an inventory' }).first().click();
  await page.getByLabel('Code shown in the CLI').fill(pairing.userCode);
  const reviewed = page.waitForResponse(response => response.url().endsWith(`/print-connector-pairings/${pairing.id}/review`));
  await page.getByRole('button', { name: 'Review connection', exact: true }).click();
  expect((await reviewed).status(), 'Real pairing review accepts the browser request').toBe(200);
  await choose(page, 'Printer destination', 'Register a new printer');
  await choose(page, 'Loaded label size', /29.*90/);
  await page.getByRole('button', { name: 'Approve connection', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Connection approved', exact: true })).toBeVisible();
  await page.goto(`/settings${scopePath}/printing`);
  await expect(page.getByRole('heading', { name: 'Printers and labels', exact: true })).toBeVisible();
  await choose(page, 'Default printer', /Connected Brother/);
  await page.getByRole('checkbox', { name: 'Print a label by default when creating an asset', exact: true }).check();
  await page.getByRole('button', { name: 'Save defaults', exact: true }).click();
  await expect(page.getByRole('status').filter({ hasText: 'Print defaults saved' })).toBeVisible();
  await page.goto(`${scopePath}/add/item`);
  await expect(page.getByRole('checkbox', { name: 'Print a label', exact: true })).toBeChecked();
  await page.getByLabel('Item name').fill('Connected printed flashlight');
  const created = page.waitForResponse(response => response.url() === `${inventoryURL}/assets` && response.request().method() === 'POST');
  await page.getByRole('button', { name: 'Save item', exact: true }).click();
  const response = await created;
  expect(response.status()).toBe(201);
  const asset = (await response.json()).data;
  expect(asset.printJobId).toBeTruthy();
  const sent = response.request();
  const key = sent.headers()['idempotency-key'];
  expect(key).toBeTruthy();
  const replay = await api(request, 'post', `${inventoryURL}/assets`, token, sent.postDataJSON(), key);
  expect(replay.status()).toBe(200);
  expect((await replay.json()).data).toMatchObject({ id: asset.id, printJobId: asset.printJobId });
  const jobs = await api(request, 'get', `${inventoryURL}/print-jobs`, token);
  expect(jobs.status()).toBe(200);
  const matching = (await jobs.json()).data.filter((job: { assetId: string }) => job.assetId === asset.id);
  expect(matching).toHaveLength(1);
  expect(matching[0]).toMatchObject({ id: asset.printJobId, status: 'queued' });
  await page.goto(`/settings${scopePath}/printing`);
  await page.reload();
  await expect(page.getByRole('status').filter({ hasText: /^Queued$/ })).toHaveCount(1);
  await expect(page.getByText('Waiting for the printer. Check its power, USB connection, and label roll.', { exact: true })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath('connected-print-queued.png'), fullPage: true });
  const label = await api(request, 'get', `${inventoryURL}/assets/${asset.id}/label`, token);
  expect(label.status()).toBe(200);
  const identity = (await label.json()).data;
  const resolutionURL = `http://localhost:8080/labels/v1/${identity.instanceId}/${identity.labelId}`;
  const resolved = await api(request, 'get', resolutionURL, token);
  expect(resolved.status()).toBe(200);
  expect((await resolved.json()).data.assetId).toBe(asset.id);
  expect((await api(request, 'get', resolutionURL)).status()).toBe(401);
  return resolutionURL;
}
export async function verifyPrintingIsolation(request: APIRequestContext, resolutionURL: string, token: string) {
  expect([403, 404]).toContain((await api(request, 'get', resolutionURL, token)).status());
}
