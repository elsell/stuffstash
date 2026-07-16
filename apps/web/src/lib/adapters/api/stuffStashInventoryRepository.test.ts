import { beforeEach, describe, expect, it, vi } from 'vitest';
import { StuffStashInventoryRepository } from './stuffStashInventoryRepository';
import { AuthenticationRequiredError } from '$lib/application/authenticationRequired';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
import { config, fakeFetch } from './stuffStashInventoryRepository.test-helpers';
import type { Asset } from '$lib/domain/inventory';

describe('StuffStashInventoryRepository workspace and assets', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it('reports API 401 responses as an authentication-required error', async () => {
    const repository = new StuffStashInventoryRepository(
      config,
      () => 'expired-token',
      new InMemoryWorkspaceObserver(),
      async () => Response.json({ error: { code: 'unauthenticated', message: 'Session expired.' } }, { status: 401 })
    );

    await expect(repository.loadWorkspace()).rejects.toBeInstanceOf(AuthenticationRequiredError);
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
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/custom-asset-types?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/custom-field-definitions?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/tags?limit=100',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/assets?limit=100&lifecycleState=active&sort=updated_desc',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/checked-out-assets?limit=100'
    ]);
  });

  it('follows tenant pagination before selecting workspace context', async () => {
    const base = fakeFetch();
    const fetchImpl: typeof fetch = async (input, init) => {
      const request = new Request(input, init);
      const url = new URL(request.url);
      if (request.method === 'GET' && url.pathname === '/me/tenants') {
        const cursor = url.searchParams.get('cursor');
        return pagedEnvelope(
          [apiTenant(cursor ? 'tenant-cabin' : 'tenant-home', cursor ? 'Cabin' : 'Home')],
          cursor ? null : 'tenant-page-two'
        );
      }
      return base.fetch(input, init);
    };
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetchImpl);

    const data = await repository.loadWorkspace();

    expect(data.context.tenants.map((tenant) => tenant.id)).toEqual(['tenant-home', 'tenant-cabin']);
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

  it('returns workspace photo metadata without blocking on thumbnail bytes', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();

    expect(data.assets[0]).toMatchObject({ id: 'asset-archived', primaryPhotoId: 'attachment-one' });
    expect(data.assets[0]?.photo).toBeUndefined();
    expect(data.context.assetTags).toEqual([{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }]);
    expect(data.assets[0]?.tags).toEqual([{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }]);
    expect(requests.some((request) => request.url.includes('/thumbnail'))).toBe(false);
  });

  it('reuses explicit primary thumbnail work for visible assets', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();
    const asset = data.assets[0]!;
    await repository.loadAssetThumbnail(asset);
    await repository.loadAssetThumbnail(asset);

    expect(requests.filter((request) => request.url.includes('/thumbnail?variant=small'))).toHaveLength(1);
  });

  it('derives cached thumbnail alt text from the current asset title', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const original = (await repository.loadWorkspace()).assets[0]!;

    await repository.loadAssetThumbnail(original);
    const renamed = await repository.loadAssetThumbnail({ ...original, title: 'Travel passport' });

    expect(renamed?.alt).toBe('Travel passport');
    expect(requests.filter((request) => request.url.includes('/thumbnail?variant=small'))).toHaveLength(1);
  });

  it('shares one in-flight thumbnail between concurrent primary and gallery callers', async () => {
    const base = fakeFetch();
    let thumbnailAttempts = 0;
    let finishThumbnail!: (response: Response) => void;
    let markStarted!: () => void;
    const started = new Promise<void>((resolve) => (markStarted = resolve));
    const fetchImpl: typeof fetch = async (input, init) => {
      const url = input instanceof Request ? input.url : input.toString();
      if (url.includes('/attachments/attachment-one/thumbnail')) {
        thumbnailAttempts += 1;
        markStarted();
        return new Promise((resolve) => (finishThumbnail = resolve));
      }
      return base.fetch(input, init);
    };
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetchImpl);
    const asset = photographedAsset('attachment-one');

    const primaryLoading = repository.loadAssetThumbnail(asset);
    const galleryLoading = repository.listAssetAttachments('tenant-home', 'inventory-household', 'asset-passport');
    await started;
    finishThumbnail(new Response(new Blob(['thumbnail'], { type: 'image/jpeg' }), { status: 200 }));
    const [primary, attachments] = await Promise.all([primaryLoading, galleryLoading]);

    expect(attachments[0]?.thumbnailUrl).toBe(primary?.url);
    expect(thumbnailAttempts).toBe(1);
  });

  it('does not start thumbnail requests after the route disposes the repository', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const asset = (await repository.loadWorkspace()).assets[0]!;

    repository.dispose();

    await expect(repository.loadAssetThumbnail(asset)).resolves.toBeNull();
    expect(requests.filter((request) => request.url.includes('/thumbnail'))).toHaveLength(0);
  });

  it('keeps gallery resources across primary A to B to null, then revokes on archive and delete', async () => {
    const base = fakeFetch();
    const fetchImpl: typeof fetch = async (input, init) => {
      const url = input instanceof Request ? input.url : input.toString();
      if (url.includes('/attachments/attachment-two/thumbnail')) {
        return new Response(new Blob(['replacement'], { type: 'image/jpeg' }), { status: 200 });
      }
      if (url.endsWith('/attachments/attachment-two')) return new Response(null, { status: 204 });
      return base.fetch(input, init);
    };
    const revokeObjectUrl = vi.spyOn(URL, 'revokeObjectURL');
    const repository = new StuffStashInventoryRepository(
      config,
      () => 'id-token',
      new InMemoryWorkspaceObserver(),
      fetchImpl
    );
    const original = await repository.loadAssetThumbnail(photographedAsset('attachment-one'));
    await repository.listAssetAttachments('tenant-home', 'inventory-household', 'asset-passport');
    const replacement = await repository.loadAssetThumbnail(photographedAsset('attachment-two'));
    await repository.loadAssetThumbnail(photographedAsset(undefined));

    expect(revokeObjectUrl).not.toHaveBeenCalled();

    await repository.archiveAssetAttachment('tenant-home', 'inventory-household', 'asset-passport', 'attachment-one');
    expect(revokeObjectUrl).toHaveBeenCalledWith(original?.url);

    await repository.deleteAssetAttachment('tenant-home', 'inventory-household', 'asset-passport', 'attachment-two');
    expect(revokeObjectUrl).toHaveBeenCalledWith(replacement?.url);

    revokeObjectUrl.mockRestore();
  });

  it('revokes a blob URL completed by active work after disposal', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const base = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    let finishThumbnail!: (response: Response) => void;
    let markThumbnailStarted!: () => void;
    const thumbnailStarted = new Promise<void>((resolve) => (markThumbnailStarted = resolve));
    const fetchImpl: typeof fetch = async (input, init) => {
      const url = input instanceof Request ? input.url : input.toString();
      if (url.includes('/attachments/attachment-one/thumbnail')) {
        markThumbnailStarted();
        return new Promise((resolve) => (finishThumbnail = resolve));
      }
      return base.fetch(input, init);
    };
    const revokeObjectUrl = vi.spyOn(URL, 'revokeObjectURL');
    const repository = new StuffStashInventoryRepository(
      config,
      () => 'id-token',
      new InMemoryWorkspaceObserver(),
      fetchImpl
    );
    const asset = (await repository.loadWorkspace()).assets[0]!;

    const loading = repository.loadAssetThumbnail(asset);
    await thumbnailStarted;
    repository.dispose();
    finishThumbnail(new Response(new Blob(['thumbnail'], { type: 'image/jpeg' }), { status: 200 }));

    await expect(loading).resolves.toBeNull();
    expect(revokeObjectUrl).toHaveBeenCalledOnce();
    revokeObjectUrl.mockRestore();
  });

  it('does not reuse a photographed item image for unphotographed containers or locations', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({
      primaryPhotoAssetIds: ['asset-archived'],
      includeUnphotographedContainerAndLocation: true
    });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();

    expect(data.assets.find((asset) => asset.id === 'asset-archived')).toMatchObject({
      id: 'asset-archived',
      kind: 'item',
      primaryPhotoId: 'attachment-one'
    });
    expect(data.assets.find((asset) => asset.id === 'asset-container')).toMatchObject({ id: 'asset-container', kind: 'container' });
    expect(data.assets.find((asset) => asset.id === 'asset-container')).not.toHaveProperty('photo');
    expect(data.assets.find((asset) => asset.id === 'asset-location')).toMatchObject({ id: 'asset-location', kind: 'location' });
    expect(data.assets.find((asset) => asset.id === 'asset-location')).not.toHaveProperty('photo');
    expect(requests.filter((request) => request.url.includes('/thumbnail'))).toHaveLength(0);
  });

  it('keeps workspace assets when a primary photo thumbnail cannot be fetched', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'], rejectedThumbnailAssetIds: ['asset-archived'] });
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);

    const data = await repository.loadWorkspace();
    await repository.loadAssetThumbnail(data.assets[0]!);

    expect(data.assets[0]?.id).toBe('asset-archived');
    expect(data.assets[0]?.photo).toBeUndefined();
    expect(observer.events).toContainEqual({
      eventName: 'workspace.asset_primary_photo_load_failed',
      attributes: { assetId: 'asset-archived' }
    });
  });

  it('retries thumbnail materialization after a transient failure', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const base = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    let thumbnailAttempts = 0;
    const fetchImpl: typeof fetch = async (input, init) => {
      const url = input instanceof Request ? input.url : input.toString();
      if (url.includes('/attachments/attachment-one/thumbnail') && ++thumbnailAttempts === 1) {
        return new Response('try again', { status: 503 });
      }
      return base.fetch(input, init);
    };
    const repository = new StuffStashInventoryRepository(
      config,
      () => 'id-token',
      new InMemoryWorkspaceObserver(),
      fetchImpl
    );
    const asset = (await repository.loadWorkspace()).assets[0]!;

    await expect(repository.loadAssetThumbnail(asset)).resolves.toBeNull();
    await expect(repository.loadAssetThumbnail(asset)).resolves.toMatchObject({ assetId: asset.id });
    expect(thumbnailAttempts).toBe(2);
  });

  it('does not let a delayed invalidated failure delete a newer successful cache entry', async () => {
    const base = fakeFetch();
    let attempts = 0;
    let finishOld!: (response: Response) => void;
    let markOldStarted!: () => void;
    const oldStarted = new Promise<void>((resolve) => (markOldStarted = resolve));
    const fetchImpl: typeof fetch = async (input, init) => {
      const url = input instanceof Request ? input.url : input.toString();
      if (url.includes('/attachments/attachment-one/thumbnail')) {
        attempts += 1;
        if (attempts === 1) {
          markOldStarted();
          return new Promise((resolve) => (finishOld = resolve));
        }
      }
      return base.fetch(input, init);
    };
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetchImpl);
    const asset = photographedAsset('attachment-one');

    const oldLoading = repository.loadAssetThumbnail(asset);
    await oldStarted;
    await repository.archiveAssetAttachment('tenant-home', 'inventory-household', 'asset-passport', 'attachment-one');
    const fresh = await repository.loadAssetThumbnail(asset);
    finishOld(new Response('try again', { status: 503 }));
    await expect(oldLoading).resolves.toBeNull();
    const reused = await repository.loadAssetThumbnail(asset);

    expect(reused?.url).toBe(fresh?.url);
    expect(attempts).toBe(2);
  });

  it('marks workspace photos unavailable when a primary photo thumbnail returns an HTTP error', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'], failedThumbnailStatusByAssetId: { 'asset-archived': 500 } });
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);

    const data = await repository.loadWorkspace();
    await repository.loadAssetThumbnail(data.assets[0]!);

    expect(data.assets[0]?.id).toBe('asset-archived');
    expect(data.assets[0]?.photo).toBeUndefined();
    expect(observer.events).toContainEqual({
      eventName: 'workspace.asset_primary_photo_load_failed',
      attributes: { assetId: 'asset-archived' }
    });
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
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/custom-asset-types?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/custom-field-definitions?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/tags?limit=100',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/assets?limit=100&lifecycleState=active&sort=updated_desc',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/checked-out-assets?limit=100'
    ]);
  });

  it('creates a tenant with its first inventory without dropping existing tenant access', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.createTenantWithInventory({ tenantName: 'Workshop', inventoryName: 'Tools' });

    expect(data.context.selectedTenantId).toBe('tenant-created');
    expect(data.context.selectedInventoryId).toBe('inventory-created-new-tenant');
    expect(data.context.tenants.map((tenant) => tenant.id)).toEqual([
      'tenant-home',
      'tenant-cabin',
      'tenant-empty',
      'tenant-created'
    ]);
    expect(data.context.inventories.map((inventory) => inventory.id)).toEqual(['inventory-created-new-tenant']);
    expect(await requests.find((request) => request.method === 'POST' && new URL(request.url).pathname === '/tenants')?.json()).toEqual({
      name: 'Workshop'
    });
    expect(
      await requests.find(
        (request) => request.method === 'POST' && new URL(request.url).pathname === '/tenants/tenant-created/inventories'
      )?.json()
    ).toEqual({ name: 'Tools' });
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

  it('hydrates API primary photos into asset detail photos', async () => {
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-passport'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const asset = await repository.getAsset('tenant-home', 'inventory-household', 'asset-passport');

    expect(asset.photo).toMatchObject({
      id: 'attachment-one',
      assetId: 'asset-passport',
      url: expect.stringContaining('blob:'),
      alt: 'Passport'
    });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail?variant=medium'
    ]);
  });

  it('keeps search result photo identities scoped to their assets without eager thumbnail requests', async () => {
    const { fetch } = fakeFetch({
      primaryPhotoAssetIds: ['asset-passport'],
      includeUnphotographedContainerAndLocation: true
    });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const results = await repository.searchAssets({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: 'passport',
      lifecycleState: 'all',
      mode: 'fuzzy'
    });

    expect(results[0]?.asset).toMatchObject({
      id: 'asset-passport',
      kind: 'item',
      primaryPhotoId: 'attachment-one'
    });
    expect(results[1]?.asset).toMatchObject({ id: 'asset-container', kind: 'container' });
    expect(results[1]?.asset).not.toHaveProperty('photo');
    expect(results[2]?.asset).toMatchObject({ id: 'asset-location', kind: 'location' });
    expect(results[2]?.asset).not.toHaveProperty('photo');
  });

  it('hydrates primary photos for Browse search results', async () => {
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-passport'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const page = await repository.browseAssets({
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'passport', tagIds: [],
      lifecycleState: 'all', checkoutState: 'any', scope: 'all', sort: 'updated_desc', mode: 'fuzzy', limit: 20
    });

    expect(page.assets[0]).toMatchObject({
      id: 'asset-passport',
      photo: expect.objectContaining({ assetId: 'asset-passport', alt: 'Passport' })
    });
    expect(page.searchResults[0]?.asset.photo).toMatchObject({ assetId: 'asset-passport' });
    expect(requests.some((request) =>
      request.url.includes('/assets/asset-passport/attachments/attachment-one/thumbnail?variant=small')
    )).toBe(true);
  });

  it('updates asset detail and movement through the generated client path', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const asset = await repository.updateAsset('tenant-home', 'inventory-household', 'asset-passport', {
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: 'asset-safe',
      tagIds: ['tag-workshop']
    });

    expect(asset).toMatchObject({
      id: 'asset-passport',
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: 'asset-safe',
      undoableOperationId: 'operation-edit-one'
    });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'PATCH http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport'
    ]);
    expect(await requests[0]?.json()).toEqual({
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: 'asset-safe',
      tagIds: ['tag-workshop']
    });
  });

  it('moves an asset with a parent-only generated update payload', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(
      repository.moveAsset('tenant-home', 'inventory-household', 'asset-passport', 'asset-safe')
    ).resolves.toMatchObject({ id: 'asset-passport', parentAssetId: 'asset-safe' });

    expect(requests).toHaveLength(1);
    expect(await requests[0]?.json()).toEqual({ parentAssetId: 'asset-safe' });
  });

  it('applies target-scoped Undo and Redo through generated client paths', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(
      repository.applyAssetOperation('tenant-home', 'inventory-household', 'operation-edit-one', 'undo')
    ).resolves.toMatchObject({ id: 'asset-passport', title: 'Passport' });
    await expect(
      repository.applyAssetOperation('tenant-home', 'inventory-household', 'operation-edit-one', 'redo')
    ).resolves.toMatchObject({ id: 'asset-passport', title: 'Updated Passport' });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/undoable-operations/operation-edit-one/undo',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/undoable-operations/operation-edit-one/redo'
    ]);
  });

  it('creates inventory tags through the generated client boundary', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const tag = await repository.createAssetTag('tenant-home', 'inventory-household', {
      displayName: 'Fragile',
      color: '#C2410C'
    });

    expect(tag).toEqual({ id: 'tag-created', key: 'fragile', displayName: 'Fragile', color: '#C2410C' });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/tags'
    ]);
    expect(await requests[0]?.json()).toEqual({ displayName: 'Fragile', color: '#C2410C' });
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
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/custom-asset-types?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/custom-field-definitions?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/tags?limit=100',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets?limit=100&lifecycleState=archived&sort=updated_desc',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/checked-out-assets?limit=100'
    ]);
  });

  it('loads checked-out assets regardless of lifecycle during workspace load', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();

    expect(data.checkedOutAssets).toMatchObject([
      {
        asset: { id: 'asset-archived', lifecycleState: 'archived', primaryPhotoId: 'attachment-one' },
        checkout: { id: 'checkout-open', state: 'open' }
      }
    ]);
    expect(requests.some((request) => request.url.includes('/thumbnail'))).toBe(false);
  });

  it('checks out, returns, and lists checkout history through generated client paths', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(
      repository.checkoutAsset('tenant-home', 'inventory-household', 'asset-passport', { details: 'using at desk' })
    ).resolves.toMatchObject({
      id: 'checkout-open',
      checkoutDetails: 'using at desk',
      state: 'open',
      undoableOperationId: 'operation-checkout-one'
    });
    await expect(repository.returnAsset('tenant-home', 'inventory-household', 'asset-passport', { details: 'back in bin' })).resolves.toMatchObject({
      id: 'checkout-open',
      state: 'returned',
      returnDetails: 'back in bin',
      undoableOperationId: 'operation-return-one'
    });
    await expect(repository.listAssetCheckoutHistory('tenant-home', 'inventory-household', 'asset-passport')).resolves.toMatchObject([
      { id: 'checkout-open', state: 'open', checkoutDetails: 'using at desk' }
    ]);

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/checkout',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/return',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/checkouts?limit=50'
    ]);
    expect(await requests[0]?.json()).toEqual({ details: 'using at desk' });
    expect(await requests[1]?.json()).toEqual({ details: 'back in bin' });
  });

  it('archives, restores, and deletes assets through generated client lifecycle paths', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.archiveAsset('tenant-home', 'inventory-household', 'asset-passport')).resolves.toMatchObject({
      id: 'asset-passport',
      lifecycleState: 'archived',
      undoableOperationId: 'operation-archive-one'
    });
    await expect(repository.restoreAsset('tenant-home', 'inventory-household', 'asset-passport')).resolves.toMatchObject({
      id: 'asset-passport',
      lifecycleState: 'active',
      undoableOperationId: 'operation-restore-one'
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

  it('sends presigned POST form fields to browser direct upload targets', async () => {
    const { fetch, requests } = fakeFetch({
      directUploadMethod: 'POST',
      directUploadHeaders: { 'Content-Type': 'image/jpeg', 'X-Test-Header': 'one' },
      directUploadFormFields: {
        key: 'tenant/inventory/asset/attachment',
        policy: 'encoded-policy',
        'x-amz-signature': 'signature'
      }
    });
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
    ).resolves.toMatchObject({ id: 'attachment-one' });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'POST https://uploads.local/object-one',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads/upload-one/complete',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail?variant=small'
    ]);
    expect(requests[1]?.headers.get('Content-Type')).toContain('multipart/form-data');
    expect(requests[1]?.headers.get('X-Test-Header')).toBe('one');
    const uploadForm = (requests[1] as (Request & { capturedFormData?: FormData }) | undefined)?.capturedFormData;
    expect(uploadForm).toBeDefined();
    expect(Array.from(uploadForm?.keys() ?? [])).toEqual(['key', 'policy', 'x-amz-signature', 'file']);
    expect(uploadForm?.get('key')).toBe('tenant/inventory/asset/attachment');
    expect(uploadForm?.get('policy')).toBe('encoded-policy');
    expect(uploadForm?.get('x-amz-signature')).toBe('signature');
    const uploadedFile = uploadForm?.get('file');
    expect(uploadedFile).toBeInstanceOf(File);
    expect((uploadedFile as File).name).toBe('photo.jpg');
    expect((uploadedFile as File).type).toBe('image/jpeg');
  });

  it('falls back to the JSON attachment upload route when direct targets are not browser-fetchable', async () => {
    const { fetch, requests } = fakeFetch({ directUploadUrl: 'stuffstash-local://direct-uploads/upload-one' });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const file = new File(['fake image'], 'photo.jpg', { type: 'image/jpeg' });

    await expect(
      repository.uploadAssetAttachment('tenant-home', 'inventory-household', 'asset-passport', {
        id: 'photo-one',
        name: 'photo.jpg',
        sizeBytes: file.size,
        contentType: 'image/jpeg',
        previewUrl: 'blob:photo-one',
        file
      })
    ).resolves.toMatchObject({ id: 'attachment-one', fileName: 'photo.jpg' });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail?variant=small'
    ]);
    expect(await requests[1]?.json()).toEqual({
      fileName: 'photo.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZSBpbWFnZQ=='
    });
  });

  it('surfaces browser direct upload target rejection instead of hiding Garage failures behind fallback', async () => {
    const { fetch, requests } = fakeFetch({ directUploadRejected: true });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const file = new File(['fake image'], 'photo.jpg', { type: 'image/jpeg' });

    await expect(
      repository.uploadAssetAttachment('tenant-home', 'inventory-household', 'asset-passport', {
        id: 'photo-one',
        name: 'photo.jpg',
        sizeBytes: file.size,
        contentType: 'image/jpeg',
        previewUrl: 'blob:photo-one',
        file
      })
    ).rejects.toThrow('Direct upload to media storage failed.');
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'PUT https://uploads.local/object-one'
    ]);
  });

  it('surfaces browser direct upload fetch failure instead of hiding Garage failures behind fallback', async () => {
    const { fetch, requests } = fakeFetch({ directUploadThrows: true });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);
    const file = new File(['fake image'], 'photo.jpg', { type: 'image/jpeg' });

    await expect(
      repository.uploadAssetAttachment('tenant-home', 'inventory-household', 'asset-passport', {
        id: 'photo-one',
        name: 'photo.jpg',
        sizeBytes: file.size,
        contentType: 'image/jpeg',
        previewUrl: 'blob:photo-one',
        file
      })
    ).rejects.toThrow('Direct upload to media storage failed.');
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'PUT https://uploads.local/object-one'
    ]);
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
      tagIds: ['tag-workshop', 'tag-travel'],
      lifecycleState: 'archived',
      mode: 'exact',
      checkoutState: 'checked_out'
    });

    expect(results).toMatchObject([
      {
        asset: {
          id: 'asset-passport',
          lifecycleState: 'archived',
          tags: [{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }]
        }
      }
    ]);
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/search/assets?q=Passport&limit=20&inventoryId=inventory-household&tagIds=tag-workshop&tagIds=tag-travel&lifecycleState=archived&mode=exact&checkoutState=checked_out'
    ]);
  });

  it('preserves Browse pagination metadata and API-backed default ordering', async () => {
    const { fetch, requests } = fakeFetch();
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);

    const page = await repository.browseAssets({
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: '', tagIds: [], lifecycleState: 'all',
      checkoutState: 'any', scope: 'all', sort: 'id_asc', mode: 'fuzzy', limit: 20
    });

    expect(page.assets[0]?.id).toBe('asset-archived');
    expect(page).toMatchObject({ hasMore: false, nextCursor: null });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets?limit=20&lifecycleState=all'
    ]);
    expect(observer.events.map((event) => event.eventName)).toEqual(['workspace.browse_started', 'workspace.browse_completed']);
  });

  it('checks Browse inventory existence across all lifecycle states', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(repository.hasAnyAssets('tenant-home', 'inventory-household')).resolves.toBe(true);
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets?limit=1&lifecycleState=all'
    ]);
  });

});

function photographedAsset(primaryPhotoId: string | undefined): Asset {
  return {
    id: 'asset-passport',
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'item',
    title: 'Passport',
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    primaryPhotoId
  };
}

function apiTenant(id: string, name: string) {
  return { id, name, access: { relationship: 'owner', permissions: ['view', 'create_inventory', 'configure'] } };
}

function pagedEnvelope(data: unknown[], nextCursor: string | null): Response {
  return Response.json({
    data,
    meta: { pagination: { limit: 50, nextCursor, hasMore: nextCursor !== null } }
  });
}
