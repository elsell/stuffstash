import type { Ref } from 'react';
import { requireNativeView } from 'expo';
import type { ViewProps } from 'react-native';
import type { SheetBoundaryPort } from './SheetBoundaryPort';

// Only the iOS sheet adapter mounts this native boundary.
export const NativeSheetBoundary = requireNativeView<ViewProps & { ref?: Ref<SheetBoundaryPort> }>('StuffStashSheetBoundary');
