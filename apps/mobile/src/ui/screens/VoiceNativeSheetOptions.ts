import { Platform } from 'react-native';
import type { HeaderOptions } from '../components/NativeHeaderActions.types';
import type { MobileColorPalette } from '../theme/tokens';

export function voiceNativeSheetOptions(palette: MobileColorPalette, platform: string = Platform.OS): HeaderOptions {
  if (platform === 'android') return {
    contentStyle: { backgroundColor: palette.surface },
    headerShown: true, title: 'Conversation', presentation: 'card'
  };
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
