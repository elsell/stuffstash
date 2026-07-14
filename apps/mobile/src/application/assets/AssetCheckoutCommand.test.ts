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
  readonly undoCalls: string[] = [];

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

  async checkoutAsset(assetIdValue: AssetSummary['id'], input: AssetCheckoutInput = {}) {
    this.checkoutCalls.push({ action: 'checkout', assetId: assetIdValue, details: input.details });
    return { id: 'checkout-one', assetId: assetIdValue, undoableOperationId: 'operation-checkout-one' };
  }

  async returnAsset(assetIdValue: AssetSummary['id'], input: AssetCheckoutInput = {}) {
    this.checkoutCalls.push({ action: 'return', assetId: assetIdValue, details: input.details });
    return { id: 'checkout-one', assetId: assetIdValue, undoableOperationId: 'operation-return-one' };
  }

  async updateReturnedCheckoutDetails(assetIdValue: AssetSummary['id'], checkoutId: string, input: AssetCheckoutInput = {}) {
    this.checkoutCalls.push({ action: `details:${checkoutId}`, assetId: assetIdValue, details: input.details });
    return { id: checkoutId, assetId: assetIdValue };
  }

  async undoInventoryOperation(operationId: string): Promise<void> {
    this.undoCalls.push(operationId);
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

    await expect(command.execute({ action: 'checkout', assetId: 'asset-drill', details: 'using at bench' })).resolves.toEqual({
      id: 'checkout-one',
      assetId: 'asset-drill',
      undoableOperationId: 'operation-checkout-one'
    });
    await expect(command.execute({ action: 'return', assetId: 'asset-drill' })).resolves.toEqual({
      id: 'checkout-one',
      assetId: 'asset-drill',
      undoableOperationId: 'operation-return-one'
    });

    expect(repository.checkoutCalls).toEqual([
      { action: 'checkout', assetId: 'asset-drill', details: 'using at bench' },
      { action: 'return', assetId: 'asset-drill', details: undefined }
    ]);
  });

  it('updates return details and applies undo through the inventory repository port', async () => {
    const repository = new FakeInventorySummaryRepository();
    const command = new AssetCheckoutCommand(repository);

    await command.updateReturnedCheckoutDetails({ assetId: 'asset-drill', checkoutId: 'checkout-one', details: 'back in bin' });
    await command.undoOperation({ operationId: 'operation-return-one' });

    expect(repository.checkoutCalls).toEqual([
      { action: 'details:checkout-one', assetId: 'asset-drill', details: 'back in bin' }
    ]);
    expect(repository.undoCalls).toEqual(['operation-return-one']);
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
