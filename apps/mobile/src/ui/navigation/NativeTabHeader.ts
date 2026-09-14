import type { HeaderOptions } from '../components/NativeHeaderActions.types';
import type { MobileColorPalette } from '../theme/tokens';

export function nativeTabHeaderOptions(palette: MobileColorPalette, platform: string, version: string | number, opaqueBackground = palette.surface): HeaderOptions {
  const appearance: HeaderOptions = platform === 'ios'
    ? {
        headerTransparent: true,
        headerShadowVisible: false,
        ...(Number.parseInt(String(version), 10) >= 26
          ? { scrollEdgeEffects: { top: 'soft' as const } }
          : { headerBlurEffect: 'systemMaterial' as const })
      }
    : { headerTransparent: false, headerStyle: { backgroundColor: opaqueBackground } };
  return {
    ...appearance,
    headerTintColor: palette.action,
    headerTitleStyle: { color: palette.text },
    contentStyle: { backgroundColor: palette.background }
  };
}
