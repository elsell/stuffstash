import type {
  PhotoSelectionProvider,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';

export class ExpoPhotoSelectionProvider implements PhotoSelectionProvider {
  async selectPhotos(existingCount: number): Promise<readonly SelectedAssetPhoto[]> {
    const ImagePicker = await import('expo-image-picker');
    const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
    if (!permission.granted) {
      throw new Error('Photo library access is required to add photos.');
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      allowsMultipleSelection: true,
      base64: true,
      mediaTypes: ['images'],
      quality: 0.82
    });

    if (result.canceled) {
      return [];
    }

    const selectedAt = Date.now().toString();

    return result.assets.flatMap((asset, index): SelectedAssetPhoto[] => {
      const contentType = normalizeImageContentType(asset.mimeType);
      if (!asset.base64 || !contentType) {
        return [];
      }

      return [
        {
          id: `${asset.assetId ?? asset.uri}-${selectedAt}-${index.toString()}`,
          uri: asset.uri,
          fileName: asset.fileName ?? `asset-photo-${existingCount + index + 1}.jpg`,
          contentType,
          contentBase64: asset.base64
        }
      ];
    });
  }
}

function normalizeImageContentType(value: string | undefined): SelectedAssetPhoto['contentType'] | undefined {
  switch (value) {
    case 'image/png':
    case 'image/webp':
      return value;
    case 'image/jpg':
    case 'image/jpeg':
    case undefined:
      return 'image/jpeg';
    default:
      return undefined;
  }
}
