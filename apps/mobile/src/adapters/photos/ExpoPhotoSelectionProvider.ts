import type {
  PhotoSelectionProvider,
  SelectedAssetPhoto
} from '../../application/add/PhotoSelectionQuery';

export class ExpoPhotoSelectionProvider implements PhotoSelectionProvider {
  async selectFromLibrary(existingCount: number): Promise<readonly SelectedAssetPhoto[]> {
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

    return mapImagePickerResult(result, existingCount);
  }

  async captureFromCamera(existingCount: number): Promise<readonly SelectedAssetPhoto[]> {
    const ImagePicker = await import('expo-image-picker');
    const permission = await ImagePicker.requestCameraPermissionsAsync();
    if (!permission.granted) {
      throw new Error('Camera access is required to take a photo.');
    }

    const result = await ImagePicker.launchCameraAsync({
      base64: true,
      mediaTypes: ['images'],
      quality: 0.82
    });

    return mapImagePickerResult(result, existingCount);
  }
}

function mapImagePickerResult(
  result: { readonly canceled: boolean; readonly assets?: readonly ImagePickerAssetLike[] | null },
  existingCount: number
): readonly SelectedAssetPhoto[] {
  if (result.canceled) {
    return [];
  }

  const selectedAt = Date.now().toString();

  return (result.assets ?? []).flatMap((asset, index): SelectedAssetPhoto[] => {
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

type ImagePickerAssetLike = {
  readonly assetId?: string | null;
  readonly uri: string;
  readonly fileName?: string | null;
  readonly mimeType?: string;
  readonly base64?: string | null;
};

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
