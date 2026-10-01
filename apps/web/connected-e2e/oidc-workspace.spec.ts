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

test('real OIDC workspace creation preserves principal isolation', async ({ page, browser, request }, testInfo) => {
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

  const otherContext = await browser.newContext({ baseURL: 'http://localhost:5173' });
  try {
    const otherPage = await otherContext.newPage();
    await signIn(otherPage, 'viewer@example.com');
    await expect(otherPage.getByRole('heading', { name: 'Home', exact: true })).toBeVisible();
    const otherToken = await sessionToken(otherPage);
    expect(Boolean(otherToken)).toBe(true);
    expect([403, 404]).toContain(await inventoryStatus(request, inventoryURL, otherToken));
  } finally { await otherContext.close(); }
});
