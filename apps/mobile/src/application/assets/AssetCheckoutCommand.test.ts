import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import type {
  AssetCheckoutInput,
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { AssetCheckoutCommand } from './AssetCheckoutCommand';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  readonly checkoutCalls: Array<{ readonly action: string; readonly assetId: string; readonly details?: string }> = [];

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

  async checkoutAsset(assetIdValue: AssetSummary['id'], input: AssetCheckoutInput = {}): Promise<void> {
    this.checkoutCalls.push({ action: 'checkout', assetId: assetIdValue, details: input.details });
  }

  async returnAsset(assetIdValue: AssetSummary['id'], input: AssetCheckoutInput = {}): Promise<void> {
    this.checkoutCalls.push({ action: 'return', assetId: assetIdValue, details: input.details });
  }

  async browseAssets() {
    return { assets: [], hasMore: false };
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
    permissions: ['view', 'edit_asset'],
    description: 'Home inventory.',
    updatedAtLabel: 'Updated today',
    locationCount: 0,
    locations: [],
    assets: []
  };
}

describe('AssetCheckoutCommand', () => {
  it('checks out and returns assets through the inventory repository port', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new AssetCheckoutCommand(repository);

    await command.execute({ action: 'checkout', assetId: 'asset-drill', details: 'using at bench' });
    await command.execute({ action: 'return', assetId: 'asset-drill' });

    expect(repository.checkoutCalls).toEqual([
      { action: 'checkout', assetId: 'asset-drill', details: 'using at bench' },
      { action: 'return', assetId: 'asset-drill', details: undefined }
    ]);
  });

  it('rejects unsupported checkout actions', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new AssetCheckoutCommand(repository);

    await expect(
      command.execute({ action: 'unsupported', assetId: 'asset-drill' })
    ).rejects.toThrow('Unsupported asset checkout action.');
    expect(repository.checkoutCalls).toEqual([]);
  });
});
