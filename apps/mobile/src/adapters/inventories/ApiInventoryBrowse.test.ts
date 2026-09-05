import { expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

it('lists paged selected-inventory assets for browse mode', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: '',
        cursor: undefined,
        limit: 1,
        lifecycleState: 'active',
        checkoutState: 'any',
        kind: 'item',
        sort: 'updated_desc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        }
      ],
      hasMore: true
    });
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 1,
      cursor: undefined,
      lifecycleState: 'active',
      sort: 'updated_desc'
    });
  });

it('continues list pagination until a kind-filtered browse page has matches', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: '',
        limit: 1,
        lifecycleState: 'active',
        checkoutState: 'any',
        kind: 'item',
        sort: 'id_asc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        }
      ],
      hasMore: false
    });
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 1,
      cursor: undefined,
      lifecycleState: 'active',
      sort: 'id_asc'
    });
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 1,
      cursor: '1',
      lifecycleState: 'active',
      sort: 'id_asc'
    });
  });

it('uses the checked-out inventory endpoint for checked-out browse mode', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        ...client.assets[1]!,
        currentCheckout: {
          id: 'checkout-filters',
          state: 'open',
          checkedOutAt: '2026-06-25T12:00:00Z',
          checkedOutByPrincipalId: 'principal-home'
        }
      },
      client.assets[0]!
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: '',
        limit: 10,
        lifecycleState: 'active',
        checkoutState: 'checked_out',
        kind: 'item',
        sort: 'updated_desc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          currentCheckout: {
            id: 'checkout-filters',
            state: 'open',
            checkedOutAt: '2026-06-25T12:00:00Z',
            checkedOutByPrincipalId: 'principal-home'
          }
        }
      ],
      hasMore: false
    });
    expect(client.listCheckedOutAssetRequests).toEqual([
      {
        inventoryId: 'inventory-home',
        limit: 10,
        cursor: undefined
      }
    ]);
  });

it('searches paged selected-inventory assets with lifecycle filtering', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(
      repository.browseAssets({
        query: 'filters',
        cursor: 'next-page',
        limit: 10,
        lifecycleState: 'all',
        checkoutState: 'any',
        kind: 'item',
        sort: 'updated_desc'
      })
    ).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        }
      ]
    });
    expect(client.searchedQuery).toBe('tenant-home:filters');
  });

it('searches selected-inventory assets with multi-tag filters without replacing the query', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.browseAssets({
      query: '',
      cursor: undefined,
      limit: 10,
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'item',
      sort: 'updated_desc',
      tagIds: ['tag-workshop', 'tag-camping']
    });

    expect(client.listAssetRequests).toEqual([]);
    expect(client.searchAssetRequests[0]).toMatchObject({
      tenantId: 'tenant-home',
      query: '',
      inventoryId: 'inventory-home',
      tagIds: ['tag-workshop', 'tag-camping'],
      lifecycleState: 'active',
      checkoutState: 'any'
    });
  });

it('maps tag-backed search matches to a user-facing mobile label', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: 'tagged',
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc',
      limit: 20
    })).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-filters'
        }
      ],
      searchMatches: [
        {
          assetId: 'asset-filters',
          labels: ['Tag']
        }
      ]
    });
  });

it('maps nested parent trails for asset workspace paths', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      ...client.assets,
      {
        id: 'asset-cabinet',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Big cabinet',
        description: '',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-24T10:00:00Z',
        updatedAt: '2026-06-24T10:00:00Z'
      },
      {
        id: 'asset-shelf',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Second shelf',
        description: '',
        parentAssetId: 'asset-cabinet',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T10:00:00Z',
        updatedAt: '2026-06-25T10:00:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: '',
      limit: 10,
      lifecycleState: 'all',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: expect.arrayContaining([
        expect.objectContaining({
          id: 'asset-shelf',
          locationTrail: ['Home Inventory', 'Garage', 'Big cabinet', 'Second shelf']
        })
      ])
    });
  });

it('deduplicates shared browse-page ancestors without scanning unrelated assets', async () => {
    const client = new FakeInventoryApiClient();
    client.assets.push({ ...client.assets[1]!, id: 'second-filter', title: 'Spare filters' });
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    const page = await repository.browseAssets({ query: '', limit: 2, lifecycleState: 'active', checkoutState: 'any', kind: 'all', sort: 'updated_desc' });
    expect(page.assets).toHaveLength(2);
    expect(client.listAssetRequests).toEqual([expect.objectContaining({ limit: 2, sort: 'updated_desc' })]);
    expect(client.getAssetRequests).toEqual([{ inventoryId: 'inventory-home', assetId: 'asset-garage' }]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads only missing page ancestors for browse result cards', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-seasonal-bin',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Holiday / seasonal bin',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T09:00:00Z',
        updatedAt: '2026-06-20T09:00:00Z'
      },
      {
        id: 'asset-hand-towels',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Christmas hand towels',
        description: '',
        parentAssetId: 'asset-seasonal-bin',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T09:59:00Z',
        updatedAt: '2026-06-25T09:59:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: '',
      limit: 1,
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: [
        expect.objectContaining({
          id: 'asset-hand-towels',
          locationTrail: ['Home Inventory', 'Holiday / seasonal bin', 'Christmas hand towels'],
          parentLocationTrail: [{ id: assetId('asset-seasonal-bin'), title: 'Holiday / seasonal bin' }]
        })
      ]
    });
    expect(client.listAssetRequests.filter((request) =>
      request.inventoryId === 'inventory-home' &&
      request.lifecycleState === 'active' &&
      request.sort === 'id_asc'
    )).toHaveLength(0);
  });

it('loads only missing page ancestors for search result cards', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-seasonal-bin',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Holiday / seasonal bin',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T09:00:00Z',
        updatedAt: '2026-06-20T09:00:00Z'
      },
      {
        id: 'asset-hand-towels',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Christmas hand towels',
        description: '',
        parentAssetId: 'asset-seasonal-bin',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T09:59:00Z',
        updatedAt: '2026-06-25T09:59:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: 'christmas',
      limit: 1,
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: [
        expect.objectContaining({
          id: 'asset-hand-towels',
          locationTrail: ['Home Inventory', 'Holiday / seasonal bin', 'Christmas hand towels'],
          parentLocationTrail: [{ id: assetId('asset-seasonal-bin'), title: 'Holiday / seasonal bin' }]
        })
      ]
    });
    expect(client.listAssetRequests.filter((request) =>
      request.inventoryId === 'inventory-home' &&
      request.lifecycleState === 'active' &&
      request.sort === 'id_asc'
    )).toHaveLength(0);
  });

it('honors an authoritative root search result without borrowing old tree linkage', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-seasonal-bin',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'container',
        title: 'Holiday / seasonal bin',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T09:00:00Z',
        updatedAt: '2026-06-20T09:00:00Z'
      },
      {
        id: 'asset-hand-towels',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Christmas hand towels',
        description: '',
        parentAssetId: 'asset-seasonal-bin',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T09:59:00Z',
        updatedAt: '2026-06-25T09:59:00Z'
      }
    ];
    client.searchResultAssetOverrides.set('asset-hand-towels', { parentAssetId: null });
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: 'christmas',
      limit: 1,
      lifecycleState: 'active',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: [
        expect.objectContaining({
          id: 'asset-hand-towels',
          locationTrail: ['Home Inventory', 'Christmas hand towels'],
          parentLocationTrail: []
        })
      ]
    });
  });

it('uses primary photo summaries for browse cards without paged attachment lookups', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-many-photos',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Photo album',
        description: '',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T10:00:00Z',
        updatedAt: '2026-06-25T10:00:00Z',
        primaryPhoto: {
          id: 'attachment-many-one',
          fileName: 'many-one.jpg',
          contentType: 'image/jpeg',
          sizeBytes: 1024,
          thumbnails: {
            small: '/tenants/tenant-home/inventories/inventory-home/assets/asset-many-photos/attachments/attachment-many-one/thumbnail?variant=small',
            medium: '/tenants/tenant-home/inventories/inventory-home/assets/asset-many-photos/attachments/attachment-many-one/thumbnail?variant=medium',
            large: '/tenants/tenant-home/inventories/inventory-home/assets/asset-many-photos/attachments/attachment-many-one/thumbnail?variant=large'
          }
        }
      },
      ...client.assets
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.browseAssets({
      query: '',
      limit: 10,
      lifecycleState: 'all',
      checkoutState: 'any',
      kind: 'all',
      sort: 'updated_desc'
    })).resolves.toMatchObject({
      assets: expect.arrayContaining([
        expect.objectContaining({
          id: 'asset-many-photos',
          photos: [
            expect.objectContaining({ id: 'attachment-many-one', fileName: 'many-one.jpg' })
          ]
        })
      ])
    });
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads bounded parent suggestions without loading attachments or the inventory tree', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    await repository.listParentCandidates('');
    expect(client.listAssetRequests).toEqual([expect.objectContaining({ limit: 5, sort: 'updated_desc' })]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('continues paged tenant search until selected-inventory asset results are found', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.searchAssets('paged')).resolves.toMatchObject([
      {
        id: 'asset-filters',
        title: 'Furnace filters'
      }
    ]);
  });

it('does not stop selected-inventory search at five tenant search pages', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.searchAssets('sixth-page')).resolves.toMatchObject([
      {
        id: 'asset-filters',
        title: 'Furnace filters'
      }
    ]);
  });

it('bounds sparse kind-filter scans and preserves the continuation cursor', async () => {
  const client = new FakeInventoryApiClient();
  const location = client.assets[0];
  client.assets = Array.from({ length: 20 }, (_, index) => ({ ...location, id: `location-${index}` }));
  const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
  const page = await repository.browseAssets({ query: '', limit: 1, lifecycleState: 'active', checkoutState: 'any', kind: 'item', sort: 'updated_desc' });
  expect(client.listAssetRequests).toHaveLength(5);
  expect(page.assets).toEqual([]); expect(page.hasMore).toBe(true); expect(page.nextCursor).toBeDefined();
});
