import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import {
  CreateInventoryAssetInput,
  InventorySummaryRepository,
  InventoryWorkspace
} from '../home/InventorySummaryRepository';
import { VoiceInteractionPreviewQuery } from './VoiceInteractionPreviewQuery';

class FakeInventorySummaryRepository implements InventorySummaryRepository {
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
    permissions: ['view', 'create_asset', 'edit_asset'],
    description: 'Home inventory.',
    updatedAtLabel: 'Updated today',
    locationCount: 0,
    locations: [],
    assets: []
  };
}

describe('VoiceInteractionPreviewQuery', () => {
  it('builds a deterministic voice journey preview for the selected inventory', async () => {
    const query = new VoiceInteractionPreviewQuery(new FakeInventorySummaryRepository());

    await expect(query.execute()).resolves.toEqual({
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      sampleUtterance: 'Move the fertilizer from the garage shelf to the wire rack.',
      assistantSummary: 'I found one likely move. Review the plan before anything changes.',
      actionPreview: {
        summary: 'Move fertilizer',
        steps: [
          'Find Fertilizer in Garage shelf',
          'Move it to Wire rack in Garage',
          'Record the change in inventory history'
        ],
        riskLabel: 'Needs approval before saving'
      }
    });
  });
});
