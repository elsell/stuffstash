import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type {
  CreateInventoryAssetInput,
  GetInventoryAssetDetailInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { AssetDetailQuery } from './AssetDetailQuery';
import type {
  AssetDetailWorkspaceRepository,
  AssetDetailWorkspaceSnapshot
} from './AssetDetailWorkspaceRepository';
import type { InventoryMapAssetRepository } from './InventoryMapQuery';
import { toAssetDetailViewModel } from './AssetViewModels';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  private inventory: InventorySummary;
  private readonly detailAssets = new Map<string, AssetSummary>();
  defaultSummaryRequestCount = 0;

  constructor(permissions: InventorySummary['permissions'] = ['view', 'create_asset', 'edit_asset']) {
    this.inventory = {
      id: inventoryId('inventory-home'),
      tenantId: tenantId('tenant-home'),
      name: 'Home',
      role: 'editor',
      permissions,
      description: 'Home inventory.',
      updatedAtLabel: 'Updated today',
      locationCount: 0,
      locations: [],
      assets: [
        {
          id: assetId('asset-passport'),
          title: 'Passport folder',
          kind: 'container',
          lifecycleState: 'active',
          locationLabel: 'Office closet',
          locationTrail: ['Home', 'Office closet', 'Passport folder'],
          parentLocationTrail: [{ id: assetId('asset-office-closet'), title: 'Office closet' }],
          customType: 'Documents',
          description: 'Travel documents and copies.',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        {
          id: assetId('asset-birth-certificate'),
          title: 'Birth certificate',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-passport'),
          locationLabel: 'Passport folder',
          locationTrail: ['Home', 'Office closet', 'Passport folder', 'Birth certificate'],
          parentLocationTrail: [
            { id: assetId('asset-office-closet'), title: 'Office closet' },
            { id: assetId('asset-passport'), title: 'Passport folder' }
          ],
          description: '',
          updatedAtLabel: 'Updated yesterday',
          hasPhoto: false
        }
      ]
    };
  }

  setDetailAsset(asset: AssetSummary): void {
    this.detailAssets.set(asset.id, asset);
  }

  setSummaryAssets(assets: readonly AssetSummary[]): void {
    this.inventory = { ...this.inventory, assets };
  }

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return {
      tenants: [{ id: tenantId('tenant-home'), name: 'Home tenant' }],
      defaultInventoryId: inventoryId('inventory-home'),
      inventories: [this.inventory]
    };
  }

  async getDefaultInventorySummary(): Promise<InventorySummary> {
    this.defaultSummaryRequestCount += 1;
    return this.inventory;
  }

  async getAssetDetail(input: GetInventoryAssetDetailInput): Promise<AssetSummary> {
    const asset = this.detailAssets.get(input.asset.id);
    if (asset) {
      return asset;
    }
    return input.asset;
  }

  async selectInventory(): Promise<void> {}

  async createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary> {
    return {
      id: assetId('created-asset'),
      title: input.title,
      kind: input.kind,
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', input.title],
      parentLocationTrail: [],
      description: input.description,
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    };
  }

  async addAssetPhoto(): Promise<void> {}

  async archiveAsset(): Promise<void> {}

  async restoreAsset(): Promise<void> {}

  async deleteAsset(): Promise<void> {}

  async browseAssets() {
    return { assets: [], hasMore: false };
  }

  async searchAssets(): Promise<readonly AssetSummary[]> {
    return [];
  }

  async searchLocations() {
    return [];
  }

}

describe('AssetDetailQuery', () => {
  it('builds a reusable asset workspace view model', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-passport')).resolves.toEqual({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      id: 'asset-passport',
      title: 'Passport folder',
      kind: 'container',
      kindLabel: 'Container',
      customTypeLabel: 'Documents',
      description: 'Travel documents and copies.',
      parentAssetId: undefined,
      locationTrailLabel: 'Office closet / Passport folder',
      parentLocationTrail: [{ id: 'asset-office-closet', title: 'Office closet', isImmediateParent: true }],
      parentLocationTrailLabel: 'Office closet',
      lifecycleLabel: 'Active',
      isActive: true,
      canEdit: true,
      canMove: true,
      canAddPhotos: true,
      canArchive: true,
      canRestore: false,
      canDeletePermanently: false,
      isCheckedOut: false,
      checkoutLabel: 'Available',
      canCheckout: true,
      canReturn: false,
      canContainAssets: true,
      canAddContainedAssets: true,
      containedAssetsLabel: '1 thing inside',
      containedSpaces: [],
      containedSpacesLabel: '0 spaces',
      containedItems: [],
      containedItemsLabel: '0 items',
      containedAssets: [{
        id: 'asset-birth-certificate',
        title: 'Birth certificate',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: '',
        locationTrailLabel: 'Office closet / Passport folder / Birth certificate',
        parentLocationTrail: [
          { id: 'asset-office-closet', title: 'Office closet', isImmediateParent: false },
          { id: 'asset-passport', title: 'Passport folder', isImmediateParent: true }
        ],
        updatedAtLabel: 'Updated yesterday',
        photoLabel: 'Needs photo',
        imagePlaceholderLabel: 'Item'
      }],
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      imagePlaceholderLabel: 'Box'
    });
  });

  it('rejects an unknown asset in the selected inventory', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository());

    await expect(query.execute('asset-missing')).rejects.toThrow(
      'Asset is not available in the selected inventory.'
    );
  });

  it('uses the detail asset photo list when the repository can load complete media', async () => {
    const repository = new FakeInventorySummaryRepository();
    repository.setDetailAsset({
      id: assetId('asset-passport'),
      title: 'Passport folder',
      kind: 'container',
      lifecycleState: 'active',
      locationLabel: 'Office closet',
      locationTrail: ['Home', 'Office closet', 'Passport folder'],
      parentLocationTrail: [{ id: assetId('asset-office-closet'), title: 'Office closet' }],
      customType: 'Documents',
      description: 'Travel documents and copies.',
      updatedAtLabel: 'Updated today',
      hasPhoto: true,
      photos: [
        { id: 'attachment-front', fileName: 'front.jpg', uri: 'https://api.example.test/front-small.jpg' },
        { id: 'attachment-back', fileName: 'back.jpg', uri: 'https://api.example.test/back-small.jpg' }
      ]
    });
    const query = new AssetDetailQuery(repository);

    await expect(query.execute('asset-passport')).resolves.toMatchObject({
      id: 'asset-passport',
      photos: [
        { id: 'attachment-front', label: 'front.jpg' },
        { id: 'attachment-back', label: 'back.jpg' }
      ],
      containedAssetsLabel: '1 thing inside'
    });
  });

  it('loads details from the selected active map tree when Map opens a visible row', async () => {
    const repository = new FakeInventorySummaryRepository();
    const query = new AssetDetailQuery(
      repository,
      new FakeInventoryMapAssetRepository([
        {
          id: assetId('asset-living-room-table'),
          title: 'Living room table',
          kind: 'location',
          lifecycleState: 'active',
          locationLabel: 'Inventory root',
          locationTrail: ['Home', 'Living room table'],
          parentLocationTrail: [],
          description: 'Temporary landing spot.',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        }
      ])
    );

    await expect(query.execute('asset-living-room-table', { source: 'map' })).resolves.toMatchObject({
      id: 'asset-living-room-table',
      title: 'Living room table',
      kind: 'location',
      canAddContainedAssets: true
    });
  });

  it('loads the complete map tree for a place opened from a partial inventory summary', async () => {
    const repository = new FakeInventorySummaryRepository();
    const garage: AssetSummary = {
      id: assetId('asset-garage'), title: 'Garage', kind: 'location', lifecycleState: 'active',
      locationLabel: 'Inventory root', locationTrail: ['Home', 'Garage'], parentLocationTrail: [],
      description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    const cabinet: AssetSummary = {
      id: assetId('asset-cabinet'), title: 'Cabinet', kind: 'container', lifecycleState: 'active',
      parentAssetId: garage.id, locationLabel: 'Garage', locationTrail: ['Home', 'Garage', 'Cabinet'],
      parentLocationTrail: [{ id: garage.id, title: garage.title }], description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    repository.setSummaryAssets([garage]);
    const query = new AssetDetailQuery(
      repository,
      new FakeInventoryMapAssetRepository([
        garage,
        cabinet,
        {
          id: assetId('asset-hammer'), title: 'Hammer', kind: 'item', lifecycleState: 'active',
          parentAssetId: cabinet.id, locationLabel: 'Cabinet', locationTrail: ['Home', 'Garage', 'Cabinet', 'Hammer'],
          parentLocationTrail: [{ id: garage.id, title: garage.title }, { id: cabinet.id, title: cabinet.title }],
          description: '', updatedAtLabel: 'Updated today', hasPhoto: false
        }
      ])
    );

    await expect(query.execute('asset-garage')).resolves.toMatchObject({
      containedSpaces: [{ id: 'asset-cabinet' }],
      containedItems: [{ id: 'asset-hammer', relativePathLabel: 'Cabinet' }]
    });
  });

  it('opens an active place from one dedicated workspace snapshot without loading summary or Map', async () => {
    const summaries = new FakeInventorySummaryRepository();
    const mapAssets = new CountingInventoryMapAssetRepository();
    const workspace: AssetDetailWorkspaceSnapshot = {
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      permissions: ['view', 'create_asset', 'edit_asset'],
      asset: {
        id: assetId('asset-garage'), title: 'Garage', kind: 'location', lifecycleState: 'active',
        locationLabel: 'Inventory root', locationTrail: ['Home', 'Garage'], parentLocationTrail: [],
        description: '', updatedAtLabel: 'Updated today', hasPhoto: false
      },
      allAssets: [{
        id: assetId('asset-hammer'), title: 'Hammer', kind: 'item', lifecycleState: 'active',
        parentAssetId: assetId('asset-garage'), locationLabel: 'Garage', locationTrail: ['Home', 'Garage', 'Hammer'],
        parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
        description: '', updatedAtLabel: 'Updated today', hasPhoto: false
      }]
    };
    const workspaces = new FakeAssetDetailWorkspaceRepository(workspace);
    const query = new AssetDetailQuery(summaries, mapAssets, workspaces);

    await expect(query.execute('asset-garage')).resolves.toMatchObject({
      id: 'asset-garage',
      containedItems: [{ id: 'asset-hammer' }]
    });
    expect(workspaces.requestedAssetIds).toEqual(['asset-garage']);
    expect(summaries.defaultSummaryRequestCount).toBe(0);
    expect(mapAssets.requestCount).toBe(0);
  });

  it('opens an ordinary item from a target-only workspace without falling back to the default summary', async () => {
    const summaries = new FakeInventorySummaryRepository();
    const mapAssets = new CountingInventoryMapAssetRepository();
    const item: AssetSummary = {
      id: assetId('asset-hammer'), title: 'Hammer', kind: 'item', lifecycleState: 'active',
      parentAssetId: assetId('asset-garage'), locationLabel: 'Garage', locationTrail: ['Home', 'Garage', 'Hammer'],
      parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
      description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    const workspaces = new FakeAssetDetailWorkspaceRepository({
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      permissions: ['view', 'edit_asset'],
      asset: item,
      allAssets: []
    });
    const query = new AssetDetailQuery(summaries, mapAssets, workspaces);

    await expect(query.execute('asset-hammer')).resolves.toMatchObject({
      id: 'asset-hammer',
      kind: 'item',
      parentLocationTrail: [{ id: 'asset-garage', title: 'Garage' }]
    });
    expect(workspaces.requestedAssetIds).toEqual(['asset-hammer']);
    expect(summaries.defaultSummaryRequestCount).toBe(0);
    expect(mapAssets.requestCount).toBe(0);
  });

  it('uses the complete detail photo list for map-sourced assets', async () => {
    const repository = new FakeInventorySummaryRepository();
    repository.setDetailAsset({
      id: assetId('asset-living-room-table'),
      title: 'Living room table',
      kind: 'location',
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Living room table'],
      parentLocationTrail: [],
      description: 'Temporary landing spot.',
      updatedAtLabel: 'Updated today',
      hasPhoto: true,
      photos: [
        { id: 'attachment-table', fileName: 'table.jpg', uri: 'https://api.example.test/table-small.jpg' }
      ]
    });
    const query = new AssetDetailQuery(
      repository,
      new FakeInventoryMapAssetRepository([
        {
          id: assetId('asset-living-room-table'),
          title: 'Living room table',
          kind: 'location',
          lifecycleState: 'active',
          locationLabel: 'Inventory root',
          locationTrail: ['Home', 'Living room table'],
          parentLocationTrail: [],
          description: 'Temporary landing spot.',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        }
      ])
    );

    await expect(query.execute('asset-living-room-table', { source: 'map' })).resolves.toMatchObject({
      id: 'asset-living-room-table',
      photos: [
        { id: 'attachment-table', label: 'table.jpg' }
      ]
    });
  });

  it('prefers the selected active map tree when Map opens an asset that also appears in the recent summary', async () => {
    const query = new AssetDetailQuery(
      new FakeInventorySummaryRepository(),
      new FakeInventoryMapAssetRepository([
        {
          id: assetId('asset-passport'),
          title: 'Passport folder',
          kind: 'container',
          lifecycleState: 'active',
          locationLabel: 'Inventory root',
          locationTrail: ['Home', 'Map workspace', 'Passport folder'],
          parentLocationTrail: [{ id: assetId('asset-map-workspace'), title: 'Map workspace' }],
          description: 'Map source description.',
          updatedAtLabel: 'Updated from map',
          hasPhoto: false
        },
        {
          id: assetId('asset-map-child'),
          title: 'Map child',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-passport'),
          locationLabel: 'Passport folder',
          locationTrail: ['Home', 'Map workspace', 'Passport folder', 'Map child'],
          parentLocationTrail: [
            { id: assetId('asset-map-workspace'), title: 'Map workspace' },
            { id: assetId('asset-passport'), title: 'Passport folder' }
          ],
          description: '',
          updatedAtLabel: 'Updated from map',
          hasPhoto: false
        }
      ])
    );

    await expect(query.execute('asset-passport', { source: 'map' })).resolves.toMatchObject({
      id: 'asset-passport',
      description: 'Map source description.',
      locationTrailLabel: 'Map workspace / Passport folder',
      containedAssets: [
        { id: 'asset-map-child', title: 'Map child' }
      ]
    });
  });

  it('does not expose lifecycle controls without edit permission', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository(['view']));

    await expect(query.execute('asset-passport')).resolves.toMatchObject({
      canArchive: false,
      canRestore: false,
      canDeletePermanently: false
    });
  });

  it('exposes return action instead of checkout for checked-out assets', () => {
    const detail = toAssetDetailViewModel({
      id: assetId('asset-drill'),
      title: 'Cordless drill',
      kind: 'item',
      lifecycleState: 'active',
      locationLabel: 'Tool bin',
      locationTrail: ['Home', 'Garage', 'Tool bin', 'Cordless drill'],
      parentLocationTrail: [
        { id: assetId('asset-garage'), title: 'Garage' },
        { id: assetId('asset-tool-bin'), title: 'Tool bin' }
      ],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false,
      currentCheckout: {
        id: 'checkout-one',
        state: 'open',
        checkedOutAt: '2026-06-24T11:00:00Z',
        checkedOutByPrincipalId: 'user-one'
      }
    });

    expect(detail).toMatchObject({
      isCheckedOut: true,
      checkoutLabel: 'Checked out Jun 24, 2026',
      canCheckout: false,
      canReturn: true
    });
    expect(detail).not.toHaveProperty('checkoutActorLabel');
  });

  it('allows an existing location checkout to be returned without offering a new checkout', () => {
    const detail = toAssetDetailViewModel({
      id: assetId('asset-workshop'),
      title: 'Workshop',
      kind: 'location',
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Workshop'],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false,
      currentCheckout: {
        id: 'checkout-legacy',
        state: 'open',
        checkedOutAt: '2026-06-24T11:00:00Z',
        checkedOutByPrincipalId: 'user-one'
      }
    });

    expect(detail).toMatchObject({ canCheckout: false, canReturn: true });
  });

  it('requires create permission before offering Add item here inside containers', async () => {
    const query = new AssetDetailQuery(new FakeInventorySummaryRepository(['view', 'edit_asset']));

    await expect(query.execute('asset-passport')).resolves.toMatchObject({
      canContainAssets: true,
      canAddContainedAssets: false
    });
  });

  it('keeps archived containers spatial but disables adding contained assets', () => {
    expect(toAssetDetailViewModel({
      id: assetId('asset-archive-box'),
      title: 'Archive box',
      kind: 'container',
      lifecycleState: 'archived',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Archive box'],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    })).toMatchObject({
      canContainAssets: true,
      canAddContainedAssets: false,
      canEdit: false,
      canMove: false
    });
  });

  it('carries safe attachment metadata into asset detail photo view models', () => {
    expect(toAssetDetailViewModel({
      id: assetId('asset-water-bottle'),
      title: 'Water bottle',
      kind: 'item',
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Water bottle'],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: true,
      photos: [{
        id: 'attachment-one',
        fileName: 'bottle.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 1536000,
        uri: 'https://photos/bottle.jpg'
      }]
    })).toMatchObject({
      photos: [{
        id: 'attachment-one',
        fileName: 'bottle.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 1536000
      }]
    });
  });

  it('orders contained assets by spatial usefulness and stable title within groups', () => {
    expect(toAssetDetailViewModel({
      id: assetId('asset-garage-shelf'),
      title: 'Garage shelf',
      kind: 'location',
      lifecycleState: 'active',
      locationLabel: 'Garage',
      locationTrail: ['Home', 'Garage', 'Garage shelf'],
      parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    }, {
      allAssets: [
        {
          id: assetId('asset-zipties'),
          title: 'Zip ties',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-garage-shelf'),
          locationLabel: 'Garage shelf',
          locationTrail: ['Home', 'Garage', 'Garage shelf', 'Zip ties'],
          parentLocationTrail: [
            { id: assetId('asset-garage'), title: 'Garage' },
            { id: assetId('asset-garage-shelf'), title: 'Garage shelf' }
          ],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        {
          id: assetId('asset-cable-bin-backup'),
          title: 'Cable bin',
          kind: 'container',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-garage-shelf'),
          locationLabel: 'Garage shelf',
          locationTrail: ['Home', 'Garage', 'Garage shelf', 'Cable bin'],
          parentLocationTrail: [
            { id: assetId('asset-garage'), title: 'Garage' },
            { id: assetId('asset-garage-shelf'), title: 'Garage shelf' }
          ],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        {
          id: assetId('asset-cable-bin'),
          title: 'Cable bin',
          kind: 'container',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-garage-shelf'),
          locationLabel: 'Garage shelf',
          locationTrail: ['Home', 'Garage', 'Garage shelf', 'Cable bin'],
          parentLocationTrail: [
            { id: assetId('asset-garage'), title: 'Garage' },
            { id: assetId('asset-garage-shelf'), title: 'Garage shelf' }
          ],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        {
          id: assetId('asset-archive-cubby'),
          title: 'Archive cubby',
          kind: 'location',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-garage-shelf'),
          locationLabel: 'Garage shelf',
          locationTrail: ['Home', 'Garage', 'Garage shelf', 'Archive cubby'],
          parentLocationTrail: [
            { id: assetId('asset-garage'), title: 'Garage' },
            { id: assetId('asset-garage-shelf'), title: 'Garage shelf' }
          ],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        {
          id: assetId('asset-aa-batteries'),
          title: 'AA batteries',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: assetId('asset-garage-shelf'),
          locationLabel: 'Garage shelf',
          locationTrail: ['Home', 'Garage', 'Garage shelf', 'AA batteries'],
          parentLocationTrail: [
            { id: assetId('asset-garage'), title: 'Garage' },
            { id: assetId('asset-garage-shelf'), title: 'Garage shelf' }
          ],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        }
      ]
    }).containedAssets.map((asset) => asset.id)).toEqual([
      'asset-archive-cubby',
      'asset-cable-bin',
      'asset-cable-bin-backup',
      'asset-aa-batteries',
      'asset-zipties'
    ]);
  });

  it('builds a place workspace from direct spaces and recursively nested items', () => {
    const garage = {
      id: assetId('asset-garage'),
      title: 'Garage',
      kind: 'location' as const,
      lifecycleState: 'active' as const,
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Garage'],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    };
    const toolCabinet = {
      id: assetId('asset-tool-cabinet'),
      title: 'Tool cabinet',
      kind: 'container' as const,
      lifecycleState: 'active' as const,
      parentAssetId: garage.id,
      locationLabel: 'Garage',
      locationTrail: ['Home', 'Garage', 'Tool cabinet'],
      parentLocationTrail: [{ id: garage.id, title: garage.title }],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    };
    const drawer = {
      id: assetId('asset-drawer-two'),
      title: 'Drawer 2',
      kind: 'container' as const,
      lifecycleState: 'active' as const,
      parentAssetId: toolCabinet.id,
      locationLabel: 'Tool cabinet',
      locationTrail: ['Home', 'Garage', 'Tool cabinet', 'Drawer 2'],
      parentLocationTrail: [
        { id: garage.id, title: garage.title },
        { id: toolCabinet.id, title: toolCabinet.title }
      ],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    };

    const detail = toAssetDetailViewModel(garage, {
      allAssets: [
        {
          id: assetId('asset-extension-cord'),
          title: 'Extension cord',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: drawer.id,
          locationLabel: 'Drawer 2',
          locationTrail: ['Home', 'Garage', 'Tool cabinet', 'Drawer 2', 'Extension cord'],
          parentLocationTrail: [
            { id: garage.id, title: garage.title },
            { id: toolCabinet.id, title: toolCabinet.title },
            { id: drawer.id, title: drawer.title }
          ],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        {
          id: assetId('asset-broom'),
          title: 'Broom',
          kind: 'item',
          lifecycleState: 'active',
          parentAssetId: garage.id,
          locationLabel: 'Garage',
          locationTrail: ['Home', 'Garage', 'Broom'],
          parentLocationTrail: [{ id: garage.id, title: garage.title }],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        },
        drawer,
        toolCabinet,
        {
          id: assetId('asset-workshop'),
          title: 'Workshop',
          kind: 'location',
          lifecycleState: 'active',
          parentAssetId: garage.id,
          locationLabel: 'Garage',
          locationTrail: ['Home', 'Garage', 'Workshop'],
          parentLocationTrail: [{ id: garage.id, title: garage.title }],
          description: '',
          updatedAtLabel: 'Updated today',
          hasPhoto: false
        }
      ]
    });

    expect(detail).toMatchObject({
      canCheckout: false,
      canReturn: false,
      containedSpacesLabel: '2 spaces',
      containedItemsLabel: '2 items',
      containedSpaces: [
        { id: 'asset-tool-cabinet', title: 'Tool cabinet' },
        { id: 'asset-workshop', title: 'Workshop' }
      ],
      containedItems: [
        { id: 'asset-broom', relativePath: [], relativePathLabel: undefined },
        {
          id: 'asset-extension-cord',
          relativePath: [
            { id: 'asset-tool-cabinet', title: 'Tool cabinet' },
            { id: 'asset-drawer-two', title: 'Drawer 2' }
          ],
          relativePathLabel: 'Tool cabinet / Drawer 2'
        }
      ]
    });
    expect(detail.containedSpaces.map((asset) => asset.id)).not.toContain('asset-drawer-two');
  });

  it('keeps container contents limited to immediate children', () => {
    const cabinet = {
      id: assetId('asset-cabinet'),
      title: 'Cabinet',
      kind: 'container' as const,
      lifecycleState: 'active' as const,
      locationLabel: 'Garage',
      locationTrail: ['Home', 'Garage', 'Cabinet'],
      parentLocationTrail: [{ id: assetId('asset-garage'), title: 'Garage' }],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    };

    const detail = toAssetDetailViewModel(cabinet, {
      allAssets: [
        {
          id: assetId('asset-drawer'), title: 'Drawer', kind: 'container', lifecycleState: 'active',
          parentAssetId: cabinet.id, locationLabel: 'Cabinet', locationTrail: ['Home', 'Garage', 'Cabinet', 'Drawer'],
          parentLocationTrail: [{ id: cabinet.id, title: cabinet.title }], description: '', updatedAtLabel: 'Updated today', hasPhoto: false
        },
        {
          id: assetId('asset-screwdriver'), title: 'Screwdriver', kind: 'item', lifecycleState: 'active',
          parentAssetId: assetId('asset-drawer'), locationLabel: 'Drawer', locationTrail: ['Home', 'Garage', 'Cabinet', 'Drawer', 'Screwdriver'],
          parentLocationTrail: [{ id: cabinet.id, title: cabinet.title }, { id: assetId('asset-drawer'), title: 'Drawer' }],
          description: '', updatedAtLabel: 'Updated today', hasPhoto: false
        }
      ]
    });

    expect(detail.containedAssets.map((asset) => asset.id)).toEqual(['asset-drawer']);
    expect(detail.containedSpaces).toEqual([]);
    expect(detail.containedItems).toEqual([]);
  });

  it('indexes a deep place subtree with linear parent-link reads', () => {
    const garage: AssetSummary = {
      id: assetId('asset-garage-linear'), title: 'Garage', kind: 'location', lifecycleState: 'active',
      locationLabel: 'Inventory root', locationTrail: ['Home', 'Garage'], parentLocationTrail: [],
      description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    let parentLinkReads = 0;
    const spaces: AssetSummary[] = [];
    let parentId = garage.id;
    for (let index = 0; index < 150; index += 1) {
      const currentParentId = parentId;
      const id = assetId(`asset-space-${index.toString()}`);
      spaces.push({
        id,
        title: `Space ${index.toString()}`,
        kind: 'container',
        lifecycleState: 'active',
        get parentAssetId() {
          parentLinkReads += 1;
          return currentParentId;
        },
        locationLabel: 'Garage',
        locationTrail: ['Home', 'Garage', `Space ${index.toString()}`],
        parentLocationTrail: [],
        description: '',
        updatedAtLabel: 'Updated today',
        hasPhoto: false
      });
      parentId = id;
    }
    const items = Array.from({ length: 150 }, (_, index): AssetSummary => ({
      id: assetId(`asset-item-${index.toString()}`),
      title: `Item ${index.toString()}`,
      kind: 'item',
      lifecycleState: 'active',
      get parentAssetId() {
        parentLinkReads += 1;
        return parentId;
      },
      locationLabel: 'Deep space',
      locationTrail: ['Home', 'Garage', 'Deep space', `Item ${index.toString()}`],
      parentLocationTrail: [],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false
    }));

    const detail = toAssetDetailViewModel(garage, { allAssets: [...spaces, ...items] });

    expect(detail.containedItems).toHaveLength(150);
    expect(detail.containedItems[0]?.relativePath).toHaveLength(150);
    expect(parentLinkReads).toBeLessThanOrEqual(900);
  });

  it('terminates cyclic child links and emits each reachable asset at most once', () => {
    const garage: AssetSummary = {
      id: assetId('asset-garage-cycle'), title: 'Garage', kind: 'location', lifecycleState: 'active',
      locationLabel: 'Inventory root', locationTrail: ['Home', 'Garage'], parentLocationTrail: [],
      description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    const spaceA: AssetSummary = {
      id: assetId('asset-space-a'), title: 'Space A', kind: 'container', lifecycleState: 'active',
      parentAssetId: garage.id, locationLabel: 'Garage', locationTrail: ['Home', 'Garage', 'Space A'],
      parentLocationTrail: [], description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    const cyclicA: AssetSummary = {
      id: assetId('asset-cycle-a'), title: 'Cycle A', kind: 'container', lifecycleState: 'active',
      parentAssetId: assetId('asset-cycle-b'), locationLabel: '', locationTrail: [], parentLocationTrail: [],
      description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };
    const cyclicB: AssetSummary = {
      ...cyclicA,
      id: assetId('asset-cycle-b'),
      title: 'Cycle B',
      parentAssetId: cyclicA.id
    };
    const wrench: AssetSummary = {
      id: assetId('asset-wrench'), title: 'Wrench', kind: 'item', lifecycleState: 'active',
      parentAssetId: spaceA.id, locationLabel: 'Space A', locationTrail: ['Home', 'Garage', 'Space A', 'Wrench'],
      parentLocationTrail: [], description: '', updatedAtLabel: 'Updated today', hasPhoto: false
    };

    const detail = toAssetDetailViewModel(garage, {
      allAssets: [cyclicA, cyclicB, wrench, spaceA, wrench]
    });

    expect(detail.containedSpaces.map((asset) => asset.id)).toEqual(['asset-space-a']);
    expect(detail.containedItems.map((asset) => asset.id)).toEqual(['asset-wrench']);
  });
});

class FakeInventoryMapAssetRepository implements InventoryMapAssetRepository {
  constructor(private readonly assets: readonly AssetSummary[]) {}

  async listActiveInventoryMapAssets() {
    return {
      sessionScopeId: 'scope-one',
      tenantId: tenantId('tenant-home'),
      inventoryId: inventoryId('inventory-home'),
      inventoryName: 'Home',
      permissions: ['view', 'create_asset', 'edit_asset'],
      assets: this.assets
    };
  }
}

class CountingInventoryMapAssetRepository implements InventoryMapAssetRepository {
  requestCount = 0;

  listActiveInventoryMapAssets(): ReturnType<InventoryMapAssetRepository['listActiveInventoryMapAssets']> {
    this.requestCount += 1;
    return Promise.reject(new Error('Map traversal should not run.'));
  }
}

class FakeAssetDetailWorkspaceRepository implements AssetDetailWorkspaceRepository {
  readonly requestedAssetIds: string[] = [];

  constructor(private readonly snapshot: AssetDetailWorkspaceSnapshot | undefined) {}

  async getAssetDetailWorkspace(selectedAssetId: AssetSummary['id']) {
    this.requestedAssetIds.push(selectedAssetId);
    return this.snapshot;
  }
}
