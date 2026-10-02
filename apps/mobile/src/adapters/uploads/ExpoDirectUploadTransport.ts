import { t } from '../../presentation/localization';
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

type NativeFileUploader = (input: DirectUploadTransportInput, method: 'POST' | 'PUT' | 'PATCH') => Promise<{ readonly status: number }>;

export class ExpoDirectUploadTransport implements DirectUploadTransport {
  constructor(
    private readonly directUploadPolicy: DirectUploadTargetPolicy = {},
    private readonly uploadFile: NativeFileUploader = uploadNativeFile
  ) {}

  async upload(input: DirectUploadTransportInput): Promise<boolean> {
    if (this.directUploadPolicy.allowLocalDevelopmentTargets === true && isLocalDirectUploadURL(input.upload.url)) {
      return false;
    }
    if (!isDirectUploadHTTPTransportAllowed(input.upload.url, this.directUploadPolicy)) {
      throw new Error(t('recovery.uploadConfiguration'));
    }
    const uploadMethod = directUploadMethod(input.upload.method);
    const result = await this.uploadFile(input, uploadMethod);
    if (result.status < 200 || result.status >= 300) {
      throw new Error(t('recovery.uploadFailed'));
    }
    return true;
  }
}

export async function attachmentContentBase64(input: CreateInventoryAssetPhotoInput): Promise<string> {
  if (input.contentBase64) {
    return input.contentBase64;
  }
  if (!input.uri) {
    throw new Error(t('recovery.uploadContentUnavailable'));
  }
  const FileSystem = await import('expo-file-system/legacy');
  return FileSystem.readAsStringAsync(input.uri, { encoding: FileSystem.EncodingType.Base64 });
}

async function uploadNativeFile(input: DirectUploadTransportInput, uploadMethod: 'POST' | 'PUT' | 'PATCH'): Promise<{ readonly status: number }> {
  const FileSystem = await import('expo-file-system/legacy');
  return FileSystem.uploadAsync(input.upload.url, input.fileUri, {
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
}
