import { describe, expect, it } from 'vitest';
import type { ParentTargetViewModel } from '$lib/domain/inventory';
import {
  addAssetKindCopy,
  addDestinationSummary,
  addFormPresentation,
  addPhotoAcceptTypes,
  addPhotoCountLabel,
  addPhotoHelpText,
  addPhotoPickerPresentation,
  addPhotoRemoveLabel,
  addPhotoSupportedTypeLabel,
  addSupportedImageTypes,
  assetKindControlOptions,
  quickParentContainerLabel,
  quickParentContainerSummary,
  quickParentContainerTrail,
  quickParentMissingNameMessage,
  quickParentKindOptions
} from './workspaceAddPresentation';

describe('workspace add presentation helpers', () => {
  it('builds stable asset-kind and quick-parent options', () => {
    expect(assetKindControlOptions()).toEqual([
      { value: 'item', label: 'Item' },
      { value: 'container', label: 'Container' },
      { value: 'location', label: 'Location' }
    ]);
    expect(quickParentKindOptions).toEqual([
      { value: 'location', label: 'Location' },
      { value: 'container', label: 'Container' }
    ]);
  });

  it('builds stable add form labels and placeholders', () => {
    expect(addFormPresentation).toEqual({
      summaryTypeLabel: 'Type',
      summaryParentLabel: 'Parent',
      summaryPhotosLabel: 'Photos',
      assetKindLegend: 'Asset kind',
      parentPickerLegend: 'Place in existing parent',
      parentPickerGroupLabel: 'Parent target',
      quickParentLegend: 'Create missing parent',
      quickParentToggleLabel: 'Create a parent first',
      quickParentToggleDescription: 'Use this when the shelf, box, or location does not exist yet.',
      quickParentContextLabel: 'Created under',
      quickParentNameLabel: 'Parent name',
      quickParentNamePlaceholder: 'Laundry shelf',
      quickParentKindLabel: 'New parent kind',
      descriptionLabel: 'Description',
      descriptionPlaceholder: 'Optional notes'
    });
  });

  it('builds kind-specific add labels and placeholders', () => {
    expect(addAssetKindCopy('item')).toEqual({
      heading: 'Add item',
      kindLabel: 'Item',
      nameLabel: 'Item name',
      namePlaceholder: 'Tomato fertilizer',
      saveLabel: 'Save item',
      selectedKindLabel: 'item'
    });
    expect(addAssetKindCopy('container').namePlaceholder).toBe('Clear storage bin');
    expect(addAssetKindCopy('location').namePlaceholder).toBe('Garage shelf');
  });

  it('summarizes existing and quick-created parent destinations', () => {
    const parent = parentTarget('garage-shelf', 'Garage shelf', 'Garage');

    expect(addDestinationSummary({ quickParentEnabled: false, quickParentKind: 'location', quickParentTitle: '', selectedParent: null })).toBe(
      'Inventory root'
    );
    expect(addDestinationSummary({ quickParentEnabled: false, quickParentKind: 'location', quickParentTitle: '', selectedParent: parent })).toBe(
      'Garage shelf'
    );
    expect(
      addDestinationSummary({
        quickParentEnabled: true,
        quickParentKind: 'container',
        quickParentTitle: '  Clear bin  ',
        selectedParent: parent
      })
    ).toBe('New Container: Clear bin in Garage shelf / Garage');
    expect(addDestinationSummary({ quickParentEnabled: true, quickParentKind: 'location', quickParentTitle: '', selectedParent: null })).toBe(
      'New Location in Inventory root'
    );
  });

  it('summarizes quick-parent container context and photo counts', () => {
    const parent = parentTarget('hall-closet', 'Hall closet', 'Hall');

    expect(quickParentContainerLabel(null)).toBe('Inventory root');
    expect(quickParentContainerLabel(parent)).toBe('Hall closet');
    expect(quickParentContainerTrail(null)).toBe('');
    expect(quickParentContainerTrail(parent)).toBe('Hall');
    expect(quickParentContainerSummary(parent)).toBe('Hall closet / Hall');
    expect(quickParentMissingNameMessage()).toBe('Enter a parent name or turn this option off.');
    expect(addPhotoCountLabel(0)).toBe('No photos');
    expect(addPhotoCountLabel(1)).toBe('1 photo');
    expect(addPhotoCountLabel(2)).toBe('2 photos');
  });

  it('builds photo picker labels and supported image type copy', () => {
    expect(addPhotoPickerPresentation).toEqual({
      actionGroupLabel: 'Photo actions',
      uploadLabel: 'Upload',
      cameraLabel: 'Camera',
      uploadInputLabel: 'Upload photos',
      cameraInputLabel: 'Take photo',
      selectedListLabel: 'Selected photos'
    });

    const supportedTypes = addSupportedImageTypes({
      supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp', 'application/pdf'],
      maxBytes: 1024
    });

    expect(supportedTypes).toEqual(['image/jpeg', 'image/png', 'image/webp']);
    expect(addPhotoAcceptTypes(supportedTypes)).toBe('image/jpeg,image/png,image/webp');
    expect(addPhotoSupportedTypeLabel([])).toBe('No image formats');
    expect(addPhotoSupportedTypeLabel(['image/png'])).toBe('PNG');
    expect(addPhotoSupportedTypeLabel(['image/jpeg', 'image/png'])).toBe('JPEG or PNG');
    expect(addPhotoSupportedTypeLabel(supportedTypes)).toBe('JPEG, PNG, or WebP');
    expect(addPhotoSupportedTypeLabel(['image/heic'])).toBe('HEIC');
    expect(addPhotoHelpText('JPEG, PNG, or WebP', '1 KB')).toBe('Optional JPEG, PNG, or WebP up to 1 KB.');
    expect(addPhotoRemoveLabel({ name: 'front.jpg' })).toBe('Remove front.jpg');
  });
});

function parentTarget(id: string, title: string, containmentTrail: string): ParentTargetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'container',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}
