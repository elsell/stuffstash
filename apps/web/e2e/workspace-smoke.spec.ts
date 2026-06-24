import { expect, test } from '@playwright/test';

test('desktop shell loads seeded tenant and inventory workspace', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Desktop shell coverage runs on the desktop project.');

  await page.goto('/');

  await expect(page.getByText('Local demo data is showing.')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Locations' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Open location Garage' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add', exact: true })).toBeEnabled();
});

test('mobile shell loads and opens the add tray', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile shell coverage runs on the mobile project.');

  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Locations' })).toBeVisible();
  await page.getByRole('button', { name: 'Add asset' }).click();
  await expect(page.getByLabel('Name')).toBeVisible();
  await expect(page.getByLabel('Create a new parent inside that place')).toBeVisible();
});

test('search entry returns seeded inventory results', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Search smoke uses the desktop header search.');

  await page.goto('/');

  await page.getByLabel('Search this inventory').fill('Tomato');
  await page.getByRole('button', { name: 'Run search' }).click();

  await expect(page.getByRole('heading', { name: 'Search' })).toBeVisible();
  await expect(page.getByRole('button', { name: /Tomato fertilizer/ })).toBeVisible();
});

test('location navigation opens asset detail and returns to the location list', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Location detail smoke runs on desktop first.');

  await page.goto('/');

  await page.getByRole('button', { name: 'Open location Garage' }).click();
  await expect(page.getByRole('heading', { name: 'Garage' })).toBeVisible();

  await page.getByRole('button', { name: /Tomato fertilizer/ }).click();
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();

  await page.getByRole('button', { name: /Back/ }).click();
  await expect(page.getByRole('heading', { name: 'Garage' })).toBeVisible();
});
