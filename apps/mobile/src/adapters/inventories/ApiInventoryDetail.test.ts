import { expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId, tenantId } from '../../domain/inventories/InventorySummary';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

it('loads the complete active image attachment set for asset detail', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getAssetDetail({
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      asset: {
        id: assetId('asset-filters'),
        title: 'Furnace filters',
        kind: 'item',
        lifecycleState: 'active',
        parentAssetId: assetId('asset-garage'),
        locationLabel: 'Garage',
        locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
        parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
        description: 'Three-pack of filters.',
        updatedAtLabel: 'Updated today',
        hasPhoto: true
      }
    })).resolves.toMatchObject({
      id: 'asset-filters',
      title: 'Furnace filters',
      hasPhoto: true,
      photos: [
        {
          id: 'attachment-filters-photo',
          fileName: 'filters.jpg',
          uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small',
          heroUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=medium',
          viewerUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=large'
        },
        {
          id: 'attachment-filters-label',
          fileName: 'filters-label.jpg'
        }
      ]
    });
    expect(client.listAttachmentRequests).toEqual([
      { assetId: 'asset-filters', limit: 50, cursor: undefined }
    ]);
    const detailThumbnailRequests = client.thumbnailRequests.slice(-6);
    expect(detailThumbnailRequests).toHaveLength(6);
    expect(detailThumbnailRequests).toEqual(expect.arrayContaining([
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-photo', variant: 'small' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-photo', variant: 'medium' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-photo', variant: 'large' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-label', variant: 'small' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-label', variant: 'medium' },
      { assetId: 'asset-filters', attachmentId: 'attachment-filters-label', variant: 'large' }
    ]));
  });

it('does not collapse asset detail attachment lookup failures into an empty photo set', async () => {
    const client = new FakeInventoryApiClient();
    client.shouldFailAttachmentLookup = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getAssetDetail({
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      asset: {
        id: assetId('asset-filters'),
        title: 'Furnace filters',
        kind: 'item',
        lifecycleState: 'active',
        locationLabel: 'Garage',
        locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
        parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
        description: 'Three-pack of filters.',
        updatedAtLabel: 'Updated today',
        hasPhoto: true
      }
    })).rejects.toThrow('Asset attachments could not be loaded.');
  });

it('loads container placement without traversing its contents', async () => {
    const client = new FakeInventoryApiClient();
    client.assets[1] = { ...client.assets[1]!, kind: 'container' };
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    const core = await repository.getAssetCore(assetId('asset-filters'));
    client.getAssetRequests.length = 0;
    await expect(repository.getAssetPlacement(core)).resolves.toMatchObject({ parentLocationTrail: [{ id: 'asset-garage', title: 'Garage' }] });
    expect(client.getAssetRequests).toEqual([{ inventoryId: 'inventory-home', assetId: 'asset-garage' }]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads asset core through one detail request without attachment or containment hydration', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    await repository.getCurrentInventoryScope();
    client.getAssetRequests.length = 0;

    await expect(repository.getAssetCore(assetId('asset-filters'))).resolves.toMatchObject({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      asset: { id: 'asset-filters', title: 'Furnace filters', photos: [] }
    });

    expect(client.getAssetRequests).toEqual([{
      inventoryId: 'inventory-home',
      assetId: 'asset-filters'
    }]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
    expect(client.thumbnailRequests).toEqual([]);
  });

it('loads placement independently without repeating the selected asset or loading photos', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    const core = await repository.getAssetCore(assetId('asset-filters'));
    client.getAssetRequests.length = 0;

    await expect(repository.getAssetContents(core)).resolves.toMatchObject({
      asset: {
        id: 'asset-filters',
        parentLocationTrail: [{ id: 'asset-garage', title: 'Garage' }]
      },
      allAssets: []
    });

    expect(client.getAssetRequests).toEqual([{
      inventoryId: 'inventory-home',
      assetId: 'asset-garage'
    }]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
    expect(client.thumbnailRequests).toEqual([]);
  });

it('loads photos independently without repeating asset or containment requests', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    const core = await repository.getAssetCore(assetId('asset-filters'));
    client.getAssetRequests.length = 0;

    await repository.getAssetPhotos(core);

    expect(client.getAssetRequests).toEqual([]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([{
      assetId: 'asset-filters',
      limit: 50,
      cursor: undefined
    }]);
  });

it('preserves contained child photo references without attachment-list hydration', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    const core = await repository.getAssetCore(assetId('asset-garage'));
    const contents = await repository.getAssetContents(core);
    expect(contents.allAssets.find((asset) => asset.id === 'asset-filters')).toMatchObject({
      hasPhoto: true,
      photo: { uri: expect.any(String) }
    });
    expect(client.listAttachmentRequests).toEqual([]);
    expect(client.getAssetRequests.filter((request) => request.assetId === 'asset-garage')).toHaveLength(1);
  });

it('uses one selected-asset request across the complete progressive detail graph', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const core = await repository.getAssetCore(assetId('asset-filters'));
    await Promise.all([
      repository.getAssetContents(core),
      repository.getAssetPhotos(core)
    ]);

    expect(client.getAssetRequests.filter((request) => request.assetId === 'asset-filters'))
      .toHaveLength(1);
    expect(client.listAttachmentRequests).toHaveLength(1);
    expect(client.listAssetRequests).toEqual([]);
  });
