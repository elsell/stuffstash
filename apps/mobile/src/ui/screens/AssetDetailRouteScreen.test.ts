import { describe, expect, it } from 'vitest';
import { navigateAfterDeletedAsset } from './AssetDetailNavigation';
import { canSaveMoveAsset, parentFromCurrentAssetPath } from './AssetDetailMovePresentation';
import { isCurrentAuditHistoryRequest } from './AssetAuditHistoryPresentation';
import { assetPhotoViewerModel } from './AssetPhotoWorkspacePresentation';

describe('navigateAfterDeletedAsset', () => {
  it('uses native back navigation when the asset detail route has history', () => {
    const calls: string[] = [];

    navigateAfterDeletedAsset({
      canGoBack: () => true,
      back: () => {
        calls.push('back');
      },
      replace: (href) => {
        calls.push(`replace:${href.toString()}`);
      }
    });

    expect(calls).toEqual(['back']);
  });

  it('replaces with Home when the deleted asset route has no back stack', () => {
    const calls: string[] = [];

    navigateAfterDeletedAsset({
      canGoBack: () => false,
      back: () => {
        calls.push('back');
      },
      replace: (href) => {
        calls.push(`replace:${href.toString()}`);
      }
    });

    expect(calls).toEqual(['replace:/']);
  });
});

describe('asset detail move helpers', () => {
  it('does not allow saving a move until the destination differs', () => {
    expect(canSaveMoveAsset({ parentAssetId: 'asset-kitchen' }, {
      id: 'asset-kitchen',
      title: 'Kitchen',
      kind: 'location',
      subtitle: 'Current parent',
      selectionHint: 'Location',
      willPromoteToContainer: false
    })).toBe(false);
    expect(canSaveMoveAsset({ parentAssetId: 'asset-kitchen' }, null)).toBe(true);
    expect(canSaveMoveAsset({ parentAssetId: undefined }, null)).toBe(false);
  });

  it('builds a current parent fallback from the visible location path', () => {
    expect(parentFromCurrentAssetPath({
      id: 'asset-mug',
      title: 'Mug',
      kind: 'item',
      kindLabel: 'Item',
      description: '',
      parentAssetId: 'asset-cabinet',
      locationTrailLabel: 'Kitchen / Big cabinet / Mug',
      lifecycleLabel: 'Active',
      isActive: true,
      canEdit: true,
      canMove: true,
      canAddPhotos: true,
      canArchive: true,
      canRestore: false,
      canDeletePermanently: false,
      containedAssets: [],
      containedAssetsLabel: '0 things inside',
      canContainAssets: false,
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      imagePlaceholderLabel: 'Item'
    })).toMatchObject({
      id: 'asset-cabinet',
      title: 'Big cabinet'
    });
  });
});

describe('asset audit history presentation helpers', () => {
  it('rejects stale audit history requests after close or navigation', () => {
    expect(isCurrentAuditHistoryRequest(3, 3)).toBe(true);
    expect(isCurrentAuditHistoryRequest(4, 3)).toBe(false);
  });
});

describe('asset photo workspace presentation helpers', () => {
  it('builds viewer navigation around the selected photo', () => {
    expect(assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', uri: 'https://photos/one.jpg' },
      { id: 'photo-two', label: 'two.jpg', uri: 'https://photos/two.jpg' },
      { id: 'photo-three', label: 'three.jpg', uri: 'https://photos/three.jpg' }
    ], 'photo-two')).toMatchObject({
      positionLabel: '2 of 3',
      previousPhotoId: 'photo-one',
      nextPhotoId: 'photo-three',
      photo: {
        id: 'photo-two',
        label: 'two.jpg'
      }
    });
  });

  it('does not show a stale selected photo after refresh', () => {
    expect(assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', uri: 'https://photos/one.jpg' }
    ], 'photo-two')).toBeUndefined();
  });
});
