import type {
  Asset
} from '@stuff-stash/api-client';
import { expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

it('pages active inventory map assets to completion instead of using only the recent summary page', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = Array.from({ length: 102 }, (_, index): Asset => ({
      id: index === 101 ? 'asset-final-child' : `asset-map-${index.toString().padStart(3, '0')}`,
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      kind: index === 0 ? 'location' : 'item',
      title: index === 101 ? 'Final child' : `Map asset ${index.toString().padStart(3, '0')}`,
      description: '',
      parentAssetId: index === 101 ? 'asset-map-000' : null,
      lifecycleState: index === 50 ? 'archived' : 'active',
      customFields: {},
      createdAt: '2026-06-20T10:00:00Z',
      updatedAt: `2026-06-20T10:${index.toString().padStart(2, '0')}:00Z`,
      ...(index > 0 ? {
        primaryPhoto: {
          id: `attachment-map-${index.toString()}`,
          fileName: `map-${index.toString()}.jpg`,
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: {
            small: 'small',
            medium: 'medium',
            large: 'large'
          }
        }
      } : {})
    }));
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home', undefined, 'scope-map-test');

    const mapAssets = await repository.listActiveInventoryMapAssets();

    expect(mapAssets).toMatchObject({
      sessionScopeId: 'scope-map-test',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      inventoryName: 'Home Inventory'
    });
    expect(mapAssets.assets).toHaveLength(101);
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-map-000')).toMatchObject({
      id: 'asset-map-000',
      title: 'Map asset 000'
    });
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-final-child')).toMatchObject({
      id: 'asset-final-child',
      title: 'Final child',
      parentAssetId: 'asset-map-000',
      locationTrail: ['Home Inventory', 'Map asset 000', 'Final child']
    });
    expect(client.thumbnailRequests).toHaveLength(100);
    expect(new Set(client.thumbnailRequests.map((request) => request.variant))).toEqual(new Set(['small']));
    const mapAssetRequests = client.listAssetRequests.filter((request) => request.lifecycleState === 'active');
    expect(mapAssetRequests).toEqual([
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: undefined,
        lifecycleState: 'active',
        sort: 'id_asc'
      },
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: '100',
        lifecycleState: 'active',
        sort: 'id_asc'
      }
    ]);
  });

it('loads one active location workspace traversal and resolves thumbnails only for its subtree', async () => {
    const client = new FakeInventoryApiClient();
    const photo = (suffix: string) => ({
      id: `attachment-${suffix}`,
      fileName: `${suffix}.jpg`,
      contentType: 'image/jpeg',
      sizeBytes: 1024,
      thumbnails: { small: 'small', medium: 'medium', large: 'large' }
    });
    client.assets = [
      {
        id: 'asset-garage', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'location',
        title: 'Garage', description: '', parentAssetId: null, lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('garage')
      },
      {
        id: 'asset-cabinet', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'container',
        title: 'Cabinet', description: '', parentAssetId: 'asset-garage', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('cabinet')
      },
      {
        id: 'asset-drawer', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'container',
        title: 'Drawer', description: '', parentAssetId: 'asset-cabinet', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('drawer')
      },
      {
        id: 'asset-hammer', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'item',
        title: 'Hammer', description: '', parentAssetId: 'asset-drawer', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('hammer')
      },
      {
        id: 'asset-shelf', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'location',
        title: 'Shelf', description: '', parentAssetId: 'asset-cabinet', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('shelf')
      },
      {
        id: 'asset-screws', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'item',
        title: 'Screws', description: '', parentAssetId: 'asset-shelf', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('screws')
      },
      {
        id: 'asset-kitchen', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'location',
        title: 'Kitchen', description: '', parentAssetId: null, lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('kitchen')
      },
      {
        id: 'asset-mug', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'item',
        title: 'Mug', description: '', parentAssetId: 'asset-kitchen', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('mug')
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const workspace = await repository.getAssetDetailWorkspace(assetId('asset-garage'));

    expect(workspace).toMatchObject({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      asset: { id: 'asset-garage' }
    });
    expect(workspace?.allAssets.map((asset) => asset.id)).toEqual([
      'asset-garage',
      'asset-cabinet',
      'asset-drawer',
      'asset-hammer',
      'asset-shelf',
      'asset-screws'
    ]);
    expect(workspace?.allAssets.find((asset) => asset.id === 'asset-drawer')).toMatchObject({
      hasPhoto: false,
      photo: undefined
    });
    expect(workspace?.allAssets.find((asset) => asset.id === 'asset-shelf')).toMatchObject({
      hasPhoto: false,
      photo: undefined
    });
    expect(workspace?.allAssets.find((asset) => asset.id === 'asset-hammer')).toMatchObject({
      locationTrail: ['Home Inventory', 'Garage', 'Cabinet', 'Drawer', 'Hammer'],
      hasPhoto: true
    });
    expect(client.thumbnailRequests.map((request) => request.assetId).sort()).toEqual([
      'asset-cabinet',
      'asset-garage',
      'asset-hammer',
      'asset-screws'
    ]);
    expect(client.thumbnailRequests.every((request) => request.variant === 'small')).toBe(true);
    expect(client.listAssetRequests).toEqual([{
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined,
      lifecycleState: 'active',
      sort: 'id_asc'
    }]);
    expect(client.getAssetRequests).toEqual([{
      inventoryId: 'inventory-home',
      assetId: 'asset-garage'
    }]);
  });

it('loads ordinary item detail context without paginating the active inventory', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const workspace = await repository.getAssetDetailWorkspace(assetId('asset-filters'));

    expect(workspace).toMatchObject({
      asset: {
        id: 'asset-filters',
        kind: 'item',
        locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
        parentLocationTrail: [{ id: 'asset-garage', title: 'Garage' }]
      },
      allAssets: []
    });
    expect(client.getAssetRequests).toEqual([
      { inventoryId: 'inventory-home', assetId: 'asset-filters' },
      { inventoryId: 'inventory-home', assetId: 'asset-garage' }
    ]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.thumbnailRequests.map((request) => request.assetId)).toEqual(['asset-filters']);
  });

it('loads an active container subtree so immediate children remain available', async () => {
    const client = new FakeInventoryApiClient();
    const photo = (suffix: string) => ({
      id: `attachment-${suffix}`,
      fileName: `${suffix}.jpg`,
      contentType: 'image/jpeg',
      sizeBytes: 1024,
      thumbnails: { small: 'small', medium: 'medium', large: 'large' }
    });
    client.assets = [
      {
        id: 'asset-cabinet', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'container',
        title: 'Cabinet', description: '', parentAssetId: null, lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('cabinet')
      },
      {
        id: 'asset-drawer', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'container',
        title: 'Drawer', description: '', parentAssetId: 'asset-cabinet', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('drawer')
      },
      {
        id: 'asset-hammer', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'item',
        title: 'Hammer', description: '', parentAssetId: 'asset-drawer', lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z', primaryPhoto: photo('hammer')
      },
      {
        id: 'asset-unrelated', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'item',
        title: 'Unrelated', description: '', parentAssetId: null, lifecycleState: 'active', customFields: {},
        createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const workspace = await repository.getAssetDetailWorkspace(assetId('asset-cabinet'));

    expect(workspace?.allAssets.map((asset) => asset.id)).toEqual(['asset-cabinet', 'asset-drawer']);
    expect(client.thumbnailRequests.map((request) => request.assetId).sort()).toEqual([
      'asset-cabinet',
      'asset-drawer'
    ]);
    expect(client.listAssetRequests).toHaveLength(1);
    expect(client.getAssetRequests).toEqual([{
      inventoryId: 'inventory-home',
      assetId: 'asset-cabinet'
    }]);
  });

it('loads an archived container target without traversing the active inventory', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [{
      id: 'asset-archive-box', tenantId: 'tenant-home', inventoryId: 'inventory-home', kind: 'container',
      title: 'Archive box', description: '', parentAssetId: null, lifecycleState: 'archived', customFields: {},
      createdAt: '2026-06-20T10:00:00Z', updatedAt: '2026-06-20T10:00:00Z'
    }];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const workspace = await repository.getAssetDetailWorkspace(assetId('asset-archive-box'));

    expect(workspace).toMatchObject({
      asset: { id: 'asset-archive-box', lifecycleState: 'archived' },
      allAssets: []
    });
    expect(client.listAssetRequests).toEqual([]);
    expect(client.getAssetRequests).toEqual([{
      inventoryId: 'inventory-home',
      assetId: 'asset-archive-box'
    }]);
  });

it('keeps inventory map structure available when one row thumbnail fails', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-garage',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'location',
        title: 'Garage',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z',
        primaryPhoto: {
          id: 'attachment-garage',
          fileName: 'garage.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: { small: 'small', medium: 'medium', large: 'large' }
        }
      },
      {
        id: 'asset-bin',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Camping bin',
        description: '',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z',
        primaryPhoto: {
          id: 'attachment-bin',
          fileName: 'bin.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: { small: 'small', medium: 'medium', large: 'large' }
        }
      }
    ];
    client.failedThumbnailAssetIds.add('asset-garage');
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    const mapAssets = await repository.listActiveInventoryMapAssets();

    expect(mapAssets.assets).toHaveLength(2);
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-garage')).toMatchObject({
      id: 'asset-garage',
      hasPhoto: false,
      photo: undefined
    });
    expect(mapAssets.assets.find((asset) => asset.id === 'asset-bin')).toMatchObject({
      id: 'asset-bin',
      hasPhoto: true,
      photo: {
        uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-bin/attachments/attachment-bin/thumbnail?variant=small'
      }
    });
  });
