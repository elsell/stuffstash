import type { Page, Route } from '@playwright/test';

const apiOrigin = 'http://127.0.0.1:18080';
const sessionKey = 'stuffstash.oidc.session';
const workspaceStates = new WeakMap<Page, WorkspaceApiState>();

type WorkspaceApiState = {
  signedUploadPutCount: number;
  assetOverrides: Record<string, AssetOverride>;
  lastAssetPatch: AssetPatch | null;
};

type AssetOverride = {
  title?: string;
  parentAssetId?: string | null;
};

export type AssetPatch = {
  assetId: string;
  title?: string;
  description?: string;
  parentAssetId?: string | null;
};

export function resetWorkspaceApiState(page: Page): void {
  workspaceStates.set(page, freshWorkspaceApiState());
}

export function signedUploadPuts(page: Page): number {
  return workspaceState(page).signedUploadPutCount;
}

export function lastAssetPatch(page: Page): AssetPatch | null {
  return workspaceState(page).lastAssetPatch;
}

export async function installAuthenticatedWorkspace(page: Page): Promise<void> {
  const state = workspaceState(page);

  await page.addInitScript((key) => {
    window.sessionStorage.setItem(
      key,
      JSON.stringify({
        idToken: 'e2e-token',
        expiresAt: Date.now() + 60 * 60 * 1000
      })
    );
  }, sessionKey);

  await page.route('**/config.json', routeRuntimeConfig);
  await page.route(`${apiOrigin}/**`, (route) => routeApiRequest(route, state));
  await page.route('https://uploads.local/**', async (route) => {
    if (route.request().method() !== 'PUT' || route.request().headers()['content-type'] !== 'image/jpeg') {
      await route.fulfill({ status: 400 });
      return;
    }
    state.signedUploadPutCount += 1;
    await route.fulfill({ status: 204 });
  });
}

async function routeRuntimeConfig(route: Route): Promise<void> {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      apiBaseUrl: apiOrigin,
      oidcIssuer: 'http://127.0.0.1:5556/dex',
      oidcClientId: 'stuff-stash-web-e2e',
      oidcRedirectUri: 'http://127.0.0.1:5197/callback',
      mediaUploadPolicy: {
        supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
        maxBytes: 5 * 1024 * 1024
      }
    })
  });
}

function workspaceState(page: Page): WorkspaceApiState {
  let state = workspaceStates.get(page);
  if (!state) {
    state = freshWorkspaceApiState();
    workspaceStates.set(page, state);
  }
  return state;
}

function freshWorkspaceApiState(): WorkspaceApiState {
  return {
    signedUploadPutCount: 0,
    assetOverrides: {},
    lastAssetPatch: null
  };
}

async function routeApiRequest(route: Route, state: WorkspaceApiState): Promise<void> {
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
    await fulfill(route, activeAssets(state));
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
        asset: tomatoAsset(state),
        matches: [{ field: 'title', value: tomatoTitle(state) }]
      }
    ]);
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato') {
    await fulfill(route, tomatoAsset(state));
    return;
  }
  if (method === 'PATCH' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-tomato') {
    const body = (await request.postDataJSON()) as Omit<AssetPatch, 'assetId'>;
    state.lastAssetPatch = { assetId: 'asset-tomato', ...body };
    state.assetOverrides['asset-tomato'] = {
      ...state.assetOverrides['asset-tomato'],
      title: body.title ?? state.assetOverrides['asset-tomato']?.title,
      parentAssetId: Object.prototype.hasOwnProperty.call(body, 'parentAssetId')
        ? body.parentAssetId ?? null
        : state.assetOverrides['asset-tomato']?.parentAssetId
    };
    await fulfill(route, tomatoAsset(state));
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
        expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString()
      },
      201
    );
    return;
  }
  if (
    method === 'POST' &&
    path.match(/^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/[^/]+\/attachments\/direct-uploads\/upload-photo\/complete$/)
  ) {
    if (state.signedUploadPutCount === 0) {
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

function activeAssets(state: WorkspaceApiState): object[] {
  return [
    asset('location-garage', 'tenant-home', 'inventory-household', 'Garage', null, 'active', false, 'location'),
    tomatoAsset(state),
    asset('asset-bin', 'tenant-home', 'inventory-household', 'Green storage bin', 'location-garage', 'active', false, 'container')
  ];
}

function tomatoAsset(state: WorkspaceApiState): object {
  const override = state.assetOverrides['asset-tomato'];
  return asset(
    'asset-tomato',
    'tenant-home',
    'inventory-household',
    tomatoTitle(state),
    override?.parentAssetId === undefined ? 'location-garage' : override.parentAssetId,
    'active',
    true
  );
}

function tomatoTitle(state: WorkspaceApiState): string {
  return state.assetOverrides['asset-tomato']?.title ?? 'Tomato fertilizer';
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
