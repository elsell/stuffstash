import { useTaskPresentation } from '../navigation/useTaskPresentation';
import { useMemo } from 'react';
import { Alert } from 'react-native';
import type { SelectedAssetPhoto } from '../../application/add/PhotoSelectionQuery';
import { FullScreenPhotoViewer, type FullScreenPhotoViewerPhoto } from '../components/FullScreenPhotoViewer';
import { photoMetadataLabel } from '../components/AssetPhotoWorkspacePresentation';

export function DraftPhotoPreviewModal({
  currentIndex,
  disabled,
  onClose,
  onRemovePhoto,
  onSetIndex,
  photos
}: {
  readonly disabled: boolean;
  readonly currentIndex: number | undefined;
  readonly onClose: () => void;
  readonly onRemovePhoto: (photoId: string) => void;
  readonly onSetIndex: (index: number | undefined) => void;
  readonly photos: readonly SelectedAssetPhoto[];
}) {
  const viewerPhotos = useMemo(() => photos.map(photo => ({
    id: photo.id,
    label: photo.fileName,
    metadataLabel: photoMetadataLabel(photo),
    uri: photo.uri
  })), [photos]);
  const capturePresentation = useTaskPresentation(undefined, JSON.stringify([
    photos.map(photo => photo.id), currentIndex, disabled
  ]));

  function removeCurrentPhoto(photo: FullScreenPhotoViewerPhoto, index: number): void {
    if (!photo.id || disabled || currentIndex === undefined) {
      return;
    }

    const isCurrent = capturePresentation();
    let accepted = false;
    Alert.alert('Remove photo?', 'This removes the photo from this new item draft.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Remove',
        style: 'destructive',
        onPress: () => {
          if (!isCurrent() || accepted) return;
          accepted = true;
          onRemovePhoto(photo.id as string);
          if (photos.length <= 1) {
            onClose();
            return;
          }

          onSetIndex(Math.min(index, photos.length - 2));
        }
      }
    ]);
  }

  return (
    <FullScreenPhotoViewer
      canRemove={!disabled}
      currentIndex={currentIndex}
      onClose={onClose}
      onRemove={removeCurrentPhoto}
      onSelectIndex={onSetIndex}
      photos={viewerPhotos}
    />
  );
}
