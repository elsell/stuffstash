import type { AppearancePreference } from '../../application/settings/AppearancePreference';
import {
  mobileColorPalette,
  type MobileColorPalette
} from './tokens';

export type ResolvedColorScheme = 'light' | 'dark';
export type NativeAppearanceColorScheme = ResolvedColorScheme | 'unspecified';
type SupportedColorScheme = ResolvedColorScheme | 'unspecified' | null | undefined;

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
  increasedContrast = false
): MobileColorPalette {
  return mobileColorPalette(colorScheme, increasedContrast);
}
