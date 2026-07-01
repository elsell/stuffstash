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
      quality: 0.72
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
      quality: 0.72
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
    const normalized = await normalizeSelectedPhoto(asset);
    if (!normalized) {
      continue;
    }
    selectedPhotos.push({
      id: `${asset.assetId ?? normalized.uri}-${selectedAt}-${index.toString()}`,
      uri: normalized.uri,
      fileName: normalizedFileName(asset.fileName, existingCount + index + 1, normalized.contentType),
      contentType: normalized.contentType,
      contentBase64: normalized.base64
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
};

type NormalizedPhoto = {
  readonly uri: string;
  readonly base64: string;
  readonly contentType: SelectedAssetPhoto['contentType'];
};

const maxVoicePhotoLongEdge = 1600;
const voicePhotoCompression = 0.72;

async function normalizeSelectedPhoto(asset: ImagePickerAssetLike): Promise<NormalizedPhoto | undefined> {
  if (!isSupportedImageContentType(asset.mimeType)) {
    return undefined;
  }
  try {
    const ImageManipulator = await import('expo-image-manipulator');
    const actions = resizeActionsForAsset(asset);
    const normalized = await ImageManipulator.manipulateAsync(asset.uri, actions, {
      base64: true,
      compress: voicePhotoCompression,
      format: ImageManipulator.SaveFormat.JPEG
    });
    if (normalized.base64) {
      return {
        uri: normalized.uri,
        base64: normalized.base64,
        contentType: 'image/jpeg'
      };
    }
  } catch {
    // Existing dev-client builds may not include newly added native modules yet.
  }
  const fallbackContentType = normalizeImageContentType(asset.mimeType);
  if (!asset.base64 || !fallbackContentType) {
    return undefined;
  }
  return { uri: asset.uri, base64: asset.base64, contentType: fallbackContentType };
}

function resizeActionsForAsset(asset: ImagePickerAssetLike): { readonly resize: { readonly width?: number; readonly height?: number } }[] {
  if (!asset.width || !asset.height) {
    return [{ resize: { width: maxVoicePhotoLongEdge } }];
  }
  if (asset.width <= maxVoicePhotoLongEdge && asset.height <= maxVoicePhotoLongEdge) {
    return [];
  }
  if (asset.width >= asset.height) {
    return [{ resize: { width: maxVoicePhotoLongEdge } }];
  }
  return [{ resize: { height: maxVoicePhotoLongEdge } }];
}

function normalizedFileName(fileName: string | null | undefined, fallbackIndex: number, contentType: SelectedAssetPhoto['contentType']): string {
  const safeName = fileName?.trim();
  const extension = extensionForContentType(contentType);
  if (!safeName) {
    return `asset-photo-${fallbackIndex}.${extension}`;
  }
  return safeName.replace(/\.[^.]+$/, '') + `.${extension}`;
}

function isSupportedImageContentType(value: string | undefined): boolean {
  return normalizeImageContentType(value) !== undefined;
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
