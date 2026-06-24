import { beforeEach, describe, expect, it } from 'vitest';
import { StuffStashInventoryRepository } from './stuffStashInventoryRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
import type { RuntimeConfig } from '$lib/runtimeConfig';

const config: RuntimeConfig = {
  apiBaseUrl: 'http://api.local',
  oidcIssuer: 'http://oidc.local',
  oidcClientId: 'web',
  oidcRedirectUri: 'http://web.local/auth/callback',
  mediaUploadPolicy: {
    supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
    maxBytes: 5242880
  }
};

describe('StuffStashInventoryRepository', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it('restores the browser-session tenant and inventory selection before loading active assets', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-cabin');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-cabin');
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();

    expect(data.context.selectedTenantId).toBe('tenant-cabin');
    expect(data.context.selectedInventoryId).toBe('inventory-cabin');
    expect(data.assets.map((asset) => asset.id)).toEqual(['asset-lantern']);
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/me',
      'GET http://api.local/me/tenants?limit=50',
      'GET http://api.local/tenants/tenant-cabin/inventories?limit=50',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/assets?limit=100&lifecycleState=active'
    ]);
  });

  it('keeps an empty tenant selected and clears the selected inventory without listing assets', async () => {
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.selectTenant('tenant-empty');

    expect(data.context.selectedTenantId).toBe('tenant-empty');
    expect(data.context.selectedInventoryId).toBe('');
    expect(data.context.inventories).toEqual([]);
    expect(data.assets).toEqual([]);
    expect(sessionStorage.getItem('stuffstash.selectedTenantId')).toBe('tenant-empty');
    expect(sessionStorage.getItem('stuffstash.selectedInventoryId')).toBeNull();
    expect(requests.map((request) => request.url)).not.toContain(
      'http://api.local/tenants/tenant-empty/inventories/inventory-household/assets?limit=100&lifecycleState=active'
    );
  });

  it('creates a starter inventory inside the selected tenant and reloads that inventory', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.createInventory('tenant-empty', 'Household');

    expect(data.context.selectedTenantId).toBe('tenant-empty');
    expect(data.context.selectedInventoryId).toBe('inventory-created');
    expect(data.context.inventories.map((inventory) => inventory.id)).toEqual(['inventory-created']);
    expect(await requests.find((request) => request.method === 'POST')?.json()).toEqual({ name: 'Household' });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/me',
      'GET http://api.local/me/tenants?limit=50',
      'POST http://api.local/tenants/tenant-empty/inventories',
      'GET http://api.local/tenants/tenant-empty/inventories?limit=50',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/assets?limit=100&lifecycleState=active'
    ]);
  });

  it('loads asset detail by ID through the generated client path', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const asset = await repository.getAsset('tenant-home', 'inventory-household', 'asset-passport');

    expect(asset).toMatchObject({
      id: 'asset-passport',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      title: 'Passport',
      parentAssetId: 'asset-closet'
    });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport'
    ]);
  });

  it('updates asset detail and movement through the generated client path', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const asset = await repository.updateAsset('tenant-home', 'inventory-household', 'asset-passport', {
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: 'asset-safe'
    });

    expect(asset).toMatchObject({
      id: 'asset-passport',
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: 'asset-safe'
    });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport'
    ]);
    expect(await requests[0]?.json()).toEqual({
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: 'asset-safe'
    });
  });

  it('loads archived assets through the generated client lifecycle query', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.selectAssetLifecycle('tenant-home', 'inventory-household', 'archived');

    expect(data.context.assetLifecycleState).toBe('archived');
    expect(data.assets.map((asset) => asset.id)).toEqual(['asset-archived']);
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/me',
      'GET http://api.local/me/tenants?limit=50',
      'GET http://api.local/tenants/tenant-home/inventories?limit=50',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets?limit=100&lifecycleState=archived'
    ]);
  });

  it('archives, restores, and deletes assets through generated client lifecycle paths', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.archiveAsset('tenant-home', 'inventory-household', 'asset-passport')).resolves.toMatchObject({
      id: 'asset-passport',
      lifecycleState: 'archived'
    });
    await expect(repository.restoreAsset('tenant-home', 'inventory-household', 'asset-passport')).resolves.toMatchObject({
      id: 'asset-passport',
      lifecycleState: 'active'
    });
    await expect(repository.deleteAsset('tenant-home', 'inventory-household', 'asset-passport')).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/archive',
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/restore',
      'DELETE http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport'
    ]);
  });

  it('direct uploads selected photos before completing attachment metadata', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const file = new File(['fake image'], 'photo.jpg', { type: 'image/jpeg' });

    const attachment = await repository.uploadAssetPhoto('tenant-home', 'inventory-household', 'asset-passport', {
      id: 'photo-one',
      name: 'photo.jpg',
      sizeBytes: file.size,
      contentType: 'image/jpeg',
      previewUrl: 'blob:photo-one',
      file
    });

    expect(attachment).toMatchObject({
      id: 'attachment-one',
      assetId: 'asset-passport',
      fileName: 'photo.jpg',
      contentType: 'image/jpeg'
    });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'PUT https://uploads.local/object-one',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads/upload-one/complete',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail?variant=small'
    ]);
    expect(await requests[0]?.json()).toEqual({
      fileName: 'photo.jpg',
      contentType: 'image/jpeg',
      sizeBytes: file.size
    });
    expect(requests[1]?.headers.get('Content-Type')).toBe('image/jpeg');
    expect(requests[1]?.body).not.toBeNull();
    expect(requests[3]?.headers.get('Authorization')).toBe('Bearer id-token');
  });

  it('rejects direct upload targets that browsers cannot fetch', async () => {
    const { fetch } = fakeFetch({ directUploadUrl: 'stuffstash-local://direct-uploads/upload-one' });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const file = new File(['fake image'], 'photo.jpg', { type: 'image/jpeg' });

    await expect(
      repository.uploadAssetPhoto('tenant-home', 'inventory-household', 'asset-passport', {
        id: 'photo-one',
        name: 'photo.jpg',
        sizeBytes: file.size,
        contentType: 'image/jpeg',
        previewUrl: 'blob:photo-one',
        file
      })
    ).rejects.toThrow('Direct upload target is not available in this browser.');
  });

  it('lists attachments with authenticated thumbnail object URLs', async () => {
    const { fetch } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const attachments = await repository.listAssetAttachments('tenant-home', 'inventory-household', 'asset-passport');

    expect(attachments).toMatchObject([
      {
        id: 'attachment-one',
        fileName: 'photo.jpg',
        thumbnailUrl: expect.stringContaining('blob:'),
        thumbnailHeaders: { Authorization: 'Bearer id-token' }
      }
    ]);
  });

  it('searches with lifecycle and mode options through the generated client path', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const results = await repository.searchAssets({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: 'Passport',
      lifecycleState: 'archived',
      mode: 'exact'
    });

    expect(results).toMatchObject([{ asset: { id: 'asset-passport', lifecycleState: 'archived' } }]);
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/search/assets?q=Passport&limit=20&lifecycleState=archived&mode=exact'
    ]);
  });

  it('manages access grants through generated client-backed repository methods', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.listInventoryAccessGrants('tenant-home', 'inventory-household')).resolves.toMatchObject({
      items: [
        {
          tenantId: 'tenant-home',
          inventoryId: 'inventory-household',
          principalId: 'principal-two',
          relationship: 'viewer'
        }
      ],
      pagination: { limit: 50, nextCursor: null, hasMore: false }
    });
    await expect(
      repository.grantInventoryAccess('tenant-home', 'inventory-household', 'principal-three', 'editor')
    ).resolves.toMatchObject({ principalId: 'principal-three', relationship: 'editor' });
    await expect(
      repository.revokeInventoryAccess('tenant-home', 'inventory-household', 'principal-two', 'viewer')
    ).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/access-grants?limit=50',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/access-grants',
      'DELETE http://api.local/tenants/tenant-home/inventories/inventory-household/access-grants/principal-two/viewer'
    ]);
    expect(await requests[1]?.json()).toEqual({ principalId: 'principal-three', relationship: 'editor' });
  });

  it('manages access invitations through generated client-backed repository methods', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.listInventoryAccessInvitations('tenant-home', 'inventory-household', 'pending')).resolves.toMatchObject({
      items: [expect.objectContaining({ id: 'invite-one', email: 'friend@example.test', relationship: 'viewer' })],
      pagination: { limit: 50, nextCursor: null, hasMore: false }
    });
    await expect(
      repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'editor@example.test', 'editor')
    ).resolves.toMatchObject({
      invitation: { email: 'editor@example.test', relationship: 'editor' },
      acceptanceToken: 'raw-token'
    });
    await expect(
      repository.updateInventoryAccessInvitationExpiration(
        'tenant-home',
        'inventory-household',
        'invite-one',
        '2026-07-01T00:00:00Z'
      )
    ).resolves.toMatchObject({ id: 'invite-one', expiresAt: '2026-07-01T00:00:00Z' });
    await expect(repository.cancelInventoryAccessInvitation('tenant-home', 'inventory-household', 'invite-one')).resolves.toBeUndefined();
    await expect(repository.deleteInventoryAccessInvitation('tenant-home', 'inventory-household', 'invite-one')).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations?limit=50&status=pending',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations',
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one/expiration',
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one/cancel',
      'DELETE http://api.local/tenants/tenant-home/inventories/inventory-household/access-invitations/invite-one'
    ]);
    expect(await requests[1]?.json()).toEqual({ email: 'editor@example.test', relationship: 'editor' });
    expect(await requests[2]?.json()).toEqual({ expiresAt: '2026-07-01T00:00:00Z' });
  });
});

function fakeFetch(options: { directUploadUrl?: string } = {}): { fetch: typeof fetch; requests: Request[] } {
  const requests: Request[] = [];
  return {
    requests,
    fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
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
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets') {
        return envelope([asset('asset-archived', 'tenant-home', 'inventory-household', 'Archived Passport', null, 'archived')]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-empty/inventories/inventory-created/assets') {
        return envelope([]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport') {
        return envelope(asset('asset-passport', 'tenant-home', 'inventory-household', 'Passport', 'asset-closet'));
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
            method: 'PUT',
            url: options.directUploadUrl ?? 'https://uploads.local/object-one',
            headers: { 'Content-Type': 'image/jpeg' },
            formFields: {},
            expiresAt: '2026-06-23T00:15:00Z'
          },
          201
        );
      }
      if (request.method === 'PUT' && request.url === 'https://uploads.local/object-one') {
        return new Response(null, { status: 204 });
      }
      if (request.method === 'POST' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads/upload-one/complete') {
        return envelope(attachment('attachment-one', 'tenant-home', 'inventory-household', 'asset-passport', 'photo.jpg'), 201);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments') {
        return envelope([attachment('attachment-one', 'tenant-home', 'inventory-household', 'asset-passport', 'photo.jpg')]);
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail') {
        return new Response(new Blob(['thumbnail'], { type: 'image/jpeg' }), { status: 200 });
      }
      if (request.method === 'GET' && path === '/tenants/tenant-home/search/assets') {
        return envelope([
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
            asset: asset('asset-passport', 'tenant-home', 'inventory-household', 'Passport', null, 'archived'),
            matches: [{ field: 'title', value: 'Passport' }]
          }
        ]);
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
      return Response.json({ error: { code: 'not_found', message: `Unhandled ${request.method} ${path}` } }, { status: 404 });
    }
  };
}

function envelope(data: unknown, status = 200): Response {
  return Response.json({
    data,
    meta: {
      pagination: Array.isArray(data) ? { limit: 50, nextCursor: null, hasMore: false } : undefined
    }
  }, { status });
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
  lifecycleState = 'active'
): object {
  return {
    id,
    tenantId,
    inventoryId,
    kind: 'item',
    title,
    description: '',
    parentAssetId,
    lifecycleState
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
