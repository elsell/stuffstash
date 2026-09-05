import {
  CreateInventoryAssetPhotoInput,
  InventoryAssetPhotoDirectUpload
} from '../../application/home/InventorySummaryRepository';
import {
  directUploadMethod,
  isDirectUploadHTTPTransportAllowed,
  isLocalDirectUploadURL,
  type DirectUploadTargetPolicy
} from '../uploads/DirectUploadPolicy';

export type DirectUploadTransport = {
  upload(input: DirectUploadTransportInput): Promise<boolean>;
};

export type DirectUploadTransportInput = {
  readonly upload: InventoryAssetPhotoDirectUpload;
  readonly fileUri: string;
  readonly fileName: string;
  readonly contentType: CreateInventoryAssetPhotoInput['contentType'];
};

export class ExpoDirectUploadTransport implements DirectUploadTransport {
  constructor(private readonly directUploadPolicy: DirectUploadTargetPolicy = {}) {}

  async upload(input: DirectUploadTransportInput): Promise<boolean> {
    if (this.directUploadPolicy.allowLocalDevelopmentTargets === true && isLocalDirectUploadURL(input.upload.url)) {
      return false;
    }
    if (!isDirectUploadHTTPTransportAllowed(input.upload.url, this.directUploadPolicy)) {
      throw new Error('Direct attachment upload target must use HTTPS or a private local development host.');
    }
    const FileSystem = await import('expo-file-system/legacy');
    const uploadMethod = directUploadMethod(input.upload.method);
    const result = await FileSystem.uploadAsync(input.upload.url, input.fileUri, {
      httpMethod: uploadMethod,
      headers: input.upload.headers,
      ...(Object.keys(input.upload.formFields).length > 0
        ? {
            uploadType: FileSystem.FileSystemUploadType.MULTIPART,
            fieldName: 'file',
            mimeType: input.contentType,
            parameters: input.upload.formFields
          }
        : {
            uploadType: FileSystem.FileSystemUploadType.BINARY_CONTENT
          })
    });
    if (result.status < 200 || result.status >= 300) {
      throw new Error('Direct attachment upload failed.');
    }
    return true;
  }
}

export async function attachmentContentBase64(input: CreateInventoryAssetPhotoInput): Promise<string> {
  if (input.contentBase64) {
    return input.contentBase64;
  }
  if (!input.uri) {
    throw new Error('Attachment content is not available for JSON upload fallback.');
  }
  const FileSystem = await import('expo-file-system/legacy');
  return FileSystem.readAsStringAsync(input.uri, { encoding: FileSystem.EncodingType.Base64 });
}

