import type { SheetBottomBoundary } from './keyboardBoundaryInset';

/** Coordinates share the React Native keyboard event window coordinate space. */
export interface SheetBoundaryPort {
  measureInKeyboardWindow(): Promise<SheetBottomBoundary | null>;
}
