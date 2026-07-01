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
      base64: false,
      mediaTypes: ['images'],
      quality: 1
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
      base64: false,
      mediaTypes: ['images'],
      quality: 1
    });

    return mapImagePickerResult(result, existingCount);
  }
}

async function mapImagePickerResult(
  result: { readonly canceled: boolean; readonly assets?: readonly ImagePickerAssetLike[] | null },
  existingCount: number
): Promise<readonly SelectedAssetPhoto[]> {
  if (result.canceled) {
    return [];
  }

  const selectedAt = Date.now().toString();
  const selectedPhotos: SelectedAssetPhoto[] = [];

  for (const [index, asset] of (result.assets ?? []).entries()) {
    const contentType = normalizeImageContentType(asset.mimeType);
    if (!contentType) {
      continue;
    }
    selectedPhotos.push({
      id: `${asset.assetId ?? asset.uri}-${selectedAt}-${index.toString()}`,
      uri: asset.uri,
      fileName: normalizedFileName(asset.fileName, existingCount + index + 1, contentType),
      contentType,
      contentBase64: asset.base64 ?? undefined,
      sizeBytes: asset.fileSize ?? decodedBase64ByteLength(asset.base64 ?? '')
    });
  }

  return selectedPhotos;
}

type ImagePickerAssetLike = {
  readonly assetId?: string | null;
  readonly uri: string;
  readonly fileName?: string | null;
  readonly mimeType?: string;
  readonly base64?: string | null;
  readonly width?: number;
  readonly height?: number;
  readonly fileSize?: number;
};

function normalizedFileName(fileName: string | null | undefined, fallbackIndex: number, contentType: SelectedAssetPhoto['contentType']): string {
  const safeName = fileName?.trim();
  const extension = extensionForContentType(contentType);
  if (!safeName) {
    return `asset-photo-${fallbackIndex}.${extension}`;
  }
  return safeName.replace(/\.[^.]+$/, '') + `.${extension}`;
}

function normalizeImageContentType(value: string | undefined): SelectedAssetPhoto['contentType'] | undefined {
  switch (value) {
    case 'image/png':
      return 'image/png';
    case 'image/webp':
      return 'image/webp';
    case 'image/jpg':
    case 'image/jpeg':
    case undefined:
      return 'image/jpeg';
    default:
      return undefined;
  }
}

function decodedBase64ByteLength(value: string): number {
  const padding = value.endsWith('==') ? 2 : value.endsWith('=') ? 1 : 0;
  return Math.max(0, Math.floor((value.length * 3) / 4) - padding);
}

function extensionForContentType(contentType: SelectedAssetPhoto['contentType']): string {
  switch (contentType) {
    case 'image/png':
      return 'png';
    case 'image/webp':
      return 'webp';
    case 'image/jpeg':
      return 'jpg';
  }
}

export const __expoPhotoSelectionProviderTestHooks = {
  mapImagePickerResult
};
