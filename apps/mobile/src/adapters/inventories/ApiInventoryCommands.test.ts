import { expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { createMobileQueryClient, mobileQueryKeys } from '../serverState/MobileQueryClient';
import { QueryClientInventoryMutationObserver } from '../serverState/QueryClientInventoryMutationObserver';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

it('creates and updates without scanning inventory or loading response attachments', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    await repository.createAsset({ kind: 'item', title: 'Root item', description: '' });
    expect(client.getAssetRequests).toEqual([]);
    await repository.updateAsset({ assetId: assetId('asset-filters'), title: 'Updated' });
    expect(client.getAssetRequests).toEqual([
      { inventoryId: 'inventory-home', assetId: 'asset-filters' },
      { inventoryId: 'inventory-home', assetId: 'asset-garage' }
    ]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('creates and searches assets through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.createAsset({
        kind: 'item',
        title: 'USB-C charger pouch',
        description: 'Chargers and spare cables.',
        parentAssetId: assetId('asset-garage'),
        tagIds: ['tag-workshop']
      })
    ).resolves.toMatchObject({
      id: 'asset-created',
      title: 'USB-C charger pouch',
      locationLabel: 'Garage'
    });
    expect(client.createdAssetInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      title: 'USB-C charger pouch',
      parentAssetId: 'asset-garage',
      tagIds: ['tag-workshop']
    });

    await repository.addAssetPhoto(assetId('asset-created'), {
      fileName: 'created.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZQ=='
    });
    expect(client.createdAttachmentInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-created',
      fileName: 'created.jpg'
    });

    await expect(repository.searchAssets('filters')).resolves.toMatchObject([
      {
        id: 'asset-filters',
        title: 'Furnace filters',
        tags: [{ id: 'tag-workshop', displayName: 'Workshop', color: '#2F80ED' }]
      }
    ]);
    await expect(repository.searchAssets('filters')).resolves.toHaveLength(1);
    expect(client.searchedQuery).toBe('tenant-home:filters');
  });

it('creates asset tags through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.createAssetTag({
      displayName: 'Camping',
      color: '#2F80ED'
    })).resolves.toEqual({
      id: 'tag-created',
      key: 'camping',
      displayName: 'Camping',
      color: '#2F80ED'
    });

    expect(client.createdAssetTagInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      displayName: 'Camping',
      color: '#2F80ED'
    });
  });

it('updates asset fields and parent placement through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.updateAsset({
      assetId: assetId('asset-filters'),
      title: 'HEPA filters',
      description: 'Replacement filters.',
      parentAssetId: null,
      tagIds: ['tag-workshop']
    })).resolves.toMatchObject({
      id: 'asset-filters',
      title: 'HEPA filters',
      parentAssetId: undefined,
      locationLabel: 'Inventory root'
    });

    expect(client.updatedAssetInput).toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-filters',
      title: 'HEPA filters',
      description: 'Replacement filters.',
      parentAssetId: null,
      tagIds: ['tag-workshop']
    });
  });

it('updates asset lifecycle through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.archiveAsset(assetId('asset-filters'));
    await repository.restoreAsset(assetId('asset-filters'));
    await repository.deleteAsset(assetId('asset-filters'));

    expect(client.lifecycleInputs).toEqual([
      {
        action: 'archive',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters'
      },
      {
        action: 'restore',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters'
      },
      {
        action: 'delete',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters'
      }
    ]);
  });

it('checks out and returns assets through the generated client wrapper', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.checkoutAsset(assetId('asset-filters'), { details: 'using at desk' })).resolves.toEqual({
      id: 'checkout-fake',
      assetId: 'asset-filters',
      undoableOperationId: 'operation-checkout-fake'
    });
    await expect(repository.returnAsset(assetId('asset-filters'))).resolves.toEqual({
      id: 'checkout-fake',
      assetId: 'asset-filters',
      undoableOperationId: 'operation-checkout-fake'
    });
    await expect(repository.updateReturnedCheckoutDetails(assetId('asset-filters'), 'checkout-fake', { details: 'back in bin' })).resolves.toEqual({
      id: 'checkout-fake',
      assetId: 'asset-filters',
      undoableOperationId: 'operation-checkout-fake'
    });
    await repository.undoInventoryOperation('operation-return-fake');

    expect(client.checkoutInputs).toEqual([
      {
        action: 'checkout',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters',
        details: 'using at desk'
      },
      {
        action: 'return',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters',
        details: undefined
      },
      {
        action: 'return_details',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        assetId: 'asset-filters',
        checkoutId: 'checkout-fake',
        details: 'back in bin'
      }
    ]);
    expect(client.undoInputs).toEqual([
      {
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        operationId: 'operation-return-fake',
        direction: 'undo'
      }
    ]);
  });

it('returns an asset without reloading the selected inventory workspace', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.getCurrentTenantId();
    client.requestKinds.length = 0;
    client.listAssetRequests.length = 0;
    client.listAttachmentRequests.length = 0;
    client.listAssetTagRequests.length = 0;

    await repository.returnAsset(assetId('asset-filters'));

    expect(client.requestKinds).toEqual(['return_asset']);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
    expect(client.listAssetTagRequests).toEqual([]);
  });

it('emits scoped mutation impact after a successful return', async () => {
    const client = new FakeInventoryApiClient();
    const mutations: unknown[] = [];
    const repository = new ApiInventorySummaryRepository(
      client,
      'tenant-home',
      undefined,
      'scope-one',
      {},
      { onInventoryMutation: (mutation) => mutations.push(mutation) }
    );

    await repository.returnAsset(assetId('asset-filters'));

    expect(mutations).toEqual([{
      kind: 'asset_checkout_changed',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      assetId: 'asset-filters'
    }]);
  });

it('invalidates the promoted item parent core when creating its first child', async () => {
    const client = new FakeInventoryApiClient();
    const cache = createMobileQueryClient();
    const parentKey = mobileQueryKeys.assetCore('scope', 'tenant-home', 'inventory-home', 'asset-filters');
    cache.setQueryData(parentKey, { snapshot: { asset: { id: 'asset-filters', kind: 'item' } } });
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', undefined, 'scope', {}, new QueryClientInventoryMutationObserver(cache, 'scope'));
    await repository.createAsset({ kind: 'item', title: 'New child', description: '', parentAssetId: assetId('asset-filters') });
    expect(cache.getQueryState(parentKey)?.isInvalidated).toBe(true);
    cache.clear();
  });

it('invalidates ancestor contents when creating inside an unvisited empty nested container', async () => {
    const client = new FakeInventoryApiClient();
    const garage = client.assets[0]!;
    client.assets = [garage,
      { ...garage, id: 'outer', kind: 'container', parentAssetId: garage.id },
      { ...garage, id: 'inner', kind: 'container', parentAssetId: 'outer' }
    ];
    const cache = createMobileQueryClient();
    const key = mobileQueryKeys.assetContents('scope', 'tenant-home', 'inventory-home', garage.id, 'location:active:root');
    cache.setQueryData(key, { containedSpaces: [{ id: 'outer' }], containedItems: [] });
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', undefined, 'scope', {},
      new QueryClientInventoryMutationObserver(cache, 'scope'));
    await repository.createAsset({ kind: 'item', title: 'New item', description: '', parentAssetId: assetId('inner') });
    expect(cache.getQueryState(key)?.isInvalidated).toBe(true);
    cache.clear();
  });
