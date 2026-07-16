import { DynamicColorIOS, Platform } from 'react-native';
import type { AppearancePreference } from '../../application/settings/AppearancePreference';
import {
  darkHighContrastPalette,
  darkPalette,
  lightHighContrastPalette,
  lightPalette,
  mobileColorPalette,
  type MobileColorPalette
} from './tokens';

export type ResolvedColorScheme = 'light' | 'dark';
export type NativeAppearanceColorScheme = ResolvedColorScheme | 'unspecified';
type SupportedColorScheme = ResolvedColorScheme | 'unspecified' | null | undefined;

let cachedIOSPalette: MobileColorPalette | undefined;

export function resolveAppearanceColorScheme(
  preference: AppearancePreference,
  systemColorScheme: SupportedColorScheme
): ResolvedColorScheme {
  if (preference !== 'system') {
    return preference;
  }
  return systemColorScheme === 'dark' ? 'dark' : 'light';
}

export function nativeAppearanceColorScheme(
  preference: AppearancePreference
): NativeAppearanceColorScheme {
  return preference === 'system' ? 'unspecified' : preference;
}

export function appearanceAwarePalette(
  colorScheme: SupportedColorScheme,
  platform = Platform.OS,
  increasedContrast = false
): MobileColorPalette {
  if (platform !== 'ios' || typeof DynamicColorIOS !== 'function') {
    return mobileColorPalette(colorScheme, increasedContrast);
  }

  cachedIOSPalette ??= Object.fromEntries(Object.keys(lightPalette).map((key) => {
    const paletteKey = key as keyof MobileColorPalette;
    return [paletteKey, DynamicColorIOS({
      light: lightPalette[paletteKey],
      dark: darkPalette[paletteKey],
      highContrastLight: lightHighContrastPalette[paletteKey],
      highContrastDark: darkHighContrastPalette[paletteKey]
    })];
  })) as unknown as MobileColorPalette;
  return cachedIOSPalette;
}
