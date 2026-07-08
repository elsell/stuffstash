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
import type { InventoryMapAssetRepository } from './InventoryMapQuery';
import { toAssetDetailViewModel } from './AssetViewModels';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  private readonly inventory: InventorySummary;
  private readonly detailAssets = new Map<string, AssetSummary>();

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

  async getInventoryWorkspace(): Promise<InventoryWorkspace> {
    return {
      tenants: [{ id: tenantId('tenant-home'), name: 'Home tenant' }],
      defaultInventoryId: inventoryId('inventory-home'),
      inventories: [this.inventory]
    };
  }

  async getDefaultInventorySummary(): Promise<InventorySummary> {
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
      id: 'asset-passport',
      title: 'Passport folder',
      kind: 'container',
      kindLabel: 'Container',
      customTypeLabel: 'Documents',
      description: 'Travel documents and copies.',
      parentAssetId: undefined,
      locationTrailLabel: 'Office closet / Passport folder',
      parentLocationTrailLabel: 'Inventory root',
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
      checkoutActorLabel: undefined,
      canCheckout: true,
      canReturn: false,
      canContainAssets: true,
      canAddContainedAssets: true,
      containedAssetsLabel: '1 thing inside',
      containedAssets: [{
        id: 'asset-birth-certificate',
        title: 'Birth certificate',
        kindLabel: 'Item',
        customTypeLabel: undefined,
        description: '',
        locationTrailLabel: 'Office closet / Passport folder / Birth certificate',
        updatedAtLabel: 'Updated yesterday',
        photoLabel: 'Needs photo',
        photo: undefined,
        imagePlaceholderLabel: 'Item'
      }],
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      photo: undefined,
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

  it('uses the complete detail photo list for map-sourced assets', async () => {
    const repository = new FakeInventorySummaryRepository();
    repository.setDetailAsset({
      id: assetId('asset-living-room-table'),
      title: 'Living room table',
      kind: 'location',
      lifecycleState: 'active',
      locationLabel: 'Inventory root',
      locationTrail: ['Home', 'Living room table'],
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
    expect(toAssetDetailViewModel({
      id: assetId('asset-drill'),
      title: 'Cordless drill',
      kind: 'item',
      lifecycleState: 'active',
      locationLabel: 'Tool bin',
      locationTrail: ['Home', 'Garage', 'Tool bin', 'Cordless drill'],
      description: '',
      updatedAtLabel: 'Updated today',
      hasPhoto: false,
      currentCheckout: {
        id: 'checkout-one',
        state: 'open',
        checkedOutAt: '2026-06-24T11:00:00Z',
        checkedOutByPrincipalId: 'user-one'
      }
    })).toMatchObject({
      isCheckedOut: true,
      checkoutLabel: 'Checked out Jun 24, 2026',
      checkoutActorLabel: 'Checked out by user-one',
      canCheckout: false,
      canReturn: true
    });
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
