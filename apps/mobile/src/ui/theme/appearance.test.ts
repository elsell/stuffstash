import { describe, expect, it, vi } from 'vitest';
import {
  darkHighContrastPalette,
  darkPalette,
  lightHighContrastPalette,
  lightPalette
} from './tokens';

vi.mock('react-native', () => ({
  DynamicColorIOS: (variants: Record<string, string>) => `dynamic:${variants.light}:${variants.dark}`,
  Platform: { OS: 'ios' }
}));

import {
  appearanceAwarePalette,
  nativeAppearanceColorScheme,
  resolveAppearanceColorScheme
} from './appearance';

describe('appearance-aware mobile palette', () => {
  it('uses the resolved concrete palette on the first iOS frame', () => {
    expect(appearanceAwarePalette('light')).toBe(lightPalette);
    expect(appearanceAwarePalette('dark')).toBe(darkPalette);
    expect(appearanceAwarePalette('light', true)).toBe(lightHighContrastPalette);
    expect(appearanceAwarePalette('dark', true)).toBe(darkHighContrastPalette);
  });

  it('uses the resolved palette independently of native dynamic-color availability', () => {
    expect(appearanceAwarePalette('dark')).toBe(darkPalette);
    expect(appearanceAwarePalette('light')).toBe(lightPalette);
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
