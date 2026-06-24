import { beforeEach, describe, expect, it } from 'vitest';
import { StuffStashInventoryRepository } from './stuffStashInventoryRepository';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';

const config = {
  apiBaseUrl: 'http://api.local',
  oidcIssuer: 'http://oidc.local',
  oidcClientId: 'web',
  oidcRedirectUri: 'http://web.local/auth/callback'
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
});

function fakeFetch(): { fetch: typeof fetch; requests: Request[] } {
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
      return Response.json({ error: { code: 'not_found', message: `Unhandled ${request.method} ${path}` } }, { status: 404 });
    }
  };
}

function envelope(data: unknown): Response {
  return Response.json({
    data,
    meta: {
      pagination: Array.isArray(data) ? { limit: 50, nextCursor: null, hasMore: false } : undefined
    }
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
