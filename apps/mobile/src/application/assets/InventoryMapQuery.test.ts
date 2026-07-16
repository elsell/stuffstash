import { describe, expect, it } from 'vitest';
import { assetId, AssetSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  InventorySummary,
  tenantId
} from '../../domain/inventories/InventorySummary';
import { InventoryMapAssetRepository, InventoryMapQuery } from './InventoryMapQuery';

class FakeInventoryMapAssetRepository implements InventoryMapAssetRepository {
  constructor(private readonly inventory: InventorySummary) {}

  async listActiveInventoryMapAssets() {
    return {
      sessionScopeId: 'scope-one',
      tenantId: this.inventory.tenantId,
      inventoryId: this.inventory.id,
      inventoryName: this.inventory.name,
      permissions: this.inventory.permissions,
      assets: this.inventory.assets.filter((asset) => asset.lifecycleState === 'active')
    };
  }
}

describe('InventoryMapQuery', () => {
  it('builds a full active containment map for the selected inventory', async () => {
    const query = new InventoryMapQuery(new FakeInventoryMapAssetRepository(inventory([
      asset('item-loose', 'Loose flashlight', 'item'),
      asset('location-garage', 'Garage', 'location'),
      asset('container-bin', 'Seasonal bin', 'container', 'location-garage'),
      asset('item-lights', 'String lights', 'item', 'container-bin'),
      {
        ...asset('item-archived', 'Archived receipt', 'item', 'container-bin'),
        lifecycleState: 'archived'
      }
    ])));

    await expect(query.execute()).resolves.toMatchObject({
      inventoryName: 'Home',
      sessionScopeId: 'scope-one',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-home',
      canCreateAsset: true,
      assets: [
        {
          id: 'location-garage',
          title: 'Garage',
          kindLabel: 'Place',
          childCountLabel: '1 inside',
          canContainAssets: true,
          canAddContainedAssets: true
        },
        {
          id: 'container-bin',
          parentAssetId: 'location-garage',
          title: 'Seasonal bin',
          kindLabel: 'Container',
          childCountLabel: '1 inside'
        },
        {
          id: 'item-lights',
          parentAssetId: 'container-bin',
          title: 'String lights',
          kindLabel: 'Item'
        },
        {
          id: 'item-loose',
          title: 'Loose flashlight',
          kindLabel: 'Item',
          childCountLabel: 'Empty',
          canContainAssets: false,
          canAddContainedAssets: false
        }
      ]
    });
  });

  it('does not expose add-here actions without edit and create permissions', async () => {
    const query = new InventoryMapQuery(new FakeInventoryMapAssetRepository(inventory([
      asset('location-garage', 'Garage', 'location')
    ], ['view', 'create_asset'])));

    await expect(query.execute()).resolves.toMatchObject({
      canCreateAsset: true,
      assets: [{
        id: 'location-garage',
        canContainAssets: true,
        canAddContainedAssets: false
      }]
    });
  });
});

function inventory(
  assets: readonly AssetSummary[],
  permissions: InventorySummary['permissions'] = ['view', 'create_asset', 'edit_asset']
): InventorySummary {
  return {
    id: inventoryId('inventory-home'),
    tenantId: tenantId('tenant-home'),
    name: 'Home',
    role: 'editor',
    permissions,
    description: 'Home inventory.',
    updatedAtLabel: 'Updated today',
    locationCount: assets.filter((candidate) => candidate.kind === 'location').length,
    locations: [],
    assets
  };
}

function asset(
  id: string,
  title: string,
  kind: AssetSummary['kind'],
  parentAssetId?: string
): AssetSummary {
  return {
    id: assetId(id),
    title,
    kind,
    lifecycleState: 'active',
    parentAssetId: parentAssetId ? assetId(parentAssetId) : undefined,
    locationLabel: parentAssetId ? 'Stored' : 'Inventory root',
    locationTrail: parentAssetId ? ['Home', 'Garage', title] : ['Home', title],
    parentLocationTrail: parentAssetId ? [{ id: assetId(parentAssetId), title: 'Garage' }] : [],
    description: `${title} description.`,
    updatedAtLabel: 'Updated today',
    hasPhoto: false
  };
}
