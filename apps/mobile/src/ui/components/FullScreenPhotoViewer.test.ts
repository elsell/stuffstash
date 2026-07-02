import { describe, expect, it } from 'vitest';
import { fullScreenPhotoViewerActionState } from './FullScreenPhotoViewerPresentation';

describe('fullScreenPhotoViewerActionState', () => {
  it('builds bottom-toolbar state for the current image', () => {
    expect(fullScreenPhotoViewerActionState([
      { id: 'photo-one', label: 'one.jpg', uri: 'file://one' },
      { id: 'photo-two', label: 'two.jpg', metadataLabel: 'JPEG image · 1.2 MB', uri: 'file://two' },
      { id: 'photo-three', label: 'three.jpg', uri: 'file://three' }
    ], 1, true)).toEqual({
      canGoPrevious: true,
      canGoNext: true,
      canRemove: true,
      fileLabel: 'two.jpg',
      metadataLabel: 'JPEG image · 1.2 MB',
      positionLabel: '2 of 3'
    });
  });

  it('disables unavailable actions at boundaries and without a removable id', () => {
    expect(fullScreenPhotoViewerActionState([
      { label: 'draft.jpg', uri: 'file://draft' }
    ], 0, true)).toEqual({
      canGoPrevious: false,
      canGoNext: false,
      canRemove: false,
      fileLabel: 'draft.jpg',
      metadataLabel: undefined,
      positionLabel: '1 of 1'
    });
  });

  it('returns a safe empty state for stale image indexes', () => {
    expect(fullScreenPhotoViewerActionState([
      { id: 'photo-one', label: 'one.jpg', uri: 'file://one' }
    ], -1, true)).toEqual({
      canGoPrevious: false,
      canGoNext: false,
      canRemove: false,
      fileLabel: 'Photo',
      metadataLabel: undefined,
      positionLabel: '0 of 0'
    });
    expect(fullScreenPhotoViewerActionState([
      { id: 'photo-one', label: 'one.jpg', uri: 'file://one' }
    ], 1, true)).toEqual({
      canGoPrevious: false,
      canGoNext: false,
      canRemove: false,
      fileLabel: 'Photo',
      metadataLabel: undefined,
      positionLabel: '0 of 0'
    });
  });
});
