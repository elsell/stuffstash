import { describe, expect, it } from 'vitest';
import {
  assetDetailHref,
  locationAssetDetailHref,
  navigateAfterDeletedAsset
} from './AssetDetailNavigation';
import { addHereParams, addHereRouteParams } from './AddAssetInitialParent';
import {
  consumeAssetActionCompletion,
  recordAssetActionCompletion
} from './AssetActionCompletion';
import {
  assetAuditNativeSheetOptions,
  assetEditNativeSheetOptions,
  assetMoveHereNativeSheetOptions,
  assetMoveNativeSheetOptions
} from './AssetNativeSheetOptions';
import {
  canCreateMoveDestination,
  canSaveMoveAsset,
  createdMoveDestinationParent,
  isSelectableMoveDestination,
  isSelectableMoveIntoCandidate,
  moveIntoEmptyState,
  moveIntoCandidateRow,
  moveDestinationRow,
  moveDestinationCreateInput,
  moveDestinationCreateButtonLabel,
  moveDestinationCreateKindHelp,
  moveDestinationCreateKindLabel,
  moveDestinationCreatePlacement,
  moveDestinationCreatePlacementLabel,
  movePlacementPreview,
  parentFromCurrentAssetPath
} from './AssetDetailMovePresentation';
import { isCurrentAuditHistoryRequest } from './AssetAuditHistoryPresentation';
import {
  assetPhotoViewerModel,
  assetPhotoViewerModelAtIndex,
  assetPhotoViewerControls,
  assetPhotoStatusLabel,
  isAssetPhotoId,
  selectedAssetPhotoViewerIndex
} from '../components/AssetPhotoWorkspacePresentation';
import {
  applyPhotoUploadProgress,
  photoUploadRows
} from './AssetPhotoUploadProgressPresentation';
import {
  assetLifecycleActionRows,
  assetLifecycleConfirmation,
  assetLifecycleFailurePresentation,
  assetLifecycleOverflowMenu,
  assetLifecycleOverflowPresentation
} from './AssetLifecyclePresentation';
import {
  assetWorkspaceSuccessStatus,
  assetWorkspaceWorkingStatus,
  visibleAssetWorkspaceStatus
} from './AssetWorkspaceStatusPresentation';
import {
  assetEditContext,
  canSaveEditAsset,
  hasDirtyEditAssetDraft,
  normalizedEditDraft
} from './AssetDetailEditPresentation';

describe('navigateAfterDeletedAsset', () => {
  it('builds explicit asset detail route params for card navigation', () => {
    expect(assetDetailHref('asset-water-bottle')).toEqual({
      pathname: '/assets/[assetId]',
      params: { assetId: 'asset-water-bottle' }
    });
    expect(locationAssetDetailHref('location-garage', 'asset-water-bottle')).toEqual({
      pathname: '/locations/[locationId]/assets/[assetId]',
      params: {
        assetId: 'asset-water-bottle',
        locationId: 'location-garage'
      }
    });
  });

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

describe('addHereParams', () => {
  it('builds route params that preselect the current container as the Add parent', () => {
    expect(addHereParams({
      id: 'asset-tv-box',
      title: 'TV box',
      kind: 'container',
      kindLabel: 'Container',
      description: '',
      parentAssetId: 'asset-living-room',
      locationTrailLabel: 'Living room / TV box',
      parentLocationTrailLabel: 'Living room',
      lifecycleLabel: 'Active',
      isActive: true,
      canEdit: true,
      canMove: true,
      canAddPhotos: true,
      canArchive: true,
      canRestore: false,
      canDeletePermanently: false,
      isCheckedOut: false,
      checkoutLabel: 'Available',
      canCheckout: true,
      canReturn: false,
      containedAssets: [],
      containedAssetsLabel: '0 things inside',
      canContainAssets: true,
      canAddContainedAssets: true,
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      imagePlaceholderLabel: 'Box'
    })).toEqual({
      parentAssetId: 'asset-tv-box',
      parentTitle: 'TV box',
      parentKind: 'container',
      parentSubtitle: 'Living room',
      parentPathLabel: 'Living room / TV box',
      parentSelectionHint: 'Container',
      parentWillPromoteToContainer: 'false'
    });
  });

  it('builds the same Add parent route params from a map row source', () => {
    expect(addHereRouteParams({
      id: 'asset-garage-shelf',
      title: 'Garage shelf',
      kind: 'location',
      kindLabel: 'Location',
      parentLocationTrailLabel: 'Garage',
      locationTrailLabel: 'Garage / Garage shelf'
    })).toEqual({
      parentAssetId: 'asset-garage-shelf',
      parentTitle: 'Garage shelf',
      parentKind: 'location',
      parentSubtitle: 'Garage',
      parentPathLabel: 'Garage / Garage shelf',
      parentSelectionHint: 'Location',
      parentWillPromoteToContainer: 'false'
    });
  });
});

describe('asset native sheet route options', () => {
  it('uses stack-native form sheets with grabbers and detents for asset actions', () => {
    for (const options of [
      assetEditNativeSheetOptions,
      assetMoveNativeSheetOptions,
      assetMoveHereNativeSheetOptions,
      assetAuditNativeSheetOptions
    ]) {
      expect(options.presentation).toBe('formSheet');
      expect(options.headerShown).toBe(false);
      expect(options.sheetGrabberVisible).toBe(true);
      expect(options.sheetExpandsWhenScrolledToEdge).toBe(true);
      expect(options.sheetLargestUndimmedDetentIndex).toBe('none');
      expect(options.sheetAllowedDetents.length).toBeGreaterThan(1);
    }
  });

  it('keeps edit cancellation explicit until dirty native sheet dismissal can be intercepted', () => {
    expect(assetEditNativeSheetOptions.gestureEnabled).toBe(false);
    expect(assetMoveNativeSheetOptions.gestureEnabled).toBeUndefined();
    expect(assetMoveHereNativeSheetOptions.gestureEnabled).toBeUndefined();
    expect(assetAuditNativeSheetOptions.gestureEnabled).toBeUndefined();
  });
});

describe('asset action completion handoff', () => {
  it('consumes one native sheet completion once for the asset detail refresh', () => {
    recordAssetActionCompletion({
      assetId: 'asset-water-bottle',
      action: 'move',
      message: 'Moved Water bottle.'
    });

    expect(consumeAssetActionCompletion('asset-water-bottle')).toEqual({
      assetId: 'asset-water-bottle',
      action: 'move',
      message: 'Moved Water bottle.'
    });
    expect(consumeAssetActionCompletion('asset-water-bottle')).toBeUndefined();
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
      isCheckedOut: false,
      checkoutLabel: 'Available',
      canCheckout: true,
      canReturn: false,
      containedAssets: [],
      containedAssetsLabel: '0 things inside',
      canContainAssets: false,
      canAddContainedAssets: false,
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

  it('builds move destination rows with path breadcrumbs and kind hints', () => {
    expect(moveDestinationRow({
      id: 'asset-shelf',
      title: 'Second shelf',
      kind: 'container',
      subtitle: 'Kitchen / Big cabinet',
      pathLabel: 'Kitchen / Big cabinet / Second shelf',
      selectionHint: 'Container',
      willPromoteToContainer: false
    })).toEqual({
      title: 'Second shelf',
      kindLabel: 'Container',
      pathLabel: 'Kitchen / Big cabinet / Second shelf'
    });
  });

  it('does not treat item assets as selectable move destinations until promotion is implemented', () => {
    const itemCandidate = {
      id: 'asset-suitcase',
      title: 'Suitcase',
      kind: 'item',
      subtitle: 'Bedroom closet',
      pathLabel: 'Bedroom closet / Suitcase',
      selectionHint: 'Item',
      willPromoteToContainer: true
    } as const;

    expect(isSelectableMoveDestination(itemCandidate)).toBe(false);
    expect(isSelectableMoveDestination({
      ...itemCandidate,
      kind: 'container',
      selectionHint: 'Container',
      willPromoteToContainer: false
    })).toBe(true);
  });

  it('builds move-here candidate rows without destination promotion copy', () => {
    expect(moveIntoCandidateRow({
      id: 'asset-suitcase',
      title: 'Suitcase',
      kind: 'item',
      subtitle: 'Bedroom closet',
      pathLabel: 'Bedroom closet / Suitcase',
      selectionHint: 'Item',
      willPromoteToContainer: true
    })).toEqual({
      title: 'Suitcase',
      kindLabel: 'Item',
      pathLabel: 'Bedroom closet / Suitcase'
    });
  });

  it('hides the current target and existing direct children from move-here candidates', () => {
    const target = {
      id: 'asset-cabinet',
      containedAssets: [
        {
          id: 'asset-mug',
          title: 'Mug',
          kindLabel: 'Item',
          description: '',
          locationTrailLabel: 'Kitchen / Cabinet / Mug',
          updatedAtLabel: 'Updated today',
          photoLabel: 'Needs photo',
          imagePlaceholderLabel: 'Item'
        }
      ]
    };

    expect(isSelectableMoveIntoCandidate({ id: 'asset-cabinet' }, target)).toBe(false);
    expect(isSelectableMoveIntoCandidate({ id: 'asset-mug' }, target)).toBe(false);
    expect(isSelectableMoveIntoCandidate({
      id: 'asset-hidden-direct-child',
      parentAssetId: 'asset-cabinet'
    }, target)).toBe(false);
    expect(isSelectableMoveIntoCandidate({ id: 'asset-spoon' }, target)).toBe(true);
  });

  it('does not claim no move-here matches before the user searches', () => {
    expect(moveIntoEmptyState('')).toEqual({
      title: 'Search for something to move here',
      message: 'Start typing to find an item, box, or place from this inventory.'
    });
    expect(moveIntoEmptyState('spoon')).toEqual({
      title: 'No movable matches',
      message: 'Search for something that is not already inside this place.'
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

  it('labels inline destination creation by the selected asset kind', () => {
    expect(moveDestinationCreateKindLabel('location')).toBe('Location');
    expect(moveDestinationCreateKindLabel('container')).toBe('Container');
    expect(moveDestinationCreateKindHelp('location')).toBe('Best for rooms, places, and areas.');
    expect(moveDestinationCreateKindHelp('container')).toBe('Best for boxes, shelves, bins, and cabinets.');
    expect(moveDestinationCreateButtonLabel('container', 'big cabinet')).toBe('Create container "big cabinet"');
    expect(moveDestinationCreatePlacementLabel({
      parentAssetId: 'asset-kitchen',
      parentPathLabel: 'Kitchen'
    })).toBe('Creates inside Kitchen');
    expect(moveDestinationCreatePlacementLabel({})).toBe('Creates at inventory root');
  });

  it('builds a selectable parent from a newly created move destination', () => {
    const placement = {
      parentAssetId: 'asset-kitchen',
      parentPathLabel: 'Kitchen'
    };

    expect(createdMoveDestinationParent({
      id: 'asset-cabinet',
      kind: 'container',
      placement,
      title: 'big cabinet'
    })).toEqual({
      id: 'asset-cabinet',
      title: 'big cabinet',
      kind: 'container',
      parentAssetId: 'asset-kitchen',
      subtitle: 'New container',
      pathLabel: 'Kitchen / big cabinet',
      selectionHint: 'Container',
      willPromoteToContainer: false
    });

    expect(createdMoveDestinationParent({
      id: 'asset-kitchen',
      kind: 'location',
      placement: {},
      title: 'Kitchen'
    })).toEqual({
      id: 'asset-kitchen',
      title: 'Kitchen',
      kind: 'location',
      parentAssetId: undefined,
      subtitle: 'New location',
      pathLabel: 'Kitchen',
      selectionHint: 'Location',
      willPromoteToContainer: false
    });
  });

  it('builds the create command input from the selected destination kind', () => {
    expect(moveDestinationCreateInput('container', '  big cabinet  ', {
      parentAssetId: 'asset-kitchen',
      parentPathLabel: 'Kitchen'
    })).toEqual({
      kind: 'container',
      title: 'big cabinet',
      description: '',
      parentAssetId: 'asset-kitchen'
    });
    expect(moveDestinationCreateInput('location', 'Kitchen', {})).toEqual({
      kind: 'location',
      title: 'Kitchen',
      description: ''
    });
  });

  it('derives inline destination creation placement from the moved asset current parent', () => {
    expect(moveDestinationCreatePlacement({
      parentAssetId: 'asset-kitchen',
      parentLocationTrailLabel: 'Kitchen'
    })).toEqual({
      parentAssetId: 'asset-kitchen',
      parentPathLabel: 'Kitchen'
    });
    expect(moveDestinationCreatePlacement({
      parentAssetId: undefined,
      parentLocationTrailLabel: 'Inventory root'
    })).toEqual({});
  });

  it('allows inline creation when an exact title exists only for another destination kind', () => {
    const matches = [
      {
        title: 'Cabinet',
        kind: 'location',
        parentAssetId: 'asset-kitchen'
      },
      {
        title: 'Shelf',
        kind: 'container',
        parentAssetId: 'asset-kitchen'
      }
    ] as const;

    expect(canCreateMoveDestination({
      kind: 'container',
      matches,
      parentAssetId: 'asset-kitchen',
      query: 'Cabinet'
    })).toBe(true);
    expect(canCreateMoveDestination({
      kind: 'location',
      matches,
      parentAssetId: 'asset-kitchen',
      query: 'cabinet'
    })).toBe(false);
    expect(canCreateMoveDestination({
      kind: 'container',
      matches,
      parentAssetId: 'asset-kitchen',
      query: '   '
    })).toBe(false);
    expect(canCreateMoveDestination({
      kind: 'container',
      matches,
      parentAssetId: 'asset-office',
      query: 'Shelf'
    })).toBe(true);
  });
});

describe('asset workspace status presentation', () => {
  it('describes in-flight workspace mutations near the primary actions', () => {
    expect(assetWorkspaceWorkingStatus('edit')).toEqual({
      kind: 'working',
      message: 'Saving changes...'
    });
    expect(assetWorkspaceWorkingStatus('move')).toEqual({
      kind: 'working',
      message: 'Moving asset...'
    });
    expect(assetWorkspaceWorkingStatus('archive')).toEqual({
      kind: 'working',
      message: 'Archiving asset...'
    });
  });

  it('builds concise success copy from command results and the loaded asset', () => {
    expect(assetWorkspaceSuccessStatus('edit', { message: 'Updated Water bottle.' })).toEqual({
      kind: 'success',
      message: 'Updated Water bottle.'
    });
    expect(assetWorkspaceSuccessStatus('move', { message: 'Moved Water bottle.' })).toEqual({
      kind: 'success',
      message: 'Moved Water bottle.'
    });
    expect(assetWorkspaceSuccessStatus('archive', {
      id: 'asset-water-bottle',
      title: 'Water bottle',
      kind: 'item',
      kindLabel: 'Item',
      description: '',
      parentAssetId: 'asset-office',
      locationTrailLabel: 'Office / Water bottle',
      parentLocationTrailLabel: 'Office',
      lifecycleLabel: 'Active',
      isActive: true,
      canEdit: true,
      canMove: true,
      canAddPhotos: true,
      canArchive: true,
      canRestore: false,
      canDeletePermanently: false,
      isCheckedOut: false,
      checkoutLabel: 'Available',
      canCheckout: true,
      canReturn: false,
      containedAssets: [],
      containedAssetsLabel: '0 things inside',
      canContainAssets: false,
      canAddContainedAssets: false,
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      imagePlaceholderLabel: 'Item'
    })).toEqual({
      kind: 'success',
      message: 'Archived Water bottle.'
    });
  });

  it('keeps photo upload progress separate from workspace mutation status', () => {
    const currentStatus = { kind: 'success' as const, message: 'Moved Water bottle.' };

    expect(visibleAssetWorkspaceStatus('move', currentStatus)).toEqual({
      kind: 'working',
      message: 'Moving asset...'
    });
    expect(visibleAssetWorkspaceStatus('photos', currentStatus)).toBe(currentStatus);
    expect(visibleAssetWorkspaceStatus(undefined, currentStatus)).toBe(currentStatus);
  });
});

describe('asset audit history presentation helpers', () => {
  it('rejects stale audit history requests after close or navigation', () => {
    expect(isCurrentAuditHistoryRequest(3, 3)).toBe(true);
    expect(isCurrentAuditHistoryRequest(4, 3)).toBe(false);
  });
});

describe('asset photo workspace presentation helpers', () => {
  it('builds upload progress rows from selected photo order', () => {
    expect(photoUploadRows([
      { id: 'selected-one', fileName: 'one.jpg', contentType: 'image/jpeg', uri: 'file://one', sizeBytes: 1 },
      { id: 'selected-two', fileName: 'two.png', contentType: 'image/png', uri: 'file://two', sizeBytes: 2 }
    ])).toEqual([
      { index: 0, fileName: 'one.jpg', status: 'pending' },
      { index: 1, fileName: 'two.png', status: 'pending' }
    ]);
  });

  it('updates only the upload progress row that matches index and file name', () => {
    const rows = photoUploadRows([
      { id: 'selected-one', fileName: 'one.jpg', contentType: 'image/jpeg', uri: 'file://one', sizeBytes: 1 },
      { id: 'selected-two', fileName: 'two.png', contentType: 'image/png', uri: 'file://two', sizeBytes: 2 }
    ]);

    expect(applyPhotoUploadProgress(rows, {
      index: 1,
      fileName: 'two.png',
      status: 'uploading'
    })).toEqual([
      { index: 0, fileName: 'one.jpg', status: 'pending' },
      { index: 1, fileName: 'two.png', status: 'uploading' }
    ]);
    expect(applyPhotoUploadProgress(rows, {
      index: 1,
      fileName: 'stale-selection.png',
      status: 'failed'
    })).toBe(rows);
  });

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

  it('builds explicit viewer control state from current photo position', () => {
    const model = assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', uri: 'https://photos/one.jpg' },
      {
        id: 'photo-two',
        label: 'fallback label',
        fileName: 'two.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 1536000,
        uri: 'https://photos/two.jpg'
      },
      { id: 'photo-three', label: 'three.jpg', uri: 'https://photos/three.jpg' }
    ], 'photo-two');

    expect(assetPhotoViewerControls(model, true)).toEqual({
      canGoPrevious: true,
      canGoNext: true,
      canRemove: true,
      fileLabel: 'two.jpg',
      metadataLabel: 'JPEG image · 1.5 MB',
      positionLabel: '2 of 3'
    });
    expect(assetPhotoViewerControls(model, false)).toMatchObject({
      canRemove: false
    });
    expect(assetPhotoViewerControls(undefined, true)).toEqual({
      canGoPrevious: false,
      canGoNext: false,
      canRemove: false,
      fileLabel: 'Photo',
      metadataLabel: undefined,
      positionLabel: '0 of 0'
    });
  });

  it('formats only safe image metadata for the photo viewer', () => {
    expect(assetPhotoViewerControls(assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', contentType: 'image/png', sizeBytes: 512, uri: 'https://photos/one.jpg' }
    ], 'photo-one'), true).metadataLabel).toBe('PNG image · 512 B');
    expect(assetPhotoViewerControls(assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', contentType: 'image/webp', sizeBytes: 2048, uri: 'https://photos/one.jpg' }
    ], 'photo-one'), true).metadataLabel).toBe('WebP image · 2.0 KB');
    expect(assetPhotoViewerControls(assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', contentType: 'image/jpeg;private=x', sizeBytes: 0, uri: 'https://photos/one.jpg' }
    ], 'photo-one'), true).metadataLabel).toBeUndefined();
    expect(assetPhotoViewerControls(assetPhotoViewerModel([
      { id: 'photo-one', label: 'one.jpg', contentType: 'application/pdf', sizeBytes: -1, uri: 'https://photos/one.jpg' }
    ], 'photo-one'), true).metadataLabel).toBeUndefined();
  });

  it('builds viewer control state from an image viewer index', () => {
    const photos = [
      { id: 'photo-one', label: 'one.jpg', uri: 'https://photos/one.jpg' },
      { id: 'photo-two', label: 'two.jpg', uri: 'https://photos/two.jpg' },
      { id: 'photo-three', label: 'three.jpg', uri: 'https://photos/three.jpg' }
    ];

    expect(assetPhotoViewerControls(assetPhotoViewerModelAtIndex(photos, 1), true)).toEqual({
      canGoPrevious: true,
      canGoNext: true,
      canRemove: true,
      fileLabel: 'two.jpg',
      metadataLabel: undefined,
      positionLabel: '2 of 3'
    });
    expect(assetPhotoViewerModelAtIndex(photos, 99)).toBeUndefined();
  });

  it('distinguishes photo ids from asset ids before opening the photo viewer', () => {
    const photos = [
      { id: 'photo-one', label: 'one.jpg', uri: 'https://photos/one.jpg' },
      { id: 'photo-two', label: 'two.jpg', uri: 'https://photos/two.jpg' }
    ];

    expect(isAssetPhotoId(photos, 'photo-one')).toBe(true);
    expect(isAssetPhotoId(photos, 'asset-contained-item')).toBe(false);
  });

  it('does not mount the photo viewer until a current asset photo is explicitly selected', () => {
    const photos = [
      { id: 'photo-one', label: 'one.jpg', uri: 'https://photos/one.jpg' },
      { id: 'photo-two', label: 'two.jpg', uri: 'https://photos/two.jpg' }
    ];

    expect(selectedAssetPhotoViewerIndex(photos, undefined)).toBeUndefined();
    expect(selectedAssetPhotoViewerIndex(photos, {
      photo: { id: 'photo-missing', label: 'missing.jpg', uri: 'https://photos/missing.jpg' },
      positionLabel: '1 of 1'
    })).toBeUndefined();
    expect(selectedAssetPhotoViewerIndex(photos, assetPhotoViewerModel(photos, 'photo-two'))).toBe(1);
  });

  it('labels the first visible photo without implying unsaved local ordering', () => {
    expect(assetPhotoStatusLabel({
      index: 0,
      label: 'one.jpg'
    })).toBe('First photo');
    expect(assetPhotoStatusLabel({
      index: 1,
      label: 'two.jpg'
    })).toBe('two.jpg');
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

  it('names the current asset in the lifecycle overflow context', () => {
    expect(assetLifecycleOverflowPresentation({
      title: 'Water bottle',
      lifecycleLabel: 'Active'
    })).toEqual({
      title: 'Water bottle actions',
      message: 'Lifecycle actions for this active asset.'
    });
  });

  it('builds native lifecycle overflow menu ordering and destructive index', () => {
    expect(assetLifecycleOverflowMenu({
      title: 'Tool box',
      lifecycleLabel: 'Archived',
      canArchive: false,
      canRestore: true,
      canDeletePermanently: true
    })).toEqual({
      title: 'Tool box actions',
      message: 'Lifecycle actions for this archived asset.',
      actionRows: [
        { kind: 'restore', label: 'Restore', isDestructive: false },
        { kind: 'delete', label: 'Delete permanently', isDestructive: true }
      ],
      options: ['Restore', 'Delete permanently', 'Checkout history', 'Audit history', 'Cancel'],
      checkoutHistoryIndex: 2,
      auditIndex: 3,
      cancelIndex: 4,
      destructiveIndex: 1
    });
  });

  it('turns lifecycle failures into action-specific recovery copy', () => {
    expect(assetLifecycleFailurePresentation('archive', {
      title: 'Tool box',
      canContainAssets: true
    }, 'Asset has active children.')).toEqual({
      title: 'Could not archive Tool box',
      message: 'Asset has active children. Move or archive active things inside this asset, then try again.'
    });

    expect(assetLifecycleFailurePresentation('restore', {
      title: 'Water bottle',
      canContainAssets: false
    }, 'Parent is archived.')).toEqual({
      title: 'Could not restore Water bottle',
      message: 'Parent is archived. Check that its parent is active, then try again.'
    });

    expect(assetLifecycleFailurePresentation('delete', {
      title: 'Tool box',
      canContainAssets: true
    }, 'Asset has active children.')).toEqual({
      title: 'Could not permanently delete Tool box',
      message: 'Asset has active children. Permanent delete will not continue while active things are inside it.'
    });

    expect(assetLifecycleFailurePresentation('archive', {
      title: 'Tool box',
      canContainAssets: true
    }, 'Network request failed.')).toEqual({
      title: 'Could not archive Tool box',
      message: 'Network request failed.'
    });

    expect(assetLifecycleFailurePresentation('restore', {
      title: 'Water bottle',
      canContainAssets: false
    }, 'Session expired.')).toEqual({
      title: 'Could not restore Water bottle',
      message: 'Session expired.'
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
