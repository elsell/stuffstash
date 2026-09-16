import type { HeaderOptions } from '../components/NativeHeaderActions.types';
import type { MobileColorPalette } from '../theme/tokens';

export function voiceNativeSheetOptions(palette: MobileColorPalette): HeaderOptions {
  return {
    contentStyle: { backgroundColor: palette.surface },
    headerShown: true,
    title: 'Conversation',
    presentation: 'formSheet',
    sheetAllowedDetents: [0.42, 0.88],
    sheetCornerRadius: 24,
    sheetExpandsWhenScrolledToEdge: true,
    sheetGrabberVisible: true,
    sheetInitialDetentIndex: 1,
    sheetLargestUndimmedDetentIndex: 'none'
  };
}
