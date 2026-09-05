import { describe, expect, it } from 'vitest';
import type { AssetDetailViewModel } from '../../application/assets/AssetViewModels';
import { mergeProgressiveAssetDetail } from './AssetDetailProgressivePresentation';

describe('mergeProgressiveAssetDetail', () => {
  it('keeps authoritative core fields while retaining independent secondary regions', () => {
    const core = detail({ title: 'Returned item', checkoutLabel: 'Available' });
    const contents = detail({
      title: 'Stale checked-out item',
      checkoutLabel: 'Checked out',
      parentLocationTrailLabel: 'Garage',
      parentLocationTrail: [{ id: 'garage', title: 'Garage', isImmediateParent: true }]
    });
    const photos = [{ id: 'photo-a', label: 'Front', uri: 'https://example.test/front' }];

    expect(mergeProgressiveAssetDetail(core, contents, photos)).toMatchObject({
      title: 'Returned item',
      checkoutLabel: 'Available',
      parentLocationTrailLabel: 'Garage',
      photos,
      photoLabel: 'Photo ready'
    });
  });
});

function detail(overrides: Partial<AssetDetailViewModel>): AssetDetailViewModel {
  return {
    id: 'asset-a',
    title: 'Item',
    kind: 'item',
    kindLabel: 'Item',
    description: '',
    locationTrailLabel: '',
    parentLocationTrailLabel: '',
    parentLocationTrail: [],
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
    containedSpaces: [],
    containedSpacesLabel: '0 spaces',
    containedItems: [],
    containedItemsLabel: '0 items',
    canContainAssets: false,
    canAddContainedAssets: false,
    updatedAtLabel: 'Updated now',
    photoLabel: 'Needs photo',
    imagePlaceholderLabel: 'Item',
    photos: [],
    ...overrides
  };
}
