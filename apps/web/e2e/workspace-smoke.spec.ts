import { expect, test, type Page, type Route } from '@playwright/test';

const apiOrigin = 'http://localhost:8080';
const sessionKey = 'stuffstash.oidc.session';
let signedUploadPutCount = 0;

test.beforeEach(async ({ page }) => {
  signedUploadPutCount = 0;
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
  await expect(page.getByRole('dialog', { name: 'Inventory context' })).toBeVisible();
  await expect(page.getByLabel('Inventories').getByRole('link', { name: /Household/ })).toBeVisible();
  await page.keyboard.press('Escape');

  await page.getByRole('link', { name: 'Add asset' }).click();
  await expect(page.getByRole('dialog', { name: 'Add item' })).toBeVisible();
  await expect(page.getByLabel('Item name')).toBeVisible();
  await expect(page.getByLabel('Find parent')).toBeVisible();
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
    name: 'front.jpg',
    mimeType: 'image/jpeg',
    buffer: Buffer.from('fake-photo')
  });
  await expect(page.locator('.photo-preview img[alt="front.jpg"]')).toBeVisible();
  await expect(page.getByLabel('Photo actions').getByText('1 photo')).toBeVisible();
  await page.getByRole('button', { name: 'Save item' }).click();
  await expect(page).toHaveURL('/tenants/tenant-home/inventories/inventory-household/assets/asset-photo-tape');
  await expect(page.getByRole('heading', { name: 'Photo tape' })).toBeVisible();
  await expect(page.locator('.asset-photo-panel img[alt="Photo tape"]')).toBeVisible();
  expect(signedUploadPutCount).toBe(1);
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

async function installAuthenticatedWorkspace(page: Page): Promise<void> {
  await page.addInitScript((key) => {
    window.sessionStorage.setItem(
      key,
      JSON.stringify({
        idToken: 'e2e-token',
        expiresAt: Date.now() + 60 * 60 * 1000
      })
    );
  }, sessionKey);

  await page.route(`${apiOrigin}/**`, routeApiRequest);
  await page.route('https://uploads.local/**', async (route) => {
    if (route.request().method() !== 'PUT' || route.request().headers()['content-type'] !== 'image/jpeg') {
      await route.fulfill({ status: 400 });
      return;
    }
    signedUploadPutCount += 1;
    await route.fulfill({ status: 204 });
  });
}

async function routeApiRequest(route: Route): Promise<void> {
  const request = route.request();
  const url = new URL(request.url());
  const path = url.pathname;
  const method = request.method();

  if (request.headers().authorization !== 'Bearer e2e-token') {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({ error: { code: 'unauthenticated', message: 'Authentication required.' } })
    });
    return;
  }

  if (method === 'GET' && path === '/me') {
    await fulfill(route, { id: 'principal-owner', email: 'owner@example.com' });
    return;
  }
  if (method === 'GET' && path === '/me/tenants') {
    await fulfill(route, [
      tenant('tenant-home', 'Home', ['view', 'create_inventory', 'configure']),
      tenant('tenant-cabin', 'Cabin', ['view'])
    ]);
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories') {
    await fulfill(route, [
      inventory('inventory-household', 'tenant-home', 'Household', ['view', 'create_asset', 'edit_asset', 'configure'])
    ]);
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-cabin/inventories') {
    await fulfill(route, [inventory('inventory-cabin', 'tenant-cabin', 'Cabin Gear', ['view'])]);
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets') {
    await fulfill(route, activeAssets());
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-cabin/inventories/inventory-cabin/assets') {
    await fulfill(route, [asset('asset-lantern', 'tenant-cabin', 'inventory-cabin', 'Lantern')]);
    return;
  }
  if (method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets') {
    const body = (await request.postDataJSON()) as { kind: string; title: string; parentAssetId?: string | null };
    await fulfill(
      route,
      asset(assetIdForTitle(body.title), 'tenant-home', 'inventory-household', body.title, body.parentAssetId ?? null, 'active', false, body.kind),
      201
    );
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/search/assets') {
    await fulfill(route, [
      {
        type: 'asset',
        tenantId: 'tenant-home',
        inventory: { id: 'inventory-household', name: 'Household' },
        asset: asset('asset-tomato', 'tenant-home', 'inventory-household', 'Tomato fertilizer', 'location-garage', 'active', true),
        matches: [{ field: 'title', value: 'Tomato fertilizer' }]
      }
    ]);
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato') {
    await fulfill(route, asset('asset-tomato', 'tenant-home', 'inventory-household', 'Tomato fertilizer', 'location-garage', 'active', true));
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-cordless-drill') {
    await fulfill(route, asset('asset-cordless-drill', 'tenant-home', 'inventory-household', 'Cordless drill'));
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-photo-tape') {
    await fulfill(route, asset('asset-photo-tape', 'tenant-home', 'inventory-household', 'Photo tape', null, 'active', true));
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/location-garage') {
    await fulfill(route, asset('location-garage', 'tenant-home', 'inventory-household', 'Garage', null, 'active', false, 'location'));
    return;
  }
  if (method === 'POST' && path.match(/^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/[^/]+\/attachments\/direct-uploads$/)) {
    await fulfill(
      route,
      {
        uploadId: 'upload-photo',
        attachmentId: 'attachment-photo',
        method: 'PUT',
        url: 'https://uploads.local/object-one',
        headers: { 'Content-Type': 'image/jpeg' },
        formFields: {},
        expiresAt: '2026-07-03T10:00:00Z'
      },
      201
    );
    return;
  }
  if (
    method === 'POST' &&
    path.match(/^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/[^/]+\/attachments\/direct-uploads\/upload-photo\/complete$/)
  ) {
    if (signedUploadPutCount === 0) {
      await route.fulfill({
        status: 409,
        contentType: 'application/json',
        body: JSON.stringify({ error: { code: 'upload_incomplete', message: 'Upload has not completed.' } })
      });
      return;
    }
    const assetId = path.split('/')[6] ?? 'asset-photo-tape';
    await fulfill(route, attachment('attachment-photo', 'tenant-home', 'inventory-household', assetId, 'front.jpg'), 201);
    return;
  }
  if (method === 'GET' && path.endsWith('/attachments')) {
    await fulfill(route, []);
    return;
  }
  if (method === 'GET' && path.endsWith('/custom-asset-types')) {
    await fulfill(route, []);
    return;
  }
  if (method === 'GET' && path.endsWith('/custom-field-definitions')) {
    await fulfill(route, []);
    return;
  }
  if (method === 'GET' && path.endsWith('/thumbnail')) {
    await route.fulfill({
      status: 200,
      contentType: 'image/svg+xml',
      body: '<svg xmlns="http://www.w3.org/2000/svg" width="80" height="80"><rect width="80" height="80" fill="#2f6f4e"/></svg>'
    });
    return;
  }

  await route.fulfill({
    status: 404,
    contentType: 'application/json',
    body: JSON.stringify({ error: { code: 'not_found', message: `Unhandled ${method} ${path}` } })
  });
}

function activeAssets(): object[] {
  return [
    asset('location-garage', 'tenant-home', 'inventory-household', 'Garage', null, 'active', false, 'location'),
    asset('asset-tomato', 'tenant-home', 'inventory-household', 'Tomato fertilizer', 'location-garage', 'active', true),
    asset('asset-bin', 'tenant-home', 'inventory-household', 'Green storage bin', 'location-garage', 'active', false, 'container')
  ];
}

async function fulfill(route: Route, data: unknown, status = 200): Promise<void> {
  await route.fulfill({
    status,
    contentType: 'application/json',
    body: JSON.stringify({
      data,
      meta: {
        pagination: Array.isArray(data) ? { limit: 50, nextCursor: null, hasMore: false } : undefined
      }
    })
  });
}

function tenant(id: string, name: string, permissions: string[]): object {
  return {
    id,
    name,
    access: { relationship: permissions.includes('create_inventory') ? 'owner' : 'viewer', permissions }
  };
}

function inventory(id: string, tenantId: string, name: string, permissions: string[]): object {
  return {
    id,
    tenantId,
    name,
    access: { relationship: permissions.includes('create_asset') ? 'editor' : 'viewer', permissions }
  };
}

function asset(
  id: string,
  tenantId: string,
  inventoryId: string,
  title: string,
  parentAssetId: string | null = null,
  lifecycleState = 'active',
  withPrimaryPhoto = false,
  kind = 'item'
): object {
  return {
    id,
    tenantId,
    inventoryId,
    kind,
    title,
    description: kind === 'location' ? 'Storage and seasonal items.' : '',
    parentAssetId,
    lifecycleState,
    primaryPhoto: withPrimaryPhoto
      ? {
          id: 'attachment-photo',
          fileName: `${id}.jpg`,
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: {
            small: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/attachment-photo/thumbnail?variant=small`,
            medium: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/attachment-photo/thumbnail?variant=medium`,
            large: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/attachment-photo/thumbnail?variant=large`
          }
        }
      : undefined
  };
}

function attachment(id: string, tenantId: string, inventoryId: string, assetId: string, fileName: string): object {
  return {
    id,
    tenantId,
    inventoryId,
    assetId,
    fileName,
    contentType: 'image/jpeg',
    sizeBytes: 1024,
    lifecycleState: 'active'
  };
}

function assetIdForTitle(title: string): string {
  return `asset-${title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')}`;
}
