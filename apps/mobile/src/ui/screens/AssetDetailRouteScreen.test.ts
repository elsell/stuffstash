import { describe, expect, it } from 'vitest';
import { navigateAfterDeletedAsset } from './AssetDetailNavigation';
import {
  canSaveMoveAsset,
  movePlacementPreview,
  parentFromCurrentAssetPath
} from './AssetDetailMovePresentation';
import { isCurrentAuditHistoryRequest } from './AssetAuditHistoryPresentation';
import { assetPhotoViewerModel } from './AssetPhotoWorkspacePresentation';
import {
  assetLifecycleActionRows,
  assetLifecycleConfirmation
} from './AssetLifecyclePresentation';
import {
  assetEditContext,
  canSaveEditAsset,
  hasDirtyEditAssetDraft,
  normalizedEditDraft
} from './AssetDetailEditPresentation';

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
      pathLabel: 'Kitchen',
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
      parentLocationTrailLabel: 'Kitchen / Big cabinet',
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
      title: 'Kitchen / Big cabinet',
      pathLabel: 'Kitchen / Big cabinet'
    });
  });

  it('summarizes current and proposed placement without repeating the moved asset title', () => {
    expect(movePlacementPreview({
      parentAssetId: 'asset-cabinet',
      parentLocationTrailLabel: 'Kitchen / Big cabinet'
    }, {
      id: 'asset-shelf',
      title: 'Second shelf',
      kind: 'container',
      subtitle: 'Kitchen / Big cabinet',
      pathLabel: 'Kitchen / Big cabinet / Second shelf',
      selectionHint: 'Container',
      willPromoteToContainer: false
    })).toEqual({
      currentLocationLabel: 'Kitchen / Big cabinet',
      proposedLocationLabel: 'Kitchen / Big cabinet / Second shelf',
      hasChanged: true
    });
  });

  it('labels root placement clearly', () => {
    expect(movePlacementPreview({
      parentAssetId: undefined,
      parentLocationTrailLabel: 'Inventory root'
    }, null)).toEqual({
      currentLocationLabel: 'Inventory root',
      proposedLocationLabel: 'Inventory root',
      hasChanged: false
    });
  });

  it('uses structured path labels instead of parsing slashes from titles', () => {
    expect(movePlacementPreview({
      parentAssetId: 'asset-cabinet',
      parentLocationTrailLabel: 'Kitchen / AC/DC bin'
    }, {
      id: 'asset-pantry',
      title: 'Pantry / dry goods',
      kind: 'location',
      subtitle: 'New location',
      pathLabel: 'Pantry / dry goods',
      selectionHint: 'Location',
      willPromoteToContainer: false
    })).toEqual({
      currentLocationLabel: 'Kitchen / AC/DC bin',
      proposedLocationLabel: 'Pantry / dry goods',
      hasChanged: true
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

describe('asset lifecycle presentation helpers', () => {
  it('shows active and archived lifecycle actions with destructive delete separated', () => {
    expect(assetLifecycleActionRows({
      canArchive: true,
      canRestore: false,
      canDeletePermanently: false
    })).toEqual([
      { kind: 'archive', label: 'Archive', isDestructive: false }
    ]);

    expect(assetLifecycleActionRows({
      canArchive: false,
      canRestore: true,
      canDeletePermanently: true
    })).toEqual([
      { kind: 'restore', label: 'Restore', isDestructive: false },
      { kind: 'delete', label: 'Delete permanently', isDestructive: true }
    ]);
  });

  it('names the asset and removed media in permanent delete confirmation copy', () => {
    expect(assetLifecycleConfirmation('delete', {
      title: 'Passport folder',
      photos: [
        { id: 'photo-one', label: 'photo-one.jpg', uri: 'https://photos/one.jpg' },
        { id: 'photo-two', label: 'photo-two.jpg', uri: 'https://photos/two.jpg' }
      ],
      containedAssetsLabel: '1 thing inside',
      canContainAssets: true
    })).toEqual({
      title: 'Delete Passport folder permanently?',
      message: 'This permanently removes Passport folder. 2 photos will be removed with it. Current contents: 1 thing inside. Deletion will not continue while active things are inside it. Audit history remains, but the asset itself cannot be restored.',
      confirmLabel: 'Delete permanently',
      isDestructive: true
    });
  });

  it('explains archive and restore without treating them like permanent delete', () => {
    expect(assetLifecycleConfirmation('archive', {
      title: 'Water bottle',
      photos: [],
      containedAssetsLabel: '0 things inside',
      canContainAssets: false
    })).toEqual({
      title: 'Archive Water bottle?',
      message: 'Water bottle will be hidden from normal inventory work. You can restore it later from archived asset views.',
      confirmLabel: 'Archive',
      isDestructive: false
    });
    expect(assetLifecycleConfirmation('restore', {
      title: 'Water bottle',
      photos: [],
      containedAssetsLabel: '0 things inside',
      canContainAssets: false
    })).toEqual({
      title: 'Restore Water bottle?',
      message: 'Water bottle will return to active inventory work.',
      confirmLabel: 'Restore',
      isDestructive: false
    });
  });
});

describe('asset edit presentation helpers', () => {
  it('saves only meaningful dirty edits with a non-empty title', () => {
    const asset = { title: 'Water bottle', description: 'On the desk.' };

    expect(canSaveEditAsset(asset, undefined)).toBe(false);
    expect(canSaveEditAsset(asset, { title: '   ', description: 'On the desk.' })).toBe(false);
    expect(canSaveEditAsset(asset, { title: 'Water bottle', description: 'On the desk.' })).toBe(false);
    expect(canSaveEditAsset(asset, { title: '  Water bottle  ', description: '  On the desk.  ' })).toBe(false);
    expect(canSaveEditAsset(asset, { title: 'Water bottle', description: 'On the shelf.' })).toBe(true);
    expect(canSaveEditAsset(asset, { title: 'Big water bottle', description: 'On the desk.' })).toBe(true);
  });

  it('uses the same normalized edit state for discard warnings and save payloads', () => {
    const asset = { title: 'Water bottle', description: 'On the desk.' };
    const whitespaceOnly = { title: '  Water bottle  ', description: '  On the desk.  ' };
    const changed = { title: '  Water bottle  ', description: '  On the shelf.  ' };

    expect(hasDirtyEditAssetDraft(asset, whitespaceOnly)).toBe(false);
    expect(hasDirtyEditAssetDraft(asset, changed)).toBe(true);
    expect(normalizedEditDraft(changed)).toEqual({
      title: 'Water bottle',
      description: 'On the shelf.'
    });
  });

  it('shows kind and custom type as read-only edit context', () => {
    expect(assetEditContext({
      kindLabel: 'Container',
      customTypeLabel: 'Documents'
    })).toEqual({
      kindLabel: 'Container',
      customTypeLabel: 'Documents',
      helperText: 'Kind and type changes need a future conversion flow.'
    });
  });
});
