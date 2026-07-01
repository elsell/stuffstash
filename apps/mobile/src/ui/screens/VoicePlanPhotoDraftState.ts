import type { SelectedAssetPhoto } from '../../application/add/PhotoSelectionQuery';

export type VoicePlanPhotoDrafts = Record<string, readonly SelectedAssetPhoto[]>;

export function appendVoicePlanPhotoDrafts(
  current: VoicePlanPhotoDrafts,
  commandKey: string,
  photos: readonly SelectedAssetPhoto[]
): VoicePlanPhotoDrafts {
  if (photos.length === 0) {
    return current;
  }

  return {
    ...current,
    [commandKey]: [...(current[commandKey] ?? []), ...photos]
  };
}

export function removeVoicePlanPhotoDraft(
  current: VoicePlanPhotoDrafts,
  commandKey: string,
  photoId: string
): VoicePlanPhotoDrafts {
  const nextPhotos = (current[commandKey] ?? []).filter((photo) => photo.id !== photoId);
  if (nextPhotos.length === (current[commandKey] ?? []).length) {
    return current;
  }
  if (nextPhotos.length > 0) {
    return {
      ...current,
      [commandKey]: nextPhotos
    };
  }

  const { [commandKey]: _removed, ...next } = current;
  return next;
}
