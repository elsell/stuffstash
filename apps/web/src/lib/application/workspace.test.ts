import { describe, expect, it } from 'vitest';
import {
  containedAssets,
  detailAssetList,
  labelAsset,
  labelAssets,
  moveParentTargets,
  parentTargets,
  recentlyAddedAssets,
  selectedAssetForDetail,
  topLevelLocations,
  withTrail
} from './workspace';
import type { Asset, CustomAssetType } from '$lib/domain/inventory';

const assets: Asset[] = [
  {
    id: 'garage',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'location',
    title: 'Garage',
    description: '',
    parentAssetId: null,
    lifecycleState: 'active'
  },
  {
    id: 'toolbox',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'container',
    title: 'Toolbox',
    description: '',
    parentAssetId: 'garage',
    lifecycleState: 'active'
  },
  {
    id: 'drill',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title: 'Cordless drill',
    description: '',
    parentAssetId: 'toolbox',
    lifecycleState: 'active'
  },
  {
    id: 'basement',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'location',
    title: 'Basement',
    description: '',
    parentAssetId: 'garage',
    lifecycleState: 'active'
  },
  {
    id: 'basement-toolbox',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'container',
    title: 'Toolbox',
    description: '',
    parentAssetId: 'basement',
    lifecycleState: 'active'
  },
  {
    id: 'archive',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'location',
    title: 'Old shelf',
    description: '',
    parentAssetId: null,
    lifecycleState: 'archived'
  }
];

const crossInventoryAssets: Asset[] = [
  ...assets,
  {
    id: 'other-room',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-two',
    kind: 'location',
    title: 'Other inventory room',
    description: '',
    parentAssetId: null,
    lifecycleState: 'active'
  }
];

const customAssetTypes: CustomAssetType[] = [
  {
    id: 'type-medicine',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    scope: 'inventory',
    key: 'medicine',
    displayName: 'Medicine',
    description: '',
    lifecycleState: 'active'
  }
];

describe('workspace domain helpers', () => {
  it('derives top-level active locations and contained asset counts', () => {
    expect(topLevelLocations(assets)).toEqual([{ location: assets[0], assetCount: 2 }]);
  });

  it('builds containment trails without exposing API DTOs to components', () => {
    expect(withTrail(assets[2]!, assets).containmentTrail).toBe('Garage / Toolbox');
    expect(containedAssets(assets, 'garage')[0]?.containmentTrail).toBe('Garage');
    expect(containedAssets(assets, 'garage').map((asset) => `${asset.title} in ${asset.containmentTrail}`)).toEqual([
      'Toolbox in Garage',
      'Basement in Garage'
    ]);
    expect(parentTargets(assets).filter((asset) => asset.title === 'Toolbox').map((asset) => asset.containmentTrail)).toEqual([
      'Garage',
      'Garage / Basement'
    ]);
  });

  it('keeps valid parent targets to active containers and locations', () => {
    expect(parentTargets(assets).map((asset) => asset.id)).toEqual(['garage', 'toolbox', 'basement', 'basement-toolbox']);
    expect(recentlyAddedAssets(assets).map((asset) => asset.id)).toEqual(['toolbox', 'drill', 'basement-toolbox']);
    expect(recentlyAddedAssets(assets).map((asset) => asset.containmentTrail)).toEqual([
      'Garage',
      'Garage / Toolbox',
      'Garage / Basement'
    ]);
  });

  it('excludes the moving asset and descendants from move parent targets', () => {
    expect(moveParentTargets(assets, 'garage').map((asset) => asset.id)).toEqual([]);
    expect(moveParentTargets(assets, 'toolbox').map((asset) => asset.id)).toEqual(['garage', 'basement', 'basement-toolbox']);
    expect(moveParentTargets(assets, 'drill').map((asset) => asset.id)).toEqual(['garage', 'toolbox', 'basement', 'basement-toolbox']);
    expect(moveParentTargets(crossInventoryAssets, 'other-room').map((asset) => asset.id)).toEqual([]);
  });

  it('labels assets with matching custom asset type display names', () => {
    const medicineAsset: Asset = { ...assets[2]!, customAssetTypeId: 'type-medicine' };
    const alreadyLabeled: Asset = { ...medicineAsset, customAssetTypeLabel: 'Existing label' };

    expect(labelAsset(medicineAsset, customAssetTypes).customAssetTypeLabel).toBe('Medicine');
    expect(labelAsset(alreadyLabeled, customAssetTypes).customAssetTypeLabel).toBe('Existing label');
    expect(labelAssets([medicineAsset], customAssetTypes)[0]?.customAssetTypeLabel).toBe('Medicine');
  });

  it('builds the detail asset list from loaded detail without duplicating known assets', () => {
    const labeledAssets = labelAssets(assets, customAssetTypes);
    const remoteDetail: Asset = {
      ...assets[2]!,
      title: 'Cordless drill with loaded detail',
      customAssetTypeId: 'type-medicine'
    };
    const missingFromList: Asset = { ...remoteDetail, id: 'asset-from-deep-link' };

    expect(detailAssetList(labeledAssets, remoteDetail, customAssetTypes).map((asset) => asset.id)).toEqual(
      labeledAssets.map((asset) => asset.id)
    );
    expect(detailAssetList(labeledAssets, missingFromList, customAssetTypes)[0]).toMatchObject({
      id: 'asset-from-deep-link',
      customAssetTypeLabel: 'Medicine'
    });
  });

  it('selects loaded detail before falling back to the current asset list', () => {
    const labeledAssets = labelAssets(assets, customAssetTypes);
    const loadedDetail: Asset = {
      ...assets[2]!,
      title: 'Loaded drill',
      customAssetTypeId: 'type-medicine'
    };

    expect(selectedAssetForDetail('drill', labeledAssets, loadedDetail, customAssetTypes)).toMatchObject({
      id: 'drill',
      title: 'Loaded drill',
      customAssetTypeLabel: 'Medicine'
    });
    expect(selectedAssetForDetail('toolbox', labeledAssets, loadedDetail, customAssetTypes)?.id).toBe('toolbox');
    expect(selectedAssetForDetail(null, labeledAssets, loadedDetail, customAssetTypes)).toBeNull();
  });
});
