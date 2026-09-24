import type { ComponentType } from 'react';
import { requireNativeView } from 'expo';
import { Platform, type ViewProps } from 'react-native';

let AndroidPhotoSystemBars: ComponentType<ViewProps> | undefined;

/** Mount inside the photo dialog; Activity-level StatusBar cannot style it. */
export function PhotoViewerSystemBars() {
  if (Platform.OS !== 'android') return null;
  AndroidPhotoSystemBars ??= requireNativeView<ViewProps>('StuffStashPhotoSystemBars');
  return <AndroidPhotoSystemBars accessible={false} pointerEvents="none" style={{ width: 0, height: 0 }} />;
}
