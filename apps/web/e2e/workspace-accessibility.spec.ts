import { expect, test } from '@playwright/test';
import { installAuthenticatedWorkspace, resetWorkspaceApiState } from './workspace-fixture';

test.beforeEach(async ({ page }) => {
  resetWorkspaceApiState(page);
  await installAuthenticatedWorkspace(page);
});

test('desktop context switcher and nav expose keyboard-friendly current state', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Desktop accessibility coverage runs on the desktop project.');

  await page.goto('/');

  const destinationNav = page.getByRole('navigation', { name: 'Inventory destinations' });
  await expect(destinationNav.getByRole('link', { name: /Home/ })).toHaveAttribute('aria-current', 'page');

  await expect(page.getByRole('button', { name: /Household/ })).toHaveCount(1);
  const contextTrigger = page.getByRole('button', { name: /Household/ });
  await contextTrigger.focus();
  await page.keyboard.press('Enter');

  const dialog = page.getByRole('dialog', { name: 'Inventory context' });
  await expect(dialog).toBeVisible();
  await expect(page.getByLabel('Inventories').getByRole('link', { name: /Household/ })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByLabel('Inventories').getByRole('link', { name: /Household/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household'
  );

  await page.keyboard.press('Escape');
  await expect(dialog).toBeHidden();
});

test('add tray exposes named modal controls and closes from the keyboard', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Add tray accessibility coverage runs on desktop.');

  await page.goto('/');

  await page.getByRole('button', { name: 'Add', exact: true }).click();
  await page.locator('#header-add-menu').getByRole('link', { name: 'Item', exact: true }).click();

  const dialog = page.getByRole('dialog', { name: 'Add item' });
  await expect(dialog).toBeVisible();
  await expect(page.locator('.product-shell')).toHaveAttribute('aria-hidden', 'true');
  expect(await page.locator('.product-shell').evaluate((element) => (element as HTMLElement & { inert: boolean }).inert)).toBe(true);
  await expect(dialog.getByRole('button', { name: 'Item', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await expect(dialog.getByRole('button', { name: 'Container', exact: true })).toHaveAttribute('aria-pressed', 'false');
  await expect(page.getByLabel('Item name')).toBeVisible();
  await expect(page.getByRole('group', { name: 'Parent target current destination' })).toBeVisible();
  await expect(page.getByRole('group', { name: 'Parent target suggested destinations' })).toBeVisible();
  await expect(page.getByRole('group', { name: 'Photo actions' })).toBeVisible();

  await page.keyboard.press('Escape');
  await expect(dialog).toBeHidden();
  await expect(page.locator('.product-shell')).not.toHaveAttribute('aria-hidden', 'true');
  expect(await page.locator('.product-shell').evaluate((element) => (element as HTMLElement & { inert: boolean }).inert)).toBe(false);
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household');
});

test('search suggestions and image results keep ordinary accessible navigation', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'Search accessibility coverage uses the desktop header search.');

  await page.goto('/');

  const search = page.getByLabel('Search this inventory');
  await search.fill('Tomato');

  const suggestion = page.getByLabel('Search suggestions').getByRole('link', { name: 'Open Tomato fertilizer' });
  await expect(suggestion).toBeVisible();
  await expect(page.getByLabel('Search suggestions').locator('img[alt="Tomato fertilizer"]')).toBeVisible();
  await search.press('ArrowDown');
  await expect(suggestion).toBeFocused();
  await suggestion.press('Escape');
  await expect(search).toBeFocused();
  await page.getByRole('button', { name: 'Run search' }).click();

  await expect(page.getByRole('heading', { name: 'Search' })).toBeVisible();
  await expect(page.locator('.asset-list').getByRole('link', { name: /Tomato fertilizer/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato'
  );
  await expect(page.locator('.asset-list img[alt="Tomato fertilizer"]')).toBeVisible();
});

test('location and detail flows expose named links, headings, and image alt text', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop-chromium', 'List and detail accessibility coverage runs on desktop.');

  await page.goto('/');

  await page.getByRole('link', { name: 'Open location Garage' }).click();
  await expect(page.getByRole('heading', { name: 'Garage' })).toBeVisible();
  await expect(page.getByRole('link', { name: /Tomato fertilizer/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato'
  );

  await page.getByRole('link', { name: /Tomato fertilizer/ }).click();
  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();
  await expect(page.locator('.asset-photo-panel img[alt="Tomato fertilizer"]')).toBeVisible();
  await expect(page.getByRole('link', { name: /Back/ })).toHaveAttribute(
    'href',
    '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
  );
});

test('asset detail photo upload action is not duplicated on mobile', async ({ page }, testInfo) => {
  await page.goto('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato');

  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();
  const visibleAddPhotoActions = await page.getByRole('button', { name: 'Add photo' }).evaluateAll((buttons) =>
    buttons.filter((button) => {
      const styles = getComputedStyle(button);
      const box = button.getBoundingClientRect();
      return styles.display !== 'none' && styles.visibility !== 'hidden' && box.width > 0 && box.height > 0;
    }).length
  );

  expect(visibleAddPhotoActions).toBe(testInfo.project.name === 'mobile-chromium' ? 1 : 2);
});

test('mobile workspace controls provide comfortable touch targets', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile target-size coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato');

  await expect(page.getByRole('heading', { name: 'Tomato fertilizer' })).toBeVisible();
  expect(await visibleControlsMeetTarget(page.locator('.mobile-nav [data-slot="button"]'), 44)).toBe(true);
  expect(await visibleControlsMeetTarget(page.locator('.detail-actions [data-slot="button"]'), 44)).toBe(true);

  await page.getByRole('button', { name: /Household/ }).click();
  await expect(page.getByRole('dialog', { name: 'Inventory context' })).toBeVisible();
  expect(await visibleControlsMeetTarget(page.locator('.mobile-context-menu [data-slot="button"]'), 44)).toBe(true);
  await page.keyboard.press('Escape');

  await page.getByRole('link', { name: 'Add asset' }).click();
  await expect(page.getByRole('dialog', { name: 'Add item' })).toBeVisible();
  expect(await visibleControlsMeetTarget(page.locator('.add-tray [data-slot="button"]'), 44)).toBe(true);
});

test('mobile context sheet keeps a named modal surface', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile accessibility coverage runs on the mobile project.');

  await page.goto('/');

  await page.getByRole('button', { name: /Household/ }).click();

  const dialog = page.getByRole('dialog', { name: 'Inventory context' });
  await expect(dialog).toBeVisible();
  await expect(dialog).toHaveAttribute('aria-modal', 'true');
  await expect(page.locator('.workspace-route-content')).toHaveAttribute('aria-hidden', 'true');
  await expect(page.locator('.mobile-nav-shell')).toHaveAttribute('aria-hidden', 'true');
  expect(await page.locator('.workspace-route-content').evaluate((element) => (element as HTMLElement & { inert: boolean }).inert)).toBe(true);
  expect(await page.locator('.mobile-nav-shell').evaluate((element) => (element as HTMLElement & { inert: boolean }).inert)).toBe(true);
  await expect(page.getByLabel('Inventories').getByRole('link', { name: /Household/ })).toBeVisible();

  await page.keyboard.press('Escape');
  await expect(dialog).toBeHidden();
  await expect(page.locator('.workspace-route-content')).not.toHaveAttribute('aria-hidden', 'true');
  await expect(page.locator('.mobile-nav-shell')).not.toHaveAttribute('aria-hidden', 'true');
  expect(await page.locator('.workspace-route-content').evaluate((element) => (element as HTMLElement & { inert: boolean }).inert)).toBe(false);
  expect(await page.locator('.mobile-nav-shell').evaluate((element) => (element as HTMLElement & { inert: boolean }).inert)).toBe(false);
});

test('mobile settings access keeps long principals inside the viewport', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile settings layout coverage runs on the mobile project.');

  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/access');

  await expect(page.getByRole('heading', { name: 'Sharing' })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link')).toHaveCount(5);
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: /Overview/ })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: /Access/ })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: /Fields/ })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: /Activity/ })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: /Admin/ })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', {
    name: /Overview\s+Inventory context and access summary/
  })).toBeVisible();
  expect(await page.locator('.settings-section-nav').evaluate((element) => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
  expect(await page.locator('.settings-section-link').evaluateAll((elements) =>
    elements.every((element) => element.getBoundingClientRect().height <= 58)
  )).toBe(true);
  expect(await page.locator('.settings-section-link strong').evaluateAll((elements) =>
    elements.every((element) => {
      const box = element.getBoundingClientRect();
      return box.width > 0 && box.height > 0 && element.scrollWidth <= element.clientWidth + 1;
    })
  )).toBe(true);
  expect(await page.locator('.settings-section-link small').evaluateAll((elements) =>
    elements.every((element) => {
      const styles = getComputedStyle(element);
      return styles.position === 'absolute' && element.getBoundingClientRect().width <= 1;
    })
  )).toBe(true);
  await expect(page.getByLabel('Direct grants').getByText('oidc_OuQU94grMoaZ8cly6ZUUpXUVhloLanDNZ')).toBeVisible();
  const disabledGrantAction = page.getByRole('button', { name: 'Grant access' });
  await expect(disabledGrantAction).toBeDisabled();
  const disabledGrantStyles = await disabledGrantAction.evaluate((element) => {
    const styles = getComputedStyle(element);
    const rootStyles = getComputedStyle(document.documentElement);
    const probe = document.createElement('span');
    probe.style.color = rootStyles.getPropertyValue('--muted-foreground');
    probe.style.backgroundColor = rootStyles.getPropertyValue('--secondary');
    document.body.append(probe);
    const probeStyles = getComputedStyle(probe);
    const expected = { backgroundColor: probeStyles.backgroundColor, color: probeStyles.color };
    probe.remove();
    return { backgroundColor: styles.backgroundColor, color: styles.color, opacity: styles.opacity, expected };
  });
  expect(disabledGrantStyles.opacity).toBe('1');
  expect(disabledGrantStyles.backgroundColor).toBe(disabledGrantStyles.expected.backgroundColor);
  expect(disabledGrantStyles.color).toBe(disabledGrantStyles.expected.color);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)).toBe(true);
  expect(await page.locator('.settings-panel').filter({ hasText: 'Sharing' }).first().evaluate((element) => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
});

async function visibleControlsMeetTarget(locator: import('@playwright/test').Locator, minimum: number): Promise<boolean> {
  return locator.evaluateAll((elements, minSize) =>
    {
      const visibleElements = elements.filter((element) => {
        const styles = getComputedStyle(element);
        const box = element.getBoundingClientRect();
        return styles.display !== 'none' && styles.visibility !== 'hidden' && box.width > 0 && box.height > 0;
      });
      return visibleElements.length > 0 && visibleElements.every((element) => {
        const box = element.getBoundingClientRect();
        return box.width >= minSize && box.height >= minSize;
      });
    },
    minimum
  );
}

test('mobile settings sections stay compact in phone landscape width', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'mobile-chromium', 'Mobile settings layout coverage runs on the mobile project.');

  await page.setViewportSize({ width: 844, height: 390 });
  await page.goto('/tenants/tenant-home/inventories/inventory-household/settings/fields');

  await expect(page.getByRole('heading', { name: 'Custom fields' })).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link')).toHaveCount(5);
  expect(await page.locator('.settings-section-nav').evaluate((element) => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
  expect(await page.locator('.settings-section-link').evaluateAll((elements) =>
    elements.every((element) => element.getBoundingClientRect().height <= 58)
  )).toBe(true);
  expect(await page.locator('.settings-section-link strong').evaluateAll((elements) =>
    elements.every((element) => {
      const box = element.getBoundingClientRect();
      return box.width > 0 && box.height > 0 && element.scrollWidth <= element.clientWidth + 1;
    })
  )).toBe(true);
});
