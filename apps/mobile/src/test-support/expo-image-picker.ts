import type { ImagePickerOptions, ImagePickerResult } from 'expo-image-picker';

export const photoPickerFake = {
  cameraGranted: false,
  libraryPermissionRequests: 0,
  cameraLaunches: 0,
  libraryLaunches: 0,
  libraryOptions: undefined as ImagePickerOptions | undefined,
  result: { canceled: true, assets: null } as ImagePickerResult,
  reset() {
    this.cameraGranted = false;
    this.libraryPermissionRequests = 0;
    this.cameraLaunches = 0;
    this.libraryLaunches = 0;
    this.libraryOptions = undefined;
    this.result = { canceled: true, assets: null };
  }
};

export async function requestMediaLibraryPermissionsAsync() {
  photoPickerFake.libraryPermissionRequests += 1;
  return { granted: false };
}

export async function requestCameraPermissionsAsync() {
  return { granted: photoPickerFake.cameraGranted };
}

export async function launchImageLibraryAsync(options: ImagePickerOptions) {
  photoPickerFake.libraryOptions = options;
  photoPickerFake.libraryLaunches += 1;
  return photoPickerFake.result;
}

export async function launchCameraAsync() {
  if (!photoPickerFake.cameraGranted) throw new Error('Camera permission denied');
  photoPickerFake.cameraLaunches += 1;
  return photoPickerFake.result;
}
