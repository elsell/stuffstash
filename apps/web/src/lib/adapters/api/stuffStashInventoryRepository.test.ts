import { beforeEach, describe, expect, it } from 'vitest';
import { StuffStashInventoryRepository } from './stuffStashInventoryRepository';
import { AuthenticationRequiredError } from '$lib/application/authenticationRequired';
import { InMemoryWorkspaceObserver } from '$lib/observability/workspaceObserver';
import { config, fakeFetch } from './stuffStashInventoryRepository.test-helpers';

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
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/custom-asset-types?limit=100',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/custom-field-definitions?limit=100',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/tags?limit=100',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/assets?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-cabin/inventories/inventory-cabin/checked-out-assets?limit=50'
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

  it('hydrates API primary photos into workspace asset photos', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch, requests } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'] });
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();

    expect(data.assets[0]?.photo).toMatchObject({
      id: 'attachment-one',
      assetId: 'asset-archived',
      url: expect.stringContaining('blob:'),
      alt: 'Archived Passport'
    });
    expect(data.context.assetTags).toEqual([{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }]);
    expect(data.assets[0]?.tags).toEqual([{ id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }]);
    const thumbnailRequest = requests.find((request) =>
      request.url.includes('/assets/asset-archived/attachments/attachment-one/thumbnail?variant=small')
    );
    expect(thumbnailRequest?.headers.get('Authorization')).toBe('Bearer id-token');
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
      photo: expect.objectContaining({ assetId: 'asset-archived' })
    });
    expect(data.assets.find((asset) => asset.id === 'asset-container')).toMatchObject({ id: 'asset-container', kind: 'container' });
    expect(data.assets.find((asset) => asset.id === 'asset-container')).not.toHaveProperty('photo');
    expect(data.assets.find((asset) => asset.id === 'asset-location')).toMatchObject({ id: 'asset-location', kind: 'location' });
    expect(data.assets.find((asset) => asset.id === 'asset-location')).not.toHaveProperty('photo');
    expect(requests.filter((request) => request.url.includes('/thumbnail'))).toHaveLength(1);
  });

  it('keeps workspace assets when a primary photo thumbnail cannot be fetched', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'], rejectedThumbnailAssetIds: ['asset-archived'] });
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);

    const data = await repository.loadWorkspace();

    expect(data.assets[0]?.id).toBe('asset-archived');
    expect(data.assets[0]?.photo).toBeUndefined();
    expect(data.assets[0]?.photoUnavailable).toBe(true);
    expect(observer.events).toContainEqual({
      eventName: 'workspace.asset_primary_photo_load_failed',
      attributes: { assetId: 'asset-archived' }
    });
  });

  it('marks workspace photos unavailable when a primary photo thumbnail returns an HTTP error', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch } = fakeFetch({ primaryPhotoAssetIds: ['asset-archived'], failedThumbnailStatusByAssetId: { 'asset-archived': 500 } });
    const observer = new InMemoryWorkspaceObserver();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', observer, fetch);

    const data = await repository.loadWorkspace();

    expect(data.assets[0]?.id).toBe('asset-archived');
    expect(data.assets[0]?.photo).toBeUndefined();
    expect(data.assets[0]?.photoUnavailable).toBe(true);
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
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/custom-asset-types?limit=100',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/custom-field-definitions?limit=100',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/tags?limit=100',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/assets?limit=100&lifecycleState=active',
      'GET http://api.local/tenants/tenant-empty/inventories/inventory-created/checked-out-assets?limit=50'
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

  it('does not reuse search result photos across items, containers, or locations', async () => {
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
      photo: expect.objectContaining({ assetId: 'asset-passport' })
    });
    expect(results[1]?.asset).toMatchObject({ id: 'asset-container', kind: 'container' });
    expect(results[1]?.asset).not.toHaveProperty('photo');
    expect(results[2]?.asset).toMatchObject({ id: 'asset-location', kind: 'location' });
    expect(results[2]?.asset).not.toHaveProperty('photo');
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
      parentAssetId: 'asset-safe'
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
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/custom-asset-types?limit=100',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/custom-field-definitions?limit=100',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/tags?limit=100',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets?limit=100&lifecycleState=archived',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/checked-out-assets?limit=50'
    ]);
  });

  it('loads checked-out assets regardless of lifecycle during workspace load', async () => {
    sessionStorage.setItem('stuffstash.selectedTenantId', 'tenant-home');
    sessionStorage.setItem('stuffstash.selectedInventoryId', 'inventory-household');
    const { fetch } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    const data = await repository.loadWorkspace();

    expect(data.checkedOutAssets).toMatchObject([
      {
        asset: { id: 'asset-archived', lifecycleState: 'archived' },
        checkout: { id: 'checkout-open', state: 'open' }
      }
    ]);
  });

  it('checks out, returns, and lists checkout history through generated client paths', async () => {
    const { fetch, requests } = fakeFetch();
    const repository = new StuffStashInventoryRepository(config, () => 'id-token', new InMemoryWorkspaceObserver(), fetch);

    await expect(
      repository.checkoutAsset('tenant-home', 'inventory-household', 'asset-passport', { details: 'using at desk' })
    ).resolves.toMatchObject({
      id: 'checkout-open',
      checkoutDetails: 'using at desk',
      state: 'open'
    });
    await expect(repository.returnAsset('tenant-home', 'inventory-household', 'asset-passport', { details: 'back in bin' })).resolves.toMatchObject({
      id: 'checkout-open',
      state: 'returned',
      returnDetails: 'back in bin'
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

  it('falls back to the JSON attachment upload route when a browser direct upload target rejects the upload', async () => {
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
    ).resolves.toMatchObject({ id: 'attachment-one', fileName: 'photo.jpg' });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'PUT https://uploads.local/object-one',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail?variant=small'
    ]);
    expect(await requests[2]?.json()).toEqual({
      fileName: 'photo.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZSBpbWFnZQ=='
    });
  });

  it('falls back to the JSON attachment upload route when browser direct upload fetch fails', async () => {
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
    ).resolves.toMatchObject({ id: 'attachment-one', fileName: 'photo.jpg' });
    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/direct-uploads',
      'PUT https://uploads.local/object-one',
      'POST http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments',
      'GET http://api.local/tenants/tenant-home/inventories/inventory-household/assets/asset-passport/attachments/attachment-one/thumbnail?variant=small'
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
      'GET http://api.local/tenants/tenant-home/search/assets?q=Passport&limit=20&inventoryId=inventory-household&lifecycleState=archived&mode=exact&checkoutState=checked_out'
    ]);
  });

});
