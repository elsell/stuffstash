import { Alert } from 'react-native';
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
  model,
  onClose,
  onRemove,
  onSelectPhoto,
  photos
}: {
  readonly canRemove: boolean;
  readonly model: AssetPhotoViewerModel | undefined;
  readonly onClose: () => void;
  readonly onRemove: (photoId: string) => void;
  readonly onSelectPhoto: (photoId: string) => void;
  readonly photos: readonly AssetPhotoViewModel[];
}) {
  const selectedIndex = selectedAssetPhotoViewerIndex(photos, model);

  if (selectedIndex === undefined) {
    return null;
  }

  function removePhoto(photo: FullScreenPhotoViewerPhoto): void {
    if (!photo.id) {
      return;
    }

    Alert.alert('Remove photo?', 'This removes the photo from this asset.', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Remove',
        style: 'destructive',
        onPress: () => onRemove(photo.id as string)
      }
    ]);
  }

  return (
    <FullScreenPhotoViewer
      canRemove={canRemove}
      currentIndex={selectedIndex}
      onClose={onClose}
      onSelectIndex={(index) => {
        const nextModel = assetPhotoViewerModelAtIndex(photos, index);
        if (nextModel?.photo.id) {
          onSelectPhoto(nextModel.photo.id);
        }
      }}
      photos={photos.map(assetPhotoToFullScreenPhoto)}
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
