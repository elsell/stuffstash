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
import { AssetLifecycleCommand } from './AssetLifecycleCommand';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
  readonly lifecycleCalls: Array<{ readonly action: string; readonly assetId: string }> = [];

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

  async searchAssets(): Promise<readonly AssetSummary[]> {
    return [];
  }

  async searchLocations() {
    return [];
  }

  async archiveAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    this.lifecycleCalls.push({ action: 'archive', assetId: assetIdValue });
  }

  async restoreAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    this.lifecycleCalls.push({ action: 'restore', assetId: assetIdValue });
  }

  async deleteAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    this.lifecycleCalls.push({ action: 'delete', assetId: assetIdValue });
  }

  async browseAssets() {
    return { assets: [], hasMore: false };
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

describe('AssetLifecycleCommand', () => {
  it('archives, restores, and deletes assets through the inventory repository port', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new AssetLifecycleCommand(repository);

    await command.execute({ action: 'archive', assetId: 'asset-garage' });
    await command.execute({ action: 'restore', assetId: 'asset-garage' });
    await command.execute({ action: 'delete', assetId: 'asset-garage' });

    expect(repository.lifecycleCalls).toEqual([
      { action: 'archive', assetId: 'asset-garage' },
      { action: 'restore', assetId: 'asset-garage' },
      { action: 'delete', assetId: 'asset-garage' }
    ]);
  });

  it('rejects unsupported lifecycle actions', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new AssetLifecycleCommand(repository);

    await expect(
      command.execute({ action: 'unsupported', assetId: 'asset-garage' })
    ).rejects.toThrow('Unsupported asset lifecycle action.');
    expect(repository.lifecycleCalls).toEqual([]);
  });
});
