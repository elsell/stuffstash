import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type {
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { CreateAssetCommand } from './CreateAssetCommand';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  createdInput: CreateInventoryAssetInput | undefined;
  addedPhotos: Array<{ readonly assetId: string; readonly fileName: string }> = [];
  shouldFailPhotoUpload = false;

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

  async selectInventory(): Promise<void> {}

  async createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary> {
    this.createdInput = input;

    return {
      id: assetId('asset-created'),
      title: input.title,
      kind: input.kind,
      lifecycleState: 'active',
      locationLabel: 'Garage',
      locationTrail: ['Home', 'Garage', input.title],
      description: input.description,
      updatedAtLabel: 'Updated now',
      hasPhoto: false
    };
  }

  async addAssetPhoto(assetIdValue: string, input: { readonly fileName: string }): Promise<void> {
    if (this.shouldFailPhotoUpload) {
      throw new Error('Photo upload failed.');
    }
    this.addedPhotos.push({ assetId: assetIdValue, fileName: input.fileName });
  }

  async searchAssets(): Promise<readonly AssetSummary[]> {
    return [];
  }

  async searchLocations() {
    return [];
  }

  private readonly inventory: InventorySummary = {
    id: inventoryId('inventory-home'),
    tenantId: tenantId('tenant-home'),
    name: 'Home',
    role: 'editor',
    permissions: ['view', 'create_asset', 'edit_asset'],
    description: 'Home inventory.',
    updatedAtLabel: 'Updated today',
    locationCount: 0,
    locations: [],
    assets: []
  };
}

describe('CreateAssetCommand', () => {
  it('trims base fields and creates an asset through the inventory repository', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new CreateAssetCommand(repository);

    await expect(
      command.execute({
        kind: 'item',
        title: '  Furnace filters  ',
        description: '  MERV 11 three-pack  ',
        parentAssetId: 'asset-garage',
        photos: [
          {
            fileName: 'filters.jpg',
            contentType: 'image/jpeg',
            contentBase64: 'ZmFrZQ=='
          }
        ]
      })
    ).resolves.toEqual({
      id: 'asset-created',
      title: 'Furnace filters',
      message: 'Saved Furnace filters.'
    });

    expect(repository.createdInput).toEqual({
      kind: 'item',
      title: 'Furnace filters',
      description: 'MERV 11 three-pack',
      parentAssetId: 'asset-garage'
    });
    expect(repository.addedPhotos).toEqual([
      { assetId: 'asset-created', fileName: 'filters.jpg' }
    ]);
  });

  it('rejects an empty title before calling the repository', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new CreateAssetCommand(repository);

    await expect(
      command.execute({ kind: 'container', title: '   ', description: 'Cable adapters' })
    ).rejects.toThrow('Name is required.');
    expect(repository.createdInput).toBeUndefined();
  });

  it('reports partial success when photo upload fails after asset creation', async () => {
    const repository = new FakeInventorySummaryRepository();
    repository.shouldFailPhotoUpload = true;
    const command = new CreateAssetCommand(repository);

    await expect(
      command.execute({
        kind: 'item',
        title: 'Flashlight',
        description: '',
        photos: [
          {
            fileName: 'flashlight.jpg',
            contentType: 'image/jpeg',
            contentBase64: 'ZmFrZQ=='
          }
        ]
      })
    ).resolves.toEqual({
      id: 'asset-created',
      title: 'Flashlight',
      message: 'Saved Flashlight, but 1 photo upload failed.'
    });
  });
});
