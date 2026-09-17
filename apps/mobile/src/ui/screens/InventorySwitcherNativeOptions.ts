import { Platform } from 'react-native';
import type { MobileColorPalette } from '../theme/tokens';

export function inventorySwitcherNativeOptions(palette: MobileColorPalette, platform: string = Platform.OS) {
  const base = { contentStyle: { backgroundColor: palette.surface }, headerShown: true };
  return platform === 'android' ? { ...base, presentation: 'card' as const } : {
    ...base, presentation: 'formSheet' as const, sheetAllowedDetents: [0.5, 1],
    sheetCornerRadius: 24, sheetGrabberVisible: true
  };
}
