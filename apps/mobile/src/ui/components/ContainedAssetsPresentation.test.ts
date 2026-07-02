import { describe, expect, it } from 'vitest';
import {
  canUseContainedAssetAction,
  containedAssetActions,
  containedAssetsEmptyState,
  containedAssetRows,
  containedAssetsSectionHeading
} from './ContainedAssetsPresentation';

describe('ContainedAssetsPresentation', () => {
  it('keeps spatial actions available for active editable containers with or without contents', () => {
    expect(containedAssetActions({
      canContainAssets: true,
      canAddContainedAssets: true
    })).toEqual([
      { kind: 'add_here', label: 'Add item here', isPrimary: true },
      { kind: 'move_here', label: 'Move items here', isPrimary: false }
    ]);
  });

  it('hides spatial actions when the asset cannot accept contained assets', () => {
    expect(containedAssetActions({
      canContainAssets: false,
      canAddContainedAssets: true
    })).toEqual([]);
    expect(containedAssetActions({
      canContainAssets: true,
      canAddContainedAssets: false
    })).toEqual([]);
  });

  it('uses action-oriented empty copy only when spatial actions are available', () => {
    expect(containedAssetsEmptyState({
      canAddContainedAssets: true
    })).toEqual({
      title: 'Nothing inside yet',
      message: 'Add an item here or move items into this space.'
    });
    expect(containedAssetsEmptyState({
      canAddContainedAssets: false
    })).toEqual({
      title: 'Nothing inside yet',
      message: 'This space is empty.'
    });
  });

  it('anchors the contained section to the current asset', () => {
    expect(containedAssetsSectionHeading({
      title: 'Garage cabinet',
      containedAssetsLabel: '4 things inside'
    })).toEqual({
      title: 'Inside Garage cabinet',
      summary: '4 things inside'
    });
  });

  it('disables contained actions while another asset action is pending', () => {
    const onPress = () => undefined;

    expect(canUseContainedAssetAction({
      isActionPending: false,
      onPress
    })).toBe(true);
    expect(canUseContainedAssetAction({
      isActionPending: true,
      onPress
    })).toBe(false);
    expect(canUseContainedAssetAction({
      isActionPending: false,
      onPress: undefined
    })).toBe(false);
  });

  it('maps the application-sorted contained rows without adding parent path metadata', () => {
    expect(containedAssetRows([{
      id: 'asset-battery-bin',
      title: 'Battery bin',
      kindLabel: 'Container',
      customTypeLabel: undefined,
      description: 'AA, AAA, and coin cells.',
      locationTrailLabel: 'Garage / Shelf / Battery bin',
      updatedAtLabel: 'Updated yesterday',
      photoLabel: 'Photo ready',
      imagePlaceholderLabel: 'Box',
      photo: { uri: 'https://photos/battery-bin.jpg' }
    }])).toEqual([{
      id: 'asset-battery-bin',
      title: 'Battery bin',
      eyebrowLabel: 'Container',
      supportingLabel: 'AA, AAA, and coin cells.',
      imagePlaceholderLabel: 'Box',
      photo: { uri: 'https://photos/battery-bin.jpg' }
    }]);
  });

  it('falls back to kind and photo readiness when a contained row has no description or type', () => {
    expect(containedAssetRows([{
      id: 'asset-spare-key',
      title: 'Spare key',
      kindLabel: 'Item',
      customTypeLabel: undefined,
      description: '',
      locationTrailLabel: 'Garage / Shelf / Spare key',
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      imagePlaceholderLabel: 'Item',
      photo: undefined
    }])).toEqual([{
      id: 'asset-spare-key',
      title: 'Spare key',
      eyebrowLabel: 'Item',
      supportingLabel: 'Needs photo',
      imagePlaceholderLabel: 'Item',
      photo: undefined
    }]);
  });
});
