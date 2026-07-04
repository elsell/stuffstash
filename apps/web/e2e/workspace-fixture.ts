import type { Page, Route } from '@playwright/test';

const apiOrigin = 'http://127.0.0.1:18080';
const sessionKey = 'stuffstash.oidc.session';
const workspaceStates = new WeakMap<Page, WorkspaceApiState>();

type WorkspaceApiState = {
  signedUploadPutCount: number;
  assetOverrides: Record<string, AssetOverride>;
  createdAssets: Record<string, CreatedAsset>;
  pendingUploads: Record<string, PendingUpload>;
  uploadedPhotos: Record<string, UploadedPhoto>;
  thumbnailRequestPaths: string[];
  lastAssetPatch: AssetPatch | null;
};

type AssetOverride = {
  title?: string;
  parentAssetId?: string | null;
};

type CreatedAsset = {
  id: string;
  title: string;
  parentAssetId: string | null;
  kind: string;
};

type UploadedPhoto = {
  id: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
};

type PendingUpload = UploadedPhoto & {
  uploadId: string;
  assetId: string;
  url: string;
  putCompleted: boolean;
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

export function thumbnailRequestPaths(page: Page): string[] {
  return workspaceState(page).thumbnailRequestPaths;
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
    const pending = Object.values(state.pendingUploads).find((upload) => upload.url === route.request().url());
    if (!pending || route.request().method() !== 'PUT' || route.request().headers()['content-type'] !== pending.contentType) {
      await route.fulfill({ status: 400 });
      return;
    }
    pending.putCompleted = true;
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
    createdAssets: {},
    pendingUploads: {},
    uploadedPhotos: {},
    thumbnailRequestPaths: [],
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
    const created = {
      id: assetIdForTitle(body.title),
      title: body.title,
      parentAssetId: body.parentAssetId ?? null,
      kind: body.kind
    };
    state.createdAssets[created.id] = created;
    await fulfill(route, createdAssetResponse(created, state), 201);
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/search/assets') {
    await fulfill(route, searchResults(state, url.searchParams.get('q') ?? ''));
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
    await fulfill(route, createdAssetResponse(state.createdAssets['asset-photo-tape'] ?? {
      id: 'asset-photo-tape',
      title: 'Photo tape',
      parentAssetId: null,
      kind: 'item'
    }, state));
    return;
  }
  const createdAssetId = createdAssetIdFromPath(path);
  if (method === 'GET' && createdAssetId && state.createdAssets[createdAssetId]) {
    await fulfill(route, createdAssetResponse(state.createdAssets[createdAssetId], state));
    return;
  }
  if (method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/location-garage') {
    await fulfill(route, asset('location-garage', 'tenant-home', 'inventory-household', 'Garage', null, 'active', false, 'location'));
    return;
  }
  if (method === 'POST' && path.match(/^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/[^/]+\/attachments\/direct-uploads$/)) {
    const body = (await request.postDataJSON()) as { contentType?: string; fileName?: string; sizeBytes?: number };
    const assetId = path.split('/')[6] ?? 'asset-photo-tape';
    const uploadId = `upload-${assetId}`;
    const pending: PendingUpload = {
      uploadId,
      id: 'attachment-photo',
      assetId,
      fileName: body.fileName ?? `${assetId}.jpg`,
      contentType: body.contentType ?? 'image/jpeg',
      sizeBytes: body.sizeBytes ?? 1024,
      url: `https://uploads.local/${assetId}/object-one`,
      putCompleted: false
    };
    state.pendingUploads[uploadId] = pending;
    await fulfill(
      route,
      {
        uploadId,
        attachmentId: pending.id,
        method: 'PUT',
        url: pending.url,
        headers: { 'Content-Type': pending.contentType },
        formFields: {},
        expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString()
      },
      201
    );
    return;
  }
  if (
    method === 'POST' &&
    path.match(/^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/[^/]+\/attachments\/direct-uploads\/[^/]+\/complete$/)
  ) {
    const match = path.match(
      /^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/([^/]+)\/attachments\/direct-uploads\/([^/]+)\/complete$/
    );
    const assetId = match?.[1] ?? 'asset-photo-tape';
    const uploadId = match?.[2] ?? '';
    const pending = state.pendingUploads[uploadId];
    if (!pending || pending.assetId !== assetId || !pending.putCompleted) {
      await route.fulfill({
        status: 409,
        contentType: 'application/json',
        body: JSON.stringify({ error: { code: 'upload_incomplete', message: 'Upload has not completed.' } })
      });
      return;
    }
    state.uploadedPhotos[assetId] = {
      id: pending.id,
      fileName: pending.fileName,
      contentType: pending.contentType,
      sizeBytes: pending.sizeBytes
    };
    await fulfill(route, attachment(pending.id, 'tenant-home', 'inventory-household', assetId, pending), 201);
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
    state.thumbnailRequestPaths.push(`${path}?${url.searchParams.toString()}`);
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
    asset('asset-bin', 'tenant-home', 'inventory-household', 'Green storage bin', 'location-garage', 'active', false, 'container'),
    ...Object.values(state.createdAssets).map((created) => createdAssetResponse(created, state))
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

function createdAssetResponse(created: CreatedAsset, state: WorkspaceApiState): object {
  return asset(
    created.id,
    'tenant-home',
    'inventory-household',
    created.title,
    created.parentAssetId,
    'active',
    state.uploadedPhotos[created.id] ?? false,
    created.kind
  );
}

function searchResults(state: WorkspaceApiState, query: string): object[] {
  const normalized = query.trim().toLowerCase();
  const candidates = [tomatoAsset(state), ...Object.values(state.createdAssets).map((created) => createdAssetResponse(created, state))];
  return candidates
    .filter((candidate) => {
      const title = typeof candidate === 'object' && candidate && 'title' in candidate ? String(candidate.title) : '';
      return !normalized || title.toLowerCase().includes(normalized);
    })
    .map((candidate) => {
      const title = typeof candidate === 'object' && candidate && 'title' in candidate ? String(candidate.title) : '';
      return {
        type: 'asset',
        tenantId: 'tenant-home',
        inventory: { id: 'inventory-household', name: 'Household' },
        asset: candidate,
        matches: [{ field: 'title', value: title }]
      };
    });
}

function createdAssetIdFromPath(path: string): string | null {
  const match = path.match(/^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/([^/]+)$/);
  return match?.[1] ?? null;
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
  primaryPhoto: boolean | UploadedPhoto = false,
  kind = 'item'
): object {
  const photo = primaryPhoto === true ? defaultPhoto(id) : primaryPhoto || null;
  return {
    id,
    tenantId,
    inventoryId,
    kind,
    title,
    description: kind === 'location' ? 'Storage and seasonal items.' : '',
    parentAssetId,
    lifecycleState,
    primaryPhoto: photo
      ? {
          id: photo.id,
          fileName: photo.fileName,
          contentType: photo.contentType,
          sizeBytes: photo.sizeBytes,
          thumbnails: {
            small: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/${photo.id}/thumbnail?variant=small`,
            medium: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/${photo.id}/thumbnail?variant=medium`,
            large: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/${photo.id}/thumbnail?variant=large`
          }
        }
      : undefined
  };
}

function defaultPhoto(assetId: string): UploadedPhoto {
  return {
    id: 'attachment-photo',
    fileName: `${assetId}.jpg`,
    contentType: 'image/jpeg',
    sizeBytes: 1024
  };
}

function attachment(id: string, tenantId: string, inventoryId: string, assetId: string, photo: UploadedPhoto): object {
  return {
    id,
    tenantId,
    inventoryId,
    assetId,
    fileName: photo.fileName,
    contentType: photo.contentType,
    sizeBytes: photo.sizeBytes,
    lifecycleState: 'active'
  };
}

function assetIdForTitle(title: string): string {
  return `asset-${title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')}`;
}
