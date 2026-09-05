import type {
  Asset
} from '@stuff-stash/api-client';
import { expect, it } from 'vitest';
import { assetId } from '../../domain/assets/AssetSummary';
import { inventoryId } from '../../domain/inventories/InventorySummary';
import { ApiInventorySummaryRepository } from './ApiInventorySummaryRepository';
import { FakeInventoryApiClient } from './testing/InventoryApiClient';

it('maps generated API client responses into mobile inventory summaries', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getInventoryWorkspace()).resolves.toMatchObject({
      tenants: [
        { id: 'tenant-home', name: 'Home' },
        { id: 'tenant-cabin', name: 'Cabin' }
      ],
      defaultInventoryId: 'inventory-home',
      inventories: [
        {
          id: 'inventory-home',
          tenantId: 'tenant-home',
          name: 'Home Inventory',
          role: 'owner',
          locationCount: 1,
          assetTags: [
            {
              id: 'tag-workshop',
              key: 'workshop',
              displayName: 'Workshop',
              color: '#2F80ED'
            }
          ],
          locations: [
            {
              id: 'asset-garage',
              title: 'Garage',
              containedAssetCount: 1,
              recentAssetTitles: ['Furnace filters']
            }
          ],
          assets: [
            {
              id: 'asset-filters',
              locationLabel: 'Garage',
              locationTrail: ['Home Inventory', 'Garage', 'Furnace filters'],
              hasPhoto: true,
              photos: [
                {
                  id: 'attachment-filters-photo',
                  fileName: 'filters.jpg'
                }
              ],
              photo: {
                uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small',
                heroUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=medium',
                heroHeaders: { Authorization: 'Bearer dev-token' },
                viewerUri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=large',
                viewerHeaders: { Authorization: 'Bearer dev-token' }
              }
            },
            {
              id: 'asset-garage',
              locationLabel: 'Inventory root',
              locationTrail: ['Home Inventory', 'Garage']
            }
          ]
        },
        {
          id: 'inventory-cabin',
          tenantId: 'tenant-cabin',
          name: 'Cabin Inventory',
          role: 'viewer',
          locationCount: 0,
          locations: [],
          assets: []
        }
      ]
    });
    expect(client.listAssetTagRequests).toContainEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined
    });
  });

it('loads every active asset tag page for mobile selection', async () => {
    const client = new FakeInventoryApiClient();
    client.paginatedAssetTags = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      assetTags: [
        { id: 'tag-workshop', displayName: 'Workshop' },
        { id: 'tag-camping', displayName: 'Camping' }
      ]
    });
    expect(client.listAssetTagRequests).toContainEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined
    });
    expect(client.listAssetTagRequests).toContainEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: 'next-tags'
    });
  });

it('uses the selected inventory for later default inventory operations', async () => {
    const repository = new ApiInventorySummaryRepository(
      new FakeInventoryApiClient(),
      'tenant-home'
    );

    await expect(repository.getCurrentTenantId()).resolves.toBe('tenant-home');
    await repository.selectInventory(inventoryId('inventory-cabin'));
    await expect(repository.getCurrentTenantId()).resolves.toBe('tenant-cabin');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      id: 'inventory-cabin',
      tenantId: 'tenant-cabin',
      name: 'Cabin Inventory'
    });
  });

it('selects from the inventory directory without hydrating either inventory', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.selectInventory(inventoryId('inventory-cabin'));

    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAssetTagRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
    await expect(repository.getCurrentInventoryScope()).resolves.toEqual({
      tenantId: 'tenant-cabin',
      inventoryId: 'inventory-cabin'
    });
  });

it('refreshes a cached inventory directory before selecting a newly accepted inventory', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
    await repository.getCurrentInventoryScope();
    client.additionalHomeInventory = {
      ...client.inventory,
      id: 'inventory-new',
      name: 'Newly shared inventory'
    };

    await repository.selectInventory(inventoryId('inventory-new'));

    await expect(repository.getCurrentInventoryScope()).resolves.toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-new'
    });
    expect(client.listInventoryRequests.filter((id) => id === 'tenant-home')).toHaveLength(2);
    expect(client.listAssetRequests).toEqual([]);
  });

it('does not list attachments while loading dense inventory summaries', async () => {
    const client = new FakeInventoryApiClient();
    client.shouldFailAttachmentLookup = true;
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      id: 'inventory-home',
      assets: [
        { id: 'asset-filters', hasPhoto: true },
        { id: 'asset-garage', hasPhoto: false }
      ]
    });
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('requests API-owned updated-descending asset order for mobile recency', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.getDefaultInventorySummary();

    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 100,
      cursor: undefined,
      lifecycleState: 'all',
      sort: 'updated_desc'
    });
  });

it('preserves API-provided recency order across asset kinds', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      {
        id: 'asset-new-batteries',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Fresh batteries',
        description: 'Just created from the Add sheet.',
        parentAssetId: 'asset-garage',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-24T11:00:00Z',
        updatedAt: '2026-06-24T11:00:00Z'
      },
      ...client.assets
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      assets: [
        {
          id: 'asset-new-batteries',
          kind: 'item',
          title: 'Fresh batteries'
        },
        {
          id: 'asset-filters',
          kind: 'item',
          title: 'Furnace filters'
        },
        {
          id: 'asset-garage',
          kind: 'location',
          title: 'Garage'
        }
      ]
    });
  });

it('uses the full active tree to build parent trails for recent asset cards', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      ...Array.from({ length: 99 }, (_, index): Asset => ({
        id: `asset-new-root-${index.toString().padStart(3, '0')}`,
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: `New root ${index.toString().padStart(3, '0')}`,
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: `2026-06-25T10:${index.toString().padStart(2, '0')}:00Z`,
        updatedAt: `2026-06-25T10:${index.toString().padStart(2, '0')}:00Z`
      })),
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
      },
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
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      assets: expect.arrayContaining([
        expect.objectContaining({
          id: 'asset-hand-towels',
          locationTrail: ['Home Inventory', 'Holiday / seasonal bin', 'Christmas hand towels'],
          parentLocationTrail: [{ id: assetId('asset-seasonal-bin'), title: 'Holiday / seasonal bin' }]
        })
      ])
    });
  });

it('loads only required ancestors to build parent trails for checked-out asset cards', async () => {
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
        currentCheckout: {
          id: 'checkout-one',
          state: 'open',
          checkedOutAt: '2026-06-25T09:59:00Z',
          checkedOutByPrincipalId: 'user-one'
        },
        createdAt: '2026-06-25T09:59:00Z',
        updatedAt: '2026-06-25T09:59:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getHomeDashboardSnapshot()).resolves.toMatchObject({
      checkedOutAssets: [
        expect.objectContaining({
          id: 'asset-hand-towels',
          locationTrail: ['Home Inventory', 'Holiday / seasonal bin', 'Christmas hand towels'],
          parentLocationTrail: [{ id: assetId('asset-seasonal-bin'), title: 'Holiday / seasonal bin' }]
        })
      ]
    });
    expect(client.listAssetRequests.filter((request) =>
      request.lifecycleState === 'active' && request.sort === 'id_asc'
    )).toEqual([]);
  });

it('loads locations from the full active inventory tree instead of only the recent summary page', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      ...Array.from({ length: 100 }, (_, index): Asset => ({
        id: `asset-recent-item-${index.toString().padStart(3, '0')}`,
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: `Recent item ${index.toString().padStart(3, '0')}`,
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-25T10:00:00Z',
        updatedAt: `2026-06-25T${(10 + Math.floor(index / 60)).toString().padStart(2, '0')}:${(index % 60).toString().padStart(2, '0')}:00Z`
      })),
      {
        id: 'asset-late-location',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'location',
        title: 'Late page closet',
        description: '',
        parentAssetId: null,
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-20T10:00:00Z'
      },
      {
        id: 'asset-late-child',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Stored blanket',
        description: '',
        parentAssetId: 'asset-late-location',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:01:00Z',
        updatedAt: '2026-06-20T10:01:00Z'
      },
      {
        id: 'asset-late-newer-child',
        tenantId: 'tenant-home',
        inventoryId: 'inventory-home',
        kind: 'item',
        title: 'Newer stored blanket',
        description: '',
        parentAssetId: 'asset-late-location',
        lifecycleState: 'active',
        customFields: {},
        createdAt: '2026-06-20T10:02:00Z',
        updatedAt: '2026-06-20T10:02:00Z'
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getDefaultInventorySummary()).resolves.toMatchObject({
      locationCount: 1,
      locations: [
        {
          id: 'asset-late-location',
          title: 'Late page closet',
          containedAssetCount: 2,
          recentAssetTitles: ['Newer stored blanket', 'Stored blanket']
        }
      ],
      assets: expect.not.arrayContaining([
        expect.objectContaining({ id: 'asset-late-location' })
      ])
    });
    expect(client.listAssetRequests.filter((request) => request.inventoryId === 'inventory-home')).toEqual([
      {
        inventoryId: 'inventory-home',
        limit: 100,
        cursor: undefined,
        lifecycleState: 'all',
        sort: 'updated_desc'
      },
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

it('loads checked-out Home summaries from the checked-out endpoint without attachment enumeration', async () => {
    const client = new FakeInventoryApiClient();
    client.assets = [
      client.assets[0]!,
      {
        ...client.assets[1]!,
        primaryPhoto: client.assets[1]!.primaryPhoto,
        currentCheckout: {
          id: 'checkout-filters',
          state: 'open',
          checkedOutAt: '2026-06-25T12:00:00Z',
          checkedOutByPrincipalId: 'principal-home'
        }
      }
    ];
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getHomeDashboardSnapshot()).resolves.toMatchObject({
      checkedOutAssets: [
        {
          id: 'asset-filters',
          hasPhoto: true,
          photo: {
            uri: 'https://api.example.test/tenants/tenant-home/inventories/inventory-home/assets/asset-filters/attachments/attachment-filters-photo/thumbnail?variant=small'
          },
          currentCheckout: {
            id: 'checkout-filters'
          }
        }
      ]
    });
    expect(client.listAttachmentRequests).toEqual([]);
    expect(client.listCheckedOutAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 10,
      cursor: undefined
    });
  });

it('does not hydrate unrelated inventories or tags for the Home dashboard', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await repository.getHomeDashboardSnapshot();

    expect(client.listAssetRequests.filter((request) => request.inventoryId === 'inventory-cabin')).toEqual([]);
    expect(client.listAssetRequests.filter((request) =>
      request.lifecycleState === 'active' && request.sort === 'id_asc'
    )).toEqual([]);
    expect(client.listAssetTagRequests).toEqual([]);
    expect(client.listAssetRequests).toContainEqual({
      inventoryId: 'inventory-home',
      limit: 10,
      cursor: undefined,
      lifecycleState: 'active',
      sort: 'updated_desc'
    });
  });

it('loads the Assets surface without hydrating tags or unrelated inventories', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getInventoryAssetsSnapshot()).resolves.toMatchObject({
      inventoryName: 'Home Inventory',
      assets: expect.arrayContaining([expect.objectContaining({ id: 'asset-filters' })])
    });

    expect(client.listAssetRequests.filter((request) => request.inventoryId === 'inventory-cabin')).toEqual([]);
    expect(client.listAssetTagRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads the Add surface with only selected inventory identity and tag choices', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getAddAssetContext()).resolves.toMatchObject({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      canAdd: true,
      assetTags: [expect.objectContaining({ id: 'tag-workshop' })]
    });

    expect(client.listAssetTagRequests).toEqual([expect.objectContaining({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home'
    })]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
    expect(client.getAssetRequests).toEqual([]);
  });

it('loads tag filters without hydrating the selected workspace', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getInventoryAssetTags()).resolves.toEqual([
      expect.objectContaining({ id: 'tag-workshop' })
    ]);

    expect(client.listAssetTagRequests).toEqual([expect.objectContaining({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home'
    })]);
    expect(client.listAssetRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads the Locations surface from only the selected active tree', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getLocationsSnapshot()).resolves.toMatchObject({
      tenantName: 'Home',
      inventoryName: 'Home Inventory',
      locations: [expect.objectContaining({ id: 'asset-garage' })]
    });

    expect(client.listAssetRequests).toEqual([expect.objectContaining({
      inventoryId: 'inventory-home',
      lifecycleState: 'active'
    })]);
    expect(client.listAssetTagRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads one location subtree without unrelated inventory or tag hydration', async () => {
    const client = new FakeInventoryApiClient();
    const repository = new ApiInventorySummaryRepository(client, 'tenant-home');

    await expect(repository.getLocationAssetsSnapshot('asset-garage')).resolves.toMatchObject({
      locationId: 'asset-garage',
      assets: [expect.objectContaining({ id: 'asset-filters' })]
    });

    expect(client.listAssetRequests).toEqual([expect.objectContaining({
      inventoryId: 'inventory-home',
      lifecycleState: 'active'
    })]);
    expect(client.listAssetTagRequests).toEqual([]);
    expect(client.listAttachmentRequests).toEqual([]);
  });

it('loads voice identity without asset, tag or attachment reads', async () => {
  const client = new FakeInventoryApiClient();
  const repository = new ApiInventorySummaryRepository(client, 'tenant-home');
  expect(await repository.getVoiceInventoryContext()).toMatchObject({ tenantId: 'tenant-home', inventoryId: 'inventory-home', inventoryName: 'Home Inventory' });
  expect(client.listAssetRequests).toEqual([]);
  expect(client.listAssetTagRequests).toEqual([]);
  expect(client.listAttachmentRequests).toEqual([]);
});
