import { useMemo } from 'react';
import { Alert } from 'react-native';
import { useTaskPresentation } from '../navigation/useTaskPresentation';
import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';
import {
  assetPhotoMetadataLabel,
  assetPhotoViewerModelAtIndex,
  selectedAssetPhotoViewerIndex,
  type AssetPhotoViewerModel
} from '../components/AssetPhotoWorkspacePresentation';
import {
  FullScreenPhotoViewer,
  type FullScreenPhotoViewerPhoto
} from '../components/FullScreenPhotoViewer';

export function AssetPhotoViewerSheet({
  canRemove,
  isRemoving = false,
  model,
  onClose,
  onRemove,
  onSelectPhoto,
  photos
}: {
  readonly canRemove: boolean;
  readonly isRemoving?: boolean;
  readonly model: AssetPhotoViewerModel | undefined;
  readonly onClose: () => void;
  readonly onRemove: (photoId: string) => void;
  readonly onSelectPhoto: (photoId: string) => void;
  readonly photos: readonly AssetPhotoViewModel[];
}) {
  const viewerPhotos = useMemo(() => photos.map(assetPhotoToFullScreenPhoto), [photos]);
  const selectedIndex = selectedAssetPhotoViewerIndex(photos, model);

  const capturePresentation = useTaskPresentation(undefined, JSON.stringify([
    photos.map(photo => photo.id), selectedIndex, canRemove, isRemoving
  ]));

  if (selectedIndex === undefined) {
    return null;
  }

  function removePhoto(photo: FullScreenPhotoViewerPhoto): void {
    if (!photo.id || !canRemove || isRemoving) {
      return;
    }

    const isCurrent = capturePresentation();
    let accepted = false;
    Alert.alert('Remove photo?', 'This removes the photo from this asset.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Remove',
        style: 'destructive',
        onPress: () => {
          if (!isCurrent() || accepted) return;
          accepted = true;
          onRemove(photo.id as string);
        }
      }
    ]);
  }

  return (
    <FullScreenPhotoViewer
      canRemove={canRemove}
      isRemoving={isRemoving}
      currentIndex={selectedIndex}
      onClose={onClose}
      onSelectIndex={(index) => {
        const nextModel = assetPhotoViewerModelAtIndex(photos, index);
        if (nextModel?.photo.id) {
          onSelectPhoto(nextModel.photo.id);
        }
      }}
      photos={viewerPhotos}
      {...(canRemove ? { onRemove: removePhoto } : {})}
    />
  );
}

function assetPhotoToFullScreenPhoto(photo: AssetPhotoViewModel): FullScreenPhotoViewerPhoto {
  return {
    id: photo.id,
    label: photo.fileName ?? photo.label,
    metadataLabel: assetPhotoMetadataLabel(photo),
    uri: photo.viewerUri ?? photo.heroUri ?? photo.uri,
    headers: photo.viewerHeaders ?? photo.heroHeaders ?? photo.headers
  };
}
