import { describe, expect, it } from 'vitest';
import type { InventoryMapViewModel } from '../../application/assets/InventoryMapQuery';
import {
  buildBrowseSurfaceOptions,
  buildInventoryMapBreadcrumbs,
  buildInventoryMapColumns,
  buildInventoryMapEmptyColumnAction,
  buildInventoryMapRowInteractionState,
  clampInventoryMapOffset,
  findInventoryMapSearchMatch,
  inventoryMapBranchSwipeOffset,
  inventoryMapGestureConfig,
  mapOverviewLabel,
  nearestInventoryMapColumnForOffset,
  pathForBreadcrumbLevel,
  preserveInventoryMapHighlightForPath,
  selectInventoryMapBranch,
  shouldActivateInventoryMapPagerPan,
  shouldActivateInventoryMapBranchSwipe,
  shouldSelectInventoryMapBranchDuringSwipe,
  shouldSuppressInventoryMapScrollForBranchSwipe
} from './InventoryMapPresentation';

describe('InventoryMapPresentation', () => {
  it('offers Browse as separate List and Map surfaces', () => {
    expect(buildBrowseSurfaceOptions()).toEqual([
      { label: 'List', value: 'list' },
      { label: 'Map', value: 'map' }
    ]);
  });

  it('builds horizontally navigable columns from the open containment path', () => {
    expect(buildInventoryMapColumns(map, ['garage', 'bin'])).toEqual([
      {
        key: 'root',
        level: 0,
        title: 'Home',
        assets: [asset('garage', undefined, 'Garage', 'location'), asset('flashlight', undefined, 'Flashlight', 'item')],
        emptyLabel: 'No active assets yet'
      },
      {
        key: 'garage',
        level: 1,
        title: 'Garage',
        parentId: 'garage',
        assets: [asset('bin', 'garage', 'Camping bin', 'container')],
        emptyLabel: 'Garage is empty'
      },
      {
        key: 'bin',
        level: 2,
        title: 'Camping bin',
        parentId: 'bin',
        assets: [asset('lantern', 'bin', 'Lantern', 'item')],
        emptyLabel: 'Camping bin is empty'
      }
    ]);
  });

  it('offers Add item here from empty containing columns only when spatial creation is permitted', () => {
    expect(buildInventoryMapEmptyColumnAction(asset('garage', undefined, 'Garage', 'location'))).toEqual({
      label: 'Add item here'
    });
    expect(buildInventoryMapEmptyColumnAction({
      ...asset('garage', undefined, 'Garage', 'location'),
      canAddContainedAssets: false
    })).toBeUndefined();
    expect(buildInventoryMapEmptyColumnAction(undefined)).toBeUndefined();
  });

  it('keeps breadcrumbs clickable and aligned to the open path', () => {
    expect(buildInventoryMapBreadcrumbs(map, ['garage', 'bin'])).toEqual([
      { key: 'root', level: 0, title: 'Home' },
      { key: 'garage', level: 1, title: 'Garage', assetId: 'garage' },
      { key: 'bin', level: 2, title: 'Camping bin', assetId: 'bin' }
    ]);
    expect(pathForBreadcrumbLevel(['garage', 'bin'], 1)).toEqual(['garage']);
    expect(pathForBreadcrumbLevel(['garage', 'bin'], 0)).toEqual([]);
  });

  it('selects only containing assets as branches', () => {
    expect(selectInventoryMapBranch(map, [], 'garage')).toEqual(['garage']);
    expect(selectInventoryMapBranch(map, ['garage'], 'bin')).toEqual(['garage', 'bin']);
    expect(selectInventoryMapBranch(map, ['garage'], 'lantern')).toEqual(['garage']);
  });

  it('finds a matching item and expands its parent path without replacing the map', () => {
    expect(findInventoryMapSearchMatch(map, 'lantern')).toEqual({
      assetId: 'lantern',
      openPath: ['garage', 'bin']
    });
    expect(findInventoryMapSearchMatch(map, 'camping')).toEqual({
      assetId: 'bin',
      openPath: ['garage', 'bin']
    });
  });

  it('summarizes the full selected inventory structure', () => {
    expect(mapOverviewLabel(map)).toBe('4 active assets · 2 root items');
  });

  it('marks every expanded branch row while preserving explicit search highlights', () => {
    expect(buildInventoryMapRowInteractionState(['garage', 'bin'], 'garage', undefined)).toEqual({
      expanded: true,
      highlighted: true
    });
    expect(buildInventoryMapRowInteractionState(['garage', 'bin'], 'lantern', 'lantern')).toEqual({
      expanded: false,
      highlighted: true
    });
    expect(buildInventoryMapRowInteractionState(['garage', 'bin'], 'flashlight', undefined)).toEqual({
      expanded: false,
      highlighted: false
    });
  });

  it('clears explicit highlights when breadcrumb navigation leaves that branch', () => {
    expect(preserveInventoryMapHighlightForPath(['garage'], 'garage')).toBe('garage');
    expect(preserveInventoryMapHighlightForPath(['garage'], 'bin')).toBeUndefined();
    expect(preserveInventoryMapHighlightForPath(['garage'], undefined)).toBeUndefined();
  });

  it('activates branch swipe only for leftward gestures on containing rows', () => {
    expect(shouldActivateInventoryMapBranchSwipe({ canContainAssets: true, dx: -4, dy: 1 })).toBe(true);
    expect(shouldActivateInventoryMapBranchSwipe({ canContainAssets: true, dx: -2, dy: 1 })).toBe(false);
    expect(shouldActivateInventoryMapBranchSwipe({ canContainAssets: false, dx: -4, dy: 1 })).toBe(false);
    expect(shouldActivateInventoryMapBranchSwipe({ canContainAssets: true, dx: 18, dy: 2 })).toBe(false);
    expect(shouldActivateInventoryMapBranchSwipe({ canContainAssets: true, dx: -12, dy: 20 })).toBe(false);
  });

  it('suppresses horizontal map scrolling only for active branch swipe gestures on containing rows', () => {
    expect(shouldSuppressInventoryMapScrollForBranchSwipe({ canContainAssets: true, dx: -4, dy: 1 })).toBe(true);
    expect(shouldSuppressInventoryMapScrollForBranchSwipe({ canContainAssets: false, dx: -4, dy: 1 })).toBe(false);
    expect(shouldSuppressInventoryMapScrollForBranchSwipe({ canContainAssets: true, dx: 20, dy: 2 })).toBe(false);
  });

  it('selects the branch during drag after the chevron reveal threshold', () => {
    expect(shouldSelectInventoryMapBranchDuringSwipe({ dx: -58 })).toBe(true);
    expect(shouldSelectInventoryMapBranchDuringSwipe({ dx: -72 })).toBe(true);
    expect(shouldSelectInventoryMapBranchDuringSwipe({ dx: -44 })).toBe(false);
  });

  it('clamps and snaps controlled map pager offsets to real columns', () => {
    expect(clampInventoryMapOffset({ offset: -20, maxOffset: 200 })).toBe(0);
    expect(clampInventoryMapOffset({ offset: 240, maxOffset: 200 })).toBe(200);
    expect(nearestInventoryMapColumnForOffset({ offset: 145, snapInterval: 100, maxLevel: 3 })).toBe(1);
    expect(nearestInventoryMapColumnForOffset({ offset: 180, snapInterval: 100, maxLevel: 3 })).toBe(2);
    expect(nearestInventoryMapColumnForOffset({ offset: 480, snapInterval: 100, maxLevel: 3 })).toBe(3);
  });

  it('converts continued branch pull into controlled horizontal pager progress', () => {
    expect(inventoryMapBranchSwipeOffset({
      dragX: inventoryMapGestureConfig.branchSwipeSelectDistance,
      fromLevel: 0,
      snapInterval: 320,
      maxLevel: 1
    })).toBe(0);
    expect(inventoryMapBranchSwipeOffset({
      dragX: -189,
      fromLevel: 0,
      snapInterval: 320,
      maxLevel: 1
    })).toBe(160);
    expect(inventoryMapBranchSwipeOffset({
      dragX: -360,
      fromLevel: 0,
      snapInterval: 320,
      maxLevel: 1
    })).toBe(320);
  });

  it('activates the controlled map pager only for deliberate horizontal pans', () => {
    expect(shouldActivateInventoryMapPagerPan({ dx: -12, dy: 2 })).toBe(true);
    expect(shouldActivateInventoryMapPagerPan({ dx: 12, dy: 2 })).toBe(true);
    expect(shouldActivateInventoryMapPagerPan({ dx: 8, dy: 0 })).toBe(false);
    expect(shouldActivateInventoryMapPagerPan({ dx: -12, dy: 14 })).toBe(false);
  });
});

const map: InventoryMapViewModel = {
  sessionScopeId: 'scope-one',
  tenantId: 'tenant-home' as InventoryMapViewModel['tenantId'],
  inventoryId: 'inventory-home' as InventoryMapViewModel['inventoryId'],
  inventoryName: 'Home',
  canCreateAsset: true,
  assets: [
    asset('garage', undefined, 'Garage', 'location'),
    asset('flashlight', undefined, 'Flashlight', 'item'),
    asset('bin', 'garage', 'Camping bin', 'container'),
    asset('lantern', 'bin', 'Lantern', 'item')
  ]
};

function asset(
  id: string,
  parentAssetId: string | undefined,
  title: string,
  kind: 'container' | 'item' | 'location'
) {
  return {
    id,
    parentAssetId,
    title,
    kind,
    kindLabel: kind === 'location' ? 'Location' : kind === 'container' ? 'Container' : 'Item',
    description: '',
    placementLabel: 'Home',
    parentPlacementLabel: parentAssetId ? 'Home' : 'Inventory root',
    updatedAtLabel: 'Updated today',
    childCount: id === 'garage' || id === 'bin' ? 1 : 0,
    childCountLabel: id === 'garage' || id === 'bin' ? '1 inside' : 'Empty',
    canContainAssets: kind !== 'item',
    canAddContainedAssets: kind !== 'item',
    imagePlaceholderLabel: kind === 'location' ? 'Place' : kind === 'container' ? 'Box' : 'Item',
    photoLabel: 'Needs photo',
    photo: undefined
  };
}
