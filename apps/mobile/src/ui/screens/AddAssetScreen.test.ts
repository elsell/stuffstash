import { describe, expect, it } from 'vitest';
import {
  resolveParentAssetId,
  resolveSelectedParent
} from './AddAssetResolution';
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import {
  addHereParams,
  applyInitialParentToDraft,
  initialParentFromParams
} from './AddAssetInitialParent';

const garage: ParentLookupResult = {
  id: 'asset-garage',
  title: 'Garage',
  kind: 'location',
  subtitle: 'No parent',
  pathLabel: 'Garage',
  selectionHint: 'Location',
  willPromoteToContainer: false
};

describe('AddAssetScreen parent resolution', () => {
  it('resolves an exact typed parent without requiring the row to be tapped', () => {
    expect(resolveParentAssetId([garage], ' garage ', undefined)).toBe('asset-garage');
  });

  it('carries an exact typed parent into the next add after save', () => {
    const resolvedParentAssetId = resolveParentAssetId([garage], 'Garage', undefined);

    expect(resolveSelectedParent([garage], resolvedParentAssetId, 'Garage', undefined)).toEqual(
      garage
    );
  });

  it('keeps a newly created place selected even before search returns it again', () => {
    const createdPlace: ParentLookupResult = {
      id: 'asset-basement-shelf',
      title: 'Basement shelf',
      kind: 'location',
      subtitle: 'New location',
      pathLabel: 'Basement shelf',
      selectionHint: 'Location',
      willPromoteToContainer: false
    };

    expect(
      resolveSelectedParent([], 'asset-basement-shelf', 'Basement shelf', createdPlace)
    ).toEqual(createdPlace);
  });

  it('rejects unknown typed parents so users create or clear the parent first', () => {
    expect(() => resolveParentAssetId([garage], 'Basement shelf', undefined)).toThrow(
      'Create this parent or clear the Put in field.'
    );
  });

});

describe('AddAssetScreen initial parent routing', () => {
  it('parses a route-scoped parent prefill from safe params', () => {
    expect(initialParentFromParams({
      parentAssetId: 'asset-office-bin',
      parentTitle: 'Office bin',
      parentKind: 'container',
      parentSubtitle: 'Office',
      parentPathLabel: 'Office / Office bin',
      parentSelectionHint: 'Container',
      parentWillPromoteToContainer: 'false'
    })).toEqual({
      id: 'asset-office-bin',
      title: 'Office bin',
      kind: 'container',
      subtitle: 'Office',
      pathLabel: 'Office / Office bin',
      selectionHint: 'Container',
      willPromoteToContainer: false
    });
  });

  it('ignores incomplete or unsupported route-scoped parent prefill params', () => {
    expect(initialParentFromParams({
      parentAssetId: 'asset-office-bin',
      parentTitle: 'Office bin',
      parentKind: 'unsupported'
    })).toBeUndefined();
    expect(initialParentFromParams({
      parentAssetId: 'asset-office-bin',
      parentKind: 'container'
    })).toBeUndefined();
    expect(initialParentFromParams({
      parentAssetId: 'asset-office-bin',
      parentTitle: 'Office bin',
      parentKind: 'container',
      parentSubtitle: 'Office',
      parentSelectionHint: 'Container',
      parentWillPromoteToContainer: 'false'
    })).toBeUndefined();
    expect(initialParentFromParams({
      parentAssetId: 'asset-water-bottle',
      parentTitle: 'Water bottle',
      parentKind: 'item',
      parentSubtitle: 'Office',
      parentPathLabel: 'Office / Water bottle',
      parentSelectionHint: 'Item',
      parentWillPromoteToContainer: 'false'
    })).toBeUndefined();
    expect(initialParentFromParams({
      parentAssetId: 'asset-office-bin',
      parentTitle: 'Office bin',
      parentKind: 'container',
      parentSubtitle: 'Office',
      parentPathLabel: 'Office / Office bin',
      parentSelectionHint: 'Container',
      parentWillPromoteToContainer: 'true'
    })).toBeUndefined();
  });

  it('round-trips detail add-here params through the typed route contract', () => {
    const params = addHereParams({
      id: 'asset-office-bin',
      title: 'Office bin',
      kind: 'container',
      kindLabel: 'Container',
      description: '',
      parentAssetId: 'asset-office',
      locationTrailLabel: 'Office / Office bin',
      parentLocationTrailLabel: 'Office',
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
      canContainAssets: true,
      canAddContainedAssets: true,
      updatedAtLabel: 'Updated today',
      photoLabel: 'Needs photo',
      photos: [],
      imagePlaceholderLabel: 'Box'
    });

    expect(initialParentFromParams(params)).toEqual({
      id: 'asset-office-bin',
      title: 'Office bin',
      kind: 'container',
      subtitle: 'Office',
      pathLabel: 'Office / Office bin',
      selectionHint: 'Container',
      willPromoteToContainer: false
    });
  });

  it('applies a route-scoped parent without clearing the current add draft', () => {
    const draft = {
      title: 'HDMI cable',
      description: 'For the Apple TV.',
      parentAssetId: undefined,
      parentQuery: '',
      selectedPhotos: [{ id: 'photo-one', fileName: 'one.jpg', uri: 'file:///one.jpg', contentType: 'image/jpeg' as const, sizeBytes: 1024 }],
      showDetails: true,
      lastParent: undefined
    };
    const parent = initialParentFromParams({
      parentAssetId: 'asset-tv-box',
      parentTitle: 'TV box',
      parentKind: 'container',
      parentSubtitle: 'Living room',
      parentPathLabel: 'Living room / TV box',
      parentSelectionHint: 'Container',
      parentWillPromoteToContainer: 'false'
    });

    expect(applyInitialParentToDraft(draft, parent)).toEqual({
      ...draft,
      parentAssetId: 'asset-tv-box',
      parentQuery: 'TV box',
      lastParent: {
        id: 'asset-tv-box',
        title: 'TV box',
        kind: 'container',
        subtitle: 'Living room',
        pathLabel: 'Living room / TV box',
        selectionHint: 'Container',
        willPromoteToContainer: false
      }
    });
  });
});
