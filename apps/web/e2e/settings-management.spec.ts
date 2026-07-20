import { expect, test } from '@playwright/test';
import { apiRequestPaths, installAuthenticatedWorkspace, resetWorkspaceApiState, seedSettingsManagementState, setHomeInventoryPermissions } from './workspace-fixture';

const tagsPath = '/settings/tenants/tenant-home/inventories/inventory-household/tags';
const fieldsPath = '/settings/tenants/tenant-home/inventories/inventory-household/fields';
const assetTypesPath = '/settings/tenants/tenant-home/inventories/inventory-household/asset-types';

test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

test('renders aligned color and no-color tag rows at wide and narrow sizes', async ({ page }, testInfo) => {
  await page.goto(tagsPath);
  await expect(page.getByRole('heading', { name: 'Tags', exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: /Workshop Blue color \(#2F80ED\)/ })).toBeVisible();
  await expect(page.getByRole('link', { name: /Reference No color/ })).toBeVisible();

  const indicators = page.locator('.settings-tag-color-indicator');
  await expect(indicators).toHaveCount(2);
  const boxes = await indicators.evaluateAll((nodes) => nodes.map((node) => node.getBoundingClientRect()));
  expect(Math.abs(boxes[0].x - boxes[1].x)).toBeLessThan(1);
  await expect(page.getByRole('link', { name: /Reference No color/ }).locator('.settings-tag-color-indicator')).toHaveClass(/settings-tag-color-empty/);

  await page.screenshot({ path: testInfo.outputPath(`tags-${testInfo.project.name}.png`), fullPage: true });
});

test('tag editor exposes native color selection and clear color', async ({ page }, testInfo) => {
  await page.goto(`${tagsPath}/tag-uncolored/edit`);
  await expect(page.getByRole('dialog', { name: 'Edit Reference' })).toBeVisible();
  await expect(page.locator('input[type="color"]')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Clear color' })).toBeDisabled();
  await page.locator('input[type="color"]').fill('#7c3aed');
  await expect(page.getByRole('button', { name: 'Clear color' })).toBeEnabled();
  await page.screenshot({ path: testInfo.outputPath(`tag-editor-${testInfo.project.name}.png`), fullPage: true });
});

test('protects unsaved field and asset type drafts before dismissal', async ({ page }, testInfo) => {
  await page.goto(`${fieldsPath}/new`);
  await page.getByLabel('Display name').fill('Warranty details');
  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(page.getByRole('alertdialog', { name: 'Discard changes?' })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath(`field-discard-confirmation-${testInfo.project.name}.png`), fullPage: true });
  await page.getByRole('button', { name: 'Keep editing' }).click();
  await expect(page.getByLabel('Display name')).toHaveValue('Warranty details');

  await page.goto(`${assetTypesPath}/new`);
  await page.getByLabel('Display name').fill('Appliance');
  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(page.getByRole('alertdialog', { name: 'Discard changes?' })).toBeVisible();
  await page.getByRole('button', { name: 'Discard changes' }).click();
  await expect(page).toHaveURL(assetTypesPath);
});

test('suggests complete field and asset type keys until the user edits them', async ({ page }) => {
  await page.goto(`${fieldsPath}/new`);
  await page.getByLabel('Display name').fill('Purchase date');
  await expect(page.getByLabel('Stable key')).toHaveValue('purchase-date');
  await page.getByLabel('Display name').fill('Original purchase date');
  await expect(page.getByLabel('Stable key')).toHaveValue('original-purchase-date');
  await page.getByLabel('Stable key').fill('bought-on');
  await page.getByLabel('Display name').fill('Date bought');
  await expect(page.getByLabel('Stable key')).toHaveValue('bought-on');

  await page.goto(`${assetTypesPath}/new`);
  await page.getByLabel('Display name').fill('Power tool');
  await expect(page.getByLabel('Stable key')).toHaveValue('power-tool');
  await page.getByLabel('Display name').fill('Cordless power tool');
  await expect(page.getByLabel('Stable key')).toHaveValue('cordless-power-tool');
  await page.getByLabel('Stable key').fill('portable-tool');
  await page.getByLabel('Display name').fill('Battery-powered tool');
  await expect(page.getByLabel('Stable key')).toHaveValue('portable-tool');
});

test('creates and edits fields, asset types, and tags through their production routes', async ({ page }) => {
  seedSettingsManagementState(page);

  await page.goto(`${fieldsPath}/new`);
  await page.getByLabel('Display name').fill('Warranty expires');
  await page.getByLabel('Display name').blur();
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page).toHaveURL(fieldsPath);
  await expect(page.getByRole('link', { name: /Warranty expires/ })).toBeVisible();
  await page.getByRole('link', { name: /Warranty expires/ }).click();
  await page.getByLabel('Display name').fill('Warranty end date');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByRole('link', { name: /Warranty end date/ })).toBeVisible();

  await page.goto(`${assetTypesPath}/new`);
  await page.getByLabel('Display name').fill('Power tool');
  await page.getByLabel('Display name').blur();
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page).toHaveURL(assetTypesPath);
  await expect(page.getByRole('link', { name: /Power tool/ })).toBeVisible();
  await page.getByRole('link', { name: /Power tool/ }).click();
  await page.getByLabel('Display name').fill('Cordless power tool');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByRole('link', { name: /Cordless power tool/ })).toBeVisible();

  await page.goto(`${tagsPath}/new`);
  await page.getByLabel('Name').fill('Insurance');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page).toHaveURL(tagsPath);
  await expect(page.getByRole('link', { name: /Insurance No color/ })).toBeVisible();
  await page.getByRole('link', { name: /Insurance No color/ }).click();
  await page.getByLabel('Name').fill('Insurance records');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByRole('link', { name: /Insurance records No color/ })).toBeVisible();
});

test('prevents duplicate field creation while a save is in flight', async ({ page }) => {
  let release!: () => void;
  const pending = new Promise<void>((resolve) => { release = resolve; });
  let creates = 0;
  await page.route('http://127.0.0.1:18080/**/custom-field-definitions', async (route) => {
    if (route.request().method() !== 'POST') return route.fallback();
    creates += 1;
    await pending;
    await route.fallback();
  });

  await page.goto(`${fieldsPath}/new`);
  await page.getByLabel('Display name').fill('Duplicate guard');
  await page.getByLabel('Display name').blur();
  const save = page.getByRole('button', { name: 'Save' });
  await save.dispatchEvent('click');
  await save.dispatchEvent('click');
  await expect(save).toBeDisabled();
  expect(creates).toBe(1);
  release();
  await expect(page).toHaveURL(fieldsPath);
  expect(apiRequestPaths(page).filter((path) => path.startsWith('POST ') && path.endsWith('/custom-field-definitions'))).toHaveLength(1);
});

test('focuses a specific validation error and preserves the invalid draft', async ({ page }, testInfo) => {
  await page.goto(`${fieldsPath}/new`);
  await page.getByLabel('Display name').fill('Condition');
  await page.getByLabel('Stable key').fill('condition');
  await page.getByRole('button', { name: 'List' }).click();
  await page.getByRole('button', { name: 'Save' }).click();
  const error = page.getByRole('alert');
  await expect(error).toContainText('Add at least one list option');
  await expect(error).toBeFocused();
  await expect(page.getByLabel('Display name')).toHaveValue('Condition');
  await page.screenshot({ path: testInfo.outputPath(`field-validation-${testInfo.project.name}.png`), fullPage: true });
});

test('shows inherited and local records without implying inherited ownership', async ({ page }, testInfo) => {
  seedSettingsManagementState(page);
  await page.goto(fieldsPath);
  await expect(page.getByRole('region', { name: 'From Home' }).getByText('Purchased on')).toBeVisible();
  await expect(page.getByRole('region', { name: 'Only in Household' }).getByText('Serial number')).toBeVisible();
  await page.getByRole('link', { name: /Purchased on/ }).click();
  await expect(page.getByRole('dialog', { name: 'Purchased on' })).toContainText('Inherited from Home');
  await page.screenshot({ path: testInfo.outputPath(`field-inherited-${testInfo.project.name}.png`), fullPage: true });
});

test('presents settings as read only for a viewer role', async ({ page }, testInfo) => {
  seedSettingsManagementState(page);
  setHomeInventoryPermissions(page, ['view']);
  await page.goto(fieldsPath);
  await expect(page.getByRole('heading', { name: 'Custom fields' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add Field' })).toHaveCount(0);
  await expect(page.getByText('Serial number')).toBeVisible();
  await page.goto(`${fieldsPath}/field-serial/edit`);
  await expect(page.getByRole('dialog', { name: /Edit Serial number/ })).toContainText('Read only');
  await expect(page.getByLabel('Display name')).toBeDisabled();
  await page.screenshot({ path: testInfo.outputPath(`field-denied-${testInfo.project.name}.png`), fullPage: true });
});

test('archives a tag and keeps blocked field lifecycle failures in context', async ({ page }, testInfo) => {
  seedSettingsManagementState(page);
  await page.goto(`${tagsPath}/tag-workshop/archive`);
  await expect(page.getByRole('alertdialog', { name: 'Archive tag' })).toBeVisible();
  await page.getByRole('button', { name: 'Archive', exact: true }).click();
  await expect(page).toHaveURL(tagsPath);
  await expect(page.getByText('Workshop')).toHaveCount(0);

  await page.route('http://127.0.0.1:18080/**/custom-field-definitions/field-serial/archive', async (route) => {
    await route.fulfill({ status: 409, contentType: 'application/json', body: JSON.stringify({ error: { code: 'definition_in_use', message: 'Archive is blocked while this field is required.' } }) });
  });
  await page.goto(`${fieldsPath}/field-serial/archive`);
  await expect(page.getByRole('alertdialog', { name: 'Archive custom field' })).toBeVisible();
  await page.getByRole('button', { name: 'Archive', exact: true }).click();
  await expect(page.getByRole('alert')).toBeVisible();
  await expect(page).toHaveURL(`${fieldsPath}/field-serial/archive`);
  await expect(page.getByRole('alertdialog')).toContainText('Serial number');
  await page.screenshot({ path: testInfo.outputPath(`field-lifecycle-blocked-${testInfo.project.name}.png`), fullPage: true });
});

test('renders archived and empty schema state', async ({ page }, testInfo) => {
  await page.goto(`${fieldsPath}?lifecycle=archived`);
  await expect(page.getByRole('link', { name: 'Archived' })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByText('No archived custom fields')).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath(`fields-archived-empty-${testInfo.project.name}.png`), fullPage: true });
});

test('keeps settings collection requests bounded across route state changes', async ({ page }) => {
  await page.goto(fieldsPath);
  await expect(page.getByRole('heading', { name: 'Custom fields', exact: true })).toBeVisible();
  await expect.poll(() => apiRequestPaths(page).filter((path) => path.includes('/custom-field-definitions?limit=50')).length).toBe(1);
  await expect.poll(() => apiRequestPaths(page).filter((path) => path.includes('/custom-asset-types?limit=50')).length).toBe(1);

  await page.getByRole('link', { name: 'Archived' }).click();
  await expect(page).toHaveURL(`${fieldsPath}?lifecycle=archived`);
  await expect.poll(() => apiRequestPaths(page).filter((path) => path.includes('/custom-field-definitions?limit=50')).length).toBe(2);
  expect(apiRequestPaths(page).filter((path) => path.includes('/custom-asset-types?limit=50'))).toHaveLength(1);

  await page.goto('/settings');
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible();
  await page.waitForTimeout(100);
  expect(apiRequestPaths(page).filter((path) => path.includes('/custom-field-definitions?limit=50'))).toHaveLength(2);
  expect(apiRequestPaths(page).filter((path) => path.includes('/custom-asset-types?limit=50'))).toHaveLength(1);
});

test('does not duplicate a pending supporting type load when lifecycle changes', async ({ page }) => {
  let release!: () => void;
  const pending = new Promise<void>((resolve) => { release = resolve; });
  let typeLoads = 0;
  await page.route('http://127.0.0.1:18080/**/custom-asset-types**', async (route) => {
    const url = new URL(route.request().url());
    if (url.searchParams.get('limit') !== '50') return route.fallback();
    typeLoads += 1;
    await pending;
    await route.fallback();
  });

  await page.goto(fieldsPath);
  await expect.poll(() => typeLoads).toBe(1);
  await page.getByRole('link', { name: 'Archived' }).click();
  await expect(page).toHaveURL(`${fieldsPath}?lifecycle=archived`);
  await page.waitForTimeout(100);
  const typeLoadsBeforeRelease = typeLoads;
  release();
  await expect(page.getByText('No archived custom fields')).toBeVisible();
  expect(typeLoadsBeforeRelease).toBe(1);
  expect(typeLoads).toBe(1);
});

test('retries a failed field list without reloading successful supporting types', async ({ page }) => {
  let fieldLoads = 0;
  let typeLoads = 0;
  await page.route('http://127.0.0.1:18080/**/custom-asset-types**', async (route) => {
    const url = new URL(route.request().url());
    if (url.searchParams.get('limit') === '50') typeLoads += 1;
    await route.fallback();
  });
  await page.route('http://127.0.0.1:18080/**/custom-field-definitions**', async (route) => {
    const url = new URL(route.request().url());
    if (url.searchParams.get('limit') !== '50') return route.fallback();
    fieldLoads += 1;
    if (fieldLoads === 1) {
      await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: { code: 'unavailable', message: 'Temporarily unavailable.' } }) });
      return;
    }
    await route.fallback();
  });

  await page.goto(fieldsPath);
  await expect(page.getByRole('alert')).toContainText('Custom fields unavailable');
  expect(fieldLoads).toBe(1);
  expect(typeLoads).toBe(1);
  await page.getByRole('button', { name: 'Try again' }).click();
  await expect(page.getByText('No active custom fields')).toBeVisible();
  expect(fieldLoads).toBe(2);
  expect(typeLoads).toBe(1);
});

test('renders honest loading and error states', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'State screenshots only need one wide reference.');
  let release!: () => void;
  const pending = new Promise<void>((resolve) => { release = resolve; });
  await page.route('http://127.0.0.1:18080/**/custom-field-definitions**', async (route) => {
    const url = new URL(route.request().url());
    if (url.searchParams.get('limit') !== '50') return route.fallback();
    await pending;
    await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: { code: 'unavailable', message: 'Temporarily unavailable.' } }) });
  });

  await page.goto(fieldsPath);
  await expect(page.getByRole('status')).toContainText('Loading active custom fields');
  await page.screenshot({ path: testInfo.outputPath('fields-loading-desktop.png'), fullPage: true });
  release();
  await expect(page.getByRole('alert')).toContainText('Custom fields unavailable');
  await page.screenshot({ path: testInfo.outputPath('fields-error-desktop.png'), fullPage: true });
});

test('retries supporting asset types after their first field-editor load fails', async ({ page }) => {
  let assetTypeLoads = 0;
  await page.route('http://127.0.0.1:18080/**/custom-asset-types**', async (route) => {
    const url = new URL(route.request().url());
    if (url.searchParams.get('limit') !== '50') return route.fallback();
    assetTypeLoads += 1;
    if (assetTypeLoads === 1) {
      await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: { code: 'unavailable', message: 'Temporarily unavailable.' } }) });
      return;
    }
    await route.fallback();
  });

  await page.goto(fieldsPath);
  await expect(page.getByRole('alert')).toContainText('Custom fields unavailable');
  await page.getByRole('button', { name: 'Try again' }).click();
  await expect(page.getByText('No active custom fields')).toBeVisible();
  expect(assetTypeLoads).toBe(2);
});
