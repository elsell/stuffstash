import { describe, expect, it } from 'vitest';
import {
  canUseContainedAssetAction,
  containedAssetActions,
  containedAssetsEmptyState
} from './ContainedAssetsPresentation';

describe('ContainedAssetsPresentation', () => {
  it('keeps spatial actions available for active editable containers with or without contents', () => {
    expect(containedAssetActions({
      canContainAssets: true,
      canAddContainedAssets: true
    })).toEqual([
      { kind: 'add_here', label: 'Add item here', isPrimary: true },
      { kind: 'move_here', label: 'Move things here', isPrimary: false }
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
      message: 'Add something here or move existing things into this place.'
    });
    expect(containedAssetsEmptyState({
      canAddContainedAssets: false
    })).toEqual({
      title: 'Nothing inside yet',
      message: 'This space is empty.'
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
});
