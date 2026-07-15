import { describe, expect, it } from 'vitest';
import {
  containedAssets,
  detailAssetList,
  filterAssets,
  homeLocationPreview,
  labelAsset,
  labelAssets,
  moveParentTargets,
  parentTargets,
  recentlyChangedAssets,
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
  it('limits only the home location preview', () => {
    const locations = Array.from({ length: 12 }, (_, index) => ({
      location: {
        id: `location-${index}`,
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        kind: 'location' as const,
        title: `Location ${index}`,
        description: '',
        parentAssetId: null,
        lifecycleState: 'active' as const
      },
      assetCount: index
    }));

    expect(homeLocationPreview(locations)).toHaveLength(9);
    expect(homeLocationPreview(locations, 'locations')).toHaveLength(12);
  });
  it('derives top-level active locations and contained asset counts', () => {
    expect(topLevelLocations(assets)).toEqual([{ location: assets[0], assetCount: 2 }]);
  });

  it('sorts top-level locations with case-insensitive natural title order', () => {
    const locations = [
      { ...assets[0]!, id: 'bin-10', title: 'Bin 10' },
      { ...assets[0]!, id: 'attic', title: 'attic' },
      { ...assets[0]!, id: 'bin-8', title: 'Bin 8' }
    ];

    expect(topLevelLocations(locations).map((summary) => summary.location.title)).toEqual(['attic', 'Bin 8', 'Bin 10']);
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
    expect(recentlyChangedAssets(assets).map((asset) => asset.id)).toEqual([
      'toolbox',
      'garage',
      'drill',
      'basement-toolbox',
      'basement'
    ]);
    expect(recentlyChangedAssets(assets).map((asset) => asset.containmentTrail)).toEqual([
      'Garage',
      'Inventory root',
      'Garage / Toolbox',
      'Garage / Basement',
      'Garage'
    ]);
  });

  it('orders recently changed places, containers, and items by update time with undated records last', () => {
    const recentCandidates: Asset[] = [
      { ...assets[1]!, id: 'undated', updatedAt: undefined },
      { ...assets[2]!, id: 'newest', updatedAt: '2026-07-14T12:00:00Z' },
      { ...assets[3]!, id: 'location-newest', updatedAt: '2026-07-14T13:00:00Z' },
      { ...assets[4]!, id: 'older', updatedAt: '2026-07-13T12:00:00Z' }
    ];

    expect(recentlyChangedAssets(recentCandidates).map((asset) => asset.id)).toEqual([
      'location-newest',
      'newest',
      'older',
      'undated'
    ]);
  });

  it('ranks search matches by title strength before looser metadata matches', () => {
    const searchAssets: Asset[] = [
      { ...assets[2]!, id: 'description-match', title: 'Drill charger', description: 'garage tape' },
      { ...assets[1]!, id: 'contains-title', title: 'Blue tape roll', description: '' },
      { ...assets[0]!, id: 'exact-title', title: 'Tape', description: '' },
      { ...assets[3]!, id: 'type-match', title: 'Packing labels', description: '', customAssetTypeLabel: 'Tape supplies' },
      { ...assets[4]!, id: 'prefix-title', title: 'Tape measure', description: '' }
    ];

    expect(filterAssets(searchAssets, 'tape').map((asset) => asset.id)).toEqual([
      'exact-title',
      'prefix-title',
      'contains-title',
      'description-match',
      'type-match'
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
