import { describe, expect, it, vi } from 'vitest';
import {
  darkHighContrastPalette,
  darkPalette,
  lightHighContrastPalette,
  lightPalette
} from './tokens';

const dynamicColor = vi.fn((variants: Record<string, string>) => `dynamic:${variants.light}:${variants.dark}`);

vi.mock('react-native', () => ({
  DynamicColorIOS: (variants: Record<string, string>) => dynamicColor(variants),
  Platform: { OS: 'ios' }
}));

import {
  appearanceAwarePalette,
  nativeAppearanceColorScheme,
  resolveAppearanceColorScheme
} from './appearance';

describe('appearance-aware mobile palette', () => {
  it('creates iOS semantic colors with both appearances and contrast variants', () => {
    const palette = appearanceAwarePalette('dark', 'ios');

    expect(palette.background).toBe(`dynamic:${lightPalette.background}:${darkPalette.background}`);
    expect(dynamicColor).toHaveBeenCalledWith({
      light: lightPalette.background,
      dark: darkPalette.background,
      highContrastLight: lightHighContrastPalette.background,
      highContrastDark: darkHighContrastPalette.background
    });
  });

  it('uses the resolved palette on platforms without DynamicColorIOS', () => {
    expect(appearanceAwarePalette('dark', 'android')).toBe(darkPalette);
    expect(appearanceAwarePalette('light', 'android')).toBe(lightPalette);
  });

  it('resolves explicit preferences independently of the system scheme', () => {
    expect(resolveAppearanceColorScheme('system', 'dark')).toBe('dark');
    expect(resolveAppearanceColorScheme('system', null)).toBe('light');
    expect(resolveAppearanceColorScheme('light', 'dark')).toBe('light');
    expect(resolveAppearanceColorScheme('dark', 'light')).toBe('dark');
  });

  it('removes the React Native override for the System preference', () => {
    expect(nativeAppearanceColorScheme('system')).toBe('unspecified');
    expect(nativeAppearanceColorScheme('light')).toBe('light');
    expect(nativeAppearanceColorScheme('dark')).toBe('dark');
  });
});
