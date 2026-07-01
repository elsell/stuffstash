import type { AddAssetPhotoProgressEvent } from '../../application/assets/AddAssetPhotosCommand';
import type { SelectedAssetPhoto } from '../../application/add/PhotoSelectionQuery';
import type { AssetPhotoUploadProgressViewModel } from '../components/AssetDetailView';

export type PhotoUploadRow = AssetPhotoUploadProgressViewModel;

export function photoUploadRows(photos: readonly SelectedAssetPhoto[]): readonly PhotoUploadRow[] {
  return photos.map((photo, index) => ({
    index,
    fileName: photo.fileName,
    status: 'pending'
  }));
}

export function applyPhotoUploadProgress(
  rows: readonly PhotoUploadRow[],
  event: AddAssetPhotoProgressEvent
): readonly PhotoUploadRow[] {
  let didUpdate = false;
  const nextRows = rows.map((row) => {
    if (row.index !== event.index || row.fileName !== event.fileName) {
      return row;
    }
    didUpdate = true;
    return { ...row, status: event.status };
  });

  return didUpdate ? nextRows : rows;
}
