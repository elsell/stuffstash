import type { RuntimeConfig } from '$lib/runtimeConfig';

export const config: RuntimeConfig = {
  apiBaseUrl: 'http://api.local',
  oidcIssuer: 'http://oidc.local',
  oidcClientId: 'web',
  oidcRedirectUri: 'http://web.local/auth/callback',
  mediaUploadPolicy: {
    supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
    maxBytes: 5242880
  }
};

export function fakeFetch(
  options: {
    directUploadUrl?: string;
    directUploadMethod?: string;
    directUploadHeaders?: Record<string, string>;
    directUploadFormFields?: Record<string, string>;
    directUploadRejected?: boolean;
    directUploadThrows?: boolean;
    primaryPhotoAssetIds?: string[];
    rejectedThumbnailAssetIds?: string[];
    failedThumbnailStatusByAssetId?: Record<string, number>;
    includeUnphotographedContainerAndLocation?: boolean;
    failedImportOperations?: Array<'list' | 'preview' | 'start' | 'cancel' | 'remove'>;
  } = {}
): { fetch: typeof fetch; requests: Request[] } {
  const requests: Request[] = [];
  return {
    requests,
    fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
      const inputUrl = input instanceof Request ? input.url : input.toString();
      if ((init?.body instanceof FormData) && inputUrl === 'https://uploads.local/object-one') {
        const headers = new Headers(init.headers);
        if (!headers.has('Content-Type')) {
          headers.set('Content-Type', 'multipart/form-data; boundary=stuffstash-test');
        }
        const request = new Request(input, { ...init, body: undefined, headers });
        Object.defineProperty(request, 'capturedFormData', { value: init.body });
        requests.push(request);
        if (options.directUploadThrows) {
          throw new TypeError('Failed to fetch');
        }
        return new Response(null, { status: options.directUploadRejected ? 403 : 204 });
      }
      const request = new Request(input, init);
      requests.push(request);
      const url = new URL(request.url);
      const path = url.pathname;

      if (request.method === 'GET' && path === '/me') {
        return envelope({ id: 'principal-one', email: 'person@example.test' });
      }
      if (request.method === 'GET' && path === '/me/tenants') {
        return envelope([
          tenant('tenant-home', 'Home', ['view', 'create_inventory']),
          tenant('tenant-cabin', 'Cabin', ['view']),
          tenant('tenant-empty', 'Empty', ['view', 'create_inventory'])
        ]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories') {
        return envelope([inventory('inventory-household', 'tenant-home', 'Household', ['view', 'create_asset'])]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-cabin/inventories') {
        return envelope([inventory('inventory-cabin', 'tenant-cabin', 'Cabin Gear', ['view'])]);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-empty/inventories') {
        return envelope(inventory('inventory-created', 'tenant-empty', 'Household', ['view', 'create_asset']));
      }
      if (request.method === 'GET' && path === '/tenants/tenant-empty/inventories') {
        const created = requests.some(
          (candidate) => candidate.method === 'POST' && new URL(candidate.url).pathname === '/tenants/tenant-empty/inventories'
        );
        return envelope(created ? [inventory('inventory-created', 'tenant-empty', 'Household', ['view', 'create_asset'])] : []);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-cabin/inventories/inventory-cabin/assets') {
        return envelope([asset('asset-lantern', 'tenant-cabin', 'inventory-cabin', 'Lantern')]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-cabin/inventories/inventory-cabin/checked-out-assets') {
        return envelope([]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-cabin/inventories/inventory-cabin/tags') {
        return envelope([]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets') {
        const assets = [
          asset(
            'asset-archived',
            'tenant-home',
            'inventory-household',
            'Archived Passport',
            null,
            'archived',
            options.primaryPhotoAssetIds?.includes('asset-archived') ?? false
          )
        ];
        if (options.includeUnphotographedContainerAndLocation) {
          assets.push(
            asset('asset-container', 'tenant-home', 'inventory-household', 'Toolbox', null, 'active', false, 'container'),
            asset('asset-location', 'tenant-home', 'inventory-household', 'Garage', null, 'active', false, 'location')
          );
        }
        return envelope(assets);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/tags') {
        return envelope([assetTag('tag-workshop', 'workshop', 'Workshop', '#2F80ED')]);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/tags') {
        const body = (await request.clone().json()) as { displayName: string; color?: string };
        return envelope(assetTag('tag-created', 'fragile', body.displayName, body.color), 201);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/checked-out-assets') {
        const checkout = currentCheckout('checkout-open');
        return envelope([
          {
            asset: {
              ...asset('asset-archived', 'tenant-home', 'inventory-household', 'Archived Passport', null, 'archived'),
              currentCheckout: checkout
            },
            checkout
          }
        ]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-empty/inventories/inventory-created/assets') {
        return envelope([]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-empty/inventories/inventory-created/checked-out-assets') {
        return envelope([]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-empty/inventories/inventory-created/tags') {
        return envelope([]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport') {
        return envelope(
          asset(
            'asset-passport',
            'tenant-home',
            'inventory-household',
            'Passport',
            'asset-closet',
            'active',
            options.primaryPhotoAssetIds?.includes('asset-passport') ?? false
          )
        );
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/checkout') {
        const body = (await request.clone().json()) as { details?: string };
        return envelope(assetCheckout('checkout-open', 'open', body.details), 201);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/return') {
        const body = (await request.clone().json()) as { details?: string };
        return envelope({
          ...assetCheckout('checkout-open', 'returned', 'using at desk'),
          returnedAt: '2026-06-24T12:00:00Z',
          returnedByPrincipalId: 'principal-one',
          returnDetails: body.details,
          updatedAt: '2026-06-24T12:00:00Z'
        });
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/checkouts') {
        return envelope([assetCheckout('checkout-open', 'open', 'using at desk')]);
      }
      if (request.method === 'PATCH' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport') {
        const body = (await request.clone().json()) as { title: string; description?: string; parentAssetId?: string | null };
        return envelope({
          ...asset('asset-passport', 'tenant-home', 'inventory-household', body.title, body.parentAssetId ?? null),
          description: body.description ?? ''
        });
      }
      if (request.method === 'PATCH' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/archive') {
        return envelope(asset('asset-passport', 'tenant-home', 'inventory-household', 'Passport', null, 'archived'));
      }
      if (request.method === 'PATCH' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/restore') {
        return envelope(asset('asset-passport', 'tenant-home', 'inventory-household', 'Passport'));
      }
      if (request.method === 'DELETE' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport') {
        return new Response(null, { status: 204 });
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads') {
        return envelope(
          {
            uploadId: 'upload-one',
            attachmentId: 'attachment-one',
            method: options.directUploadMethod ?? 'PUT',
            url: options.directUploadUrl ?? 'https://uploads.local/object-one',
            headers: options.directUploadHeaders ?? { 'Content-Type': 'image/jpeg' },
            formFields: options.directUploadFormFields ?? {},
            expiresAt: '2026-06-23T00:15:00Z'
          },
          201
        );
      }
      if ((request.method === 'PUT' || request.method === 'POST') && request.url === 'https://uploads.local/object-one') {
        if (options.directUploadThrows) {
          throw new TypeError('Failed to fetch');
        }
        return new Response(null, { status: options.directUploadRejected ? 403 : 204 });
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads/upload-one/complete') {
        return envelope(attachment('attachment-one', 'tenant-home', 'inventory-household', 'asset-passport', 'photo.jpg'), 201);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments') {
        return envelope(attachment('attachment-one', 'tenant-home', 'inventory-household', 'asset-passport', 'photo.jpg'), 201);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments') {
        return envelope([attachment('attachment-one', 'tenant-home', 'inventory-household', 'asset-passport', 'photo.jpg')]);
      }
      const thumbnailAssetID = matchingThumbnailAssetID(path, url.searchParams);
      if (request.method === 'GET' && thumbnailAssetID) {
        if (options.rejectedThumbnailAssetIds?.includes(thumbnailAssetID)) {
          throw new Error('Thumbnail fetch failed.');
        }
        const failedStatus = options.failedThumbnailStatusByAssetId?.[thumbnailAssetID];
        if (failedStatus) {
          return new Response('Thumbnail unavailable.', { status: failedStatus });
        }
        return new Response(new Blob(['thumbnail'], { type: 'image/jpeg' }), { status: 200 });
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/search/assets') {
        const tenantWideResults = [
          {
            type: 'asset',
            tenantId: 'tenant-home',
            inventory: { id: 'inventory-other', name: 'Other' },
            asset: asset('asset-other', 'tenant-home', 'inventory-other', 'Passport', null, 'archived'),
            matches: [{ field: 'title', value: 'Passport' }]
          },
          {
            type: 'asset',
            tenantId: 'tenant-home',
            inventory: { id: 'inventory-household', name: 'Household' },
            asset: asset(
              'asset-passport',
              'tenant-home',
              'inventory-household',
              'Passport',
              null,
              'archived',
              options.primaryPhotoAssetIds?.includes('asset-passport') ?? false
            ),
            matches: [{ field: 'title', value: 'Passport' }]
          }
        ];
        const results =
          url.searchParams.get('inventoryId') === 'inventory-household'
            ? tenantWideResults.filter((result) => result.inventory.id === 'inventory-household')
            : tenantWideResults;
        if (options.includeUnphotographedContainerAndLocation) {
          results.push(
            {
              type: 'asset',
              tenantId: 'tenant-home',
              inventory: { id: 'inventory-household', name: 'Household' },
              asset: asset('asset-container', 'tenant-home', 'inventory-household', 'Toolbox', null, 'active', false, 'container'),
              matches: [{ field: 'title', value: 'Toolbox' }]
            },
            {
              type: 'asset',
              tenantId: 'tenant-home',
              inventory: { id: 'inventory-household', name: 'Household' },
              asset: asset('asset-location', 'tenant-home', 'inventory-household', 'Garage', null, 'active', false, 'location'),
              matches: [{ field: 'title', value: 'Garage' }]
            }
          );
        }
        return envelope(results);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/imports/jobs') {
        if (options.failedImportOperations?.includes('list')) return importFailure();
        return envelope({ jobs: [importJob('import-job-one')] });
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/imports/jobs/preview') {
        if (options.failedImportOperations?.includes('preview')) return importFailure();
        return envelope(importJob('import-job-one', await importJobSourceFromRequest(request)), 201);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/imports/jobs/import-job-one/start') {
        if (options.failedImportOperations?.includes('start')) return importFailure();
        return envelope({ ...importJob('import-job-one', await importJobSourceFromRequest(request)), status: 'running' });
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/imports/jobs/import-job-one/cancel') {
        if (options.failedImportOperations?.includes('cancel')) return importFailure();
        return envelope({ ...importJob('import-job-one'), status: 'cancel_requested', cancellationMode: 'discard_partial_progress' });
      }
      if (request.method === 'DELETE' && path === '/tenants/tenant-home/inventories/inventory-household/imports/jobs/import-job-one') {
        if (options.failedImportOperations?.includes('remove')) return importFailure();
        return new Response(null, { status: 204 });
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/access-grants') {
        return envelope([
          {
            tenantId: 'tenant-home',
            inventoryId: 'inventory-household',
            principalId: 'principal-two',
            relationship: 'viewer'
          }
        ]);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/access-grants') {
        const body = (await request.clone().json()) as { principalId: string; relationship: string };
        return envelope(
          {
            tenantId: 'tenant-home',
            inventoryId: 'inventory-household',
            principalId: body.principalId,
            relationship: body.relationship
          },
          201
        );
      }
      if (
        request.method === 'DELETE' &&
        path === '/tenants/tenant-home/inventories/inventory-household/access-grants/principal-two/viewer'
      ) {
        return new Response(null, { status: 204 });
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/access-invitations') {
        return envelope([invitation('invite-one', 'friend@example.test', 'viewer')]);
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/access-invitations') {
        const body = (await request.clone().json()) as { email: string; relationship: string };
        return envelope({ ...invitation('invite-created', body.email, body.relationship), acceptanceToken: 'raw-token' }, 201);
      }
      if (
        request.method === 'PATCH' &&
        path === '/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one/expiration'
      ) {
        const body = (await request.clone().json()) as { expiresAt: string };
        return envelope({ ...invitation('invite-one', 'friend@example.test', 'viewer'), expiresAt: body.expiresAt });
      }
      if (
        request.method === 'PATCH' &&
        path === '/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one/cancel'
      ) {
        return new Response(null, { status: 204 });
      }
      if (
        request.method === 'DELETE' &&
        path === '/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one'
      ) {
        return new Response(null, { status: 204 });
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/audit-records') {
        return envelope([auditRecord()]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/audit-records') {
        return envelope([auditRecord()]);
      }
      if (request.method === 'GET' && path.endsWith('/custom-asset-types')) {
        return envelope([]);
      }
      if (request.method === 'GET' && path.endsWith('/custom-field-definitions')) {
        return envelope([]);
      }
      return Response.json({ error: { code: 'not_found', message: `Unhandled ${request.method} ${path}` } }, { status: 404 });
    }
  };
}

function matchingThumbnailAssetID(path: string, searchParams: URLSearchParams): string | null {
  const matches = path.match(
    /^\/tenants\/tenant-home\/inventories\/inventory-household\/assets\/(asset-passport|asset-archived)\/attachments\/attachment-one\/thumbnail$/
  );
  if (!matches || (searchParams.get('variant') !== 'small' && searchParams.get('variant') !== 'medium')) {
    return null;
  }
  return matches[1] ?? null;
}

function envelope(data: unknown, status = 200): Response {
  return Response.json({
    data,
    meta: {
      pagination: Array.isArray(data) ? { limit: 50, nextCursor: null, hasMore: false } : undefined
    }
  }, { status });
}

function importFailure(): Response {
  return Response.json(
    {
      error: {
        code: 'import_failed',
        message: 'Import operation failed.',
        detail: 'provider-stacktrace password=secret'
      }
    },
    { status: 500 }
  );
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
    description: '',
    parentAssetId,
    lifecycleState,
    tags: id === 'asset-passport' || id === 'asset-archived' ? [assetTag('tag-workshop', 'workshop', 'Workshop', '#2F80ED')] : [],
    primaryPhoto: withPrimaryPhoto
      ? {
          id: 'attachment-one',
          fileName: 'photo.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 10,
          thumbnails: {
            small: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/attachment-one/thumbnail?variant=small`,
            medium: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/attachment-one/thumbnail?variant=medium`,
            large: `/tenants/${tenantId}/inventories/${inventoryId}/assets/${id}/attachments/attachment-one/thumbnail?variant=large`
          }
        }
      : undefined
  };
}

function assetTag(id: string, key: string, displayName: string, color?: string): object {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    key,
    displayName,
    color,
    lifecycleState: 'active',
    createdAt: '2026-07-07T12:00:00Z',
    updatedAt: '2026-07-07T12:00:00Z'
  };
}

function currentCheckout(id: string): object {
  return {
    id,
    state: 'open',
    checkedOutAt: '2026-06-24T11:00:00Z',
    checkedOutByPrincipalId: 'principal-one'
  };
}

function assetCheckout(id: string, state: string, checkoutDetails?: string): object {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    assetId: 'asset-passport',
    state,
    checkedOutAt: '2026-06-24T11:00:00Z',
    checkedOutByPrincipalId: 'principal-one',
    checkoutDetails,
    createdAt: '2026-06-24T11:00:00Z',
    updatedAt: '2026-06-24T11:00:00Z'
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
    sizeBytes: 10,
    lifecycleState: 'active'
  };
}

async function importJobSourceFromRequest(request: Request): Promise<object> {
  const body = (await request.clone().json()) as {
    sourceType?: string;
    baseUrl?: string;
    includeImages?: boolean;
    allowPrivateNetwork?: boolean;
    allowInsecureTLS?: boolean;
    fileName?: string;
  };
  if (body.sourceType === 'legacy_homebox_csv') {
    return {
      type: 'legacy_homebox_csv',
      name: 'Homebox CSV',
      imageImport: 'disabled',
      allowPrivateNetwork: false,
      allowInsecureTLS: false,
      fingerprint: 'csv-fingerprint-one'
    };
  }
  return {
    type: 'legacy_homebox',
    name: 'Homebox',
    baseUrl: body.baseUrl ?? 'https://stuff.jsksell.com',
    imageImport: body.includeImages === false ? 'disabled' : 'enabled',
    allowPrivateNetwork: body.allowPrivateNetwork ?? false,
    allowInsecureTLS: body.allowInsecureTLS ?? false,
    fingerprint: 'fingerprint-one'
  };
}

function importJob(id: string, source: object = liveHomeboxImportSource()): object {
  return {
    id,
    status: 'previewed',
    source,
    counts: {
      fields: 0,
      locations: 0,
      assets: 0,
      attachments: 0,
      warnings: 0,
      errors: 0,
      fieldsCreated: 0,
      fieldsExisting: 0,
      locationsCreated: 0,
      assetsCreated: 0,
      assetsSkipped: 0,
      attachmentsCreated: 0,
      attachmentsSkipped: 0,
      recordsDiscarded: 0,
      sourceLinksDiscarded: 0
    },
    progress: {
      phase: 'ready',
      done: 0,
      total: 0,
      updatedAt: '2026-07-06T00:00:00Z'
    },
    progressHistory: [
      {
        phase: 'ready',
        done: 0,
        total: 0,
        updatedAt: '2026-07-06T00:00:00Z'
      }
    ],
    createdAt: '2026-07-06T00:00:00Z',
    updatedAt: '2026-07-06T00:00:00Z',
    resources: [],
    messages: []
  };
}

function liveHomeboxImportSource(): object {
  return {
    type: 'legacy_homebox',
    name: 'Homebox',
    baseUrl: 'https://stuff.jsksell.com',
    imageImport: 'enabled',
    allowPrivateNetwork: false,
    allowInsecureTLS: false,
    fingerprint: 'fingerprint-one'
  };
}

function invitation(id: string, email: string, relationship: string): object {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    email,
    relationship,
    status: 'pending',
    isExpired: false,
    expiresAt: '2026-06-30T00:00:00Z',
    inviterPrincipalId: 'principal-one'
  };
}

function auditRecord(): object {
  return {
    id: 'audit-one',
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    principalId: 'principal-one',
    action: 'asset.created',
    source: 'api',
    targetType: 'asset',
    targetId: 'asset-passport',
    occurredAt: '2026-06-24T12:00:00Z',
    requestId: 'request-one',
    metadata: { operation_id: 'operation-one' }
  };
}
