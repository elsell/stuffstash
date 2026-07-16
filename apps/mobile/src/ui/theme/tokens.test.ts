import { describe, expect, it } from 'vitest';
import {
  darkHighContrastPalette,
  darkPalette,
  lightHighContrastPalette,
  lightPalette,
  mobileColorPalette
} from './tokens';

describe('mobile theme color palettes', () => {
  it.each([
    ['light text', lightPalette.text, lightPalette.background],
    ['light muted text', lightPalette.textMuted, lightPalette.background],
    ['light action text', lightPalette.action, lightPalette.background],
    ['light filled action', lightPalette.onAction, lightPalette.action],
    ['light photo badge', lightPalette.onScrim, compositeOverWhite(lightPalette.scrim)],
    ['dark text', darkPalette.text, darkPalette.background],
    ['dark muted text', darkPalette.textMuted, darkPalette.background],
    ['dark action text', darkPalette.action, darkPalette.background],
    ['dark filled action', darkPalette.onAction, darkPalette.action],
    ['dark photo badge', darkPalette.onScrim, compositeOverWhite(darkPalette.scrim)]
  ])('%s meets WCAG AA text contrast', (_label, foreground, background) => {
    expect(contrastRatio(foreground, background)).toBeGreaterThanOrEqual(4.5);
  });

  it.each([
    ['light selected text', lightPalette.text, lightPalette.selected],
    ['light selected muted text', lightPalette.textMuted, lightPalette.selected],
    ['light warning badge', lightPalette.warning, lightPalette.warningSurface],
    ['light pressed action', lightPalette.onAction, lightPalette.actionPressed],
    ['dark selected text', darkPalette.text, darkPalette.selected],
    ['dark selected muted text', darkPalette.textMuted, darkPalette.selected],
    ['dark warning badge', darkPalette.warning, darkPalette.warningSurface],
    ['dark pressed action', darkPalette.onAction, darkPalette.actionPressed]
  ])('%s keeps text readable in interactive card states', (_label, foreground, background) => {
    expect(contrastRatio(foreground, background)).toBeGreaterThanOrEqual(4.5);
  });

  it.each([
    ['light control border', lightPalette.controlBorder, lightPalette.surface],
    ['light focus ring', lightPalette.focusRing, lightPalette.surface],
    ['dark control border', darkPalette.controlBorder, darkPalette.surface],
    ['dark focus ring', darkPalette.focusRing, darkPalette.surface],
    ['high-contrast light control border', lightHighContrastPalette.controlBorder, lightHighContrastPalette.surface],
    ['high-contrast light focus ring', lightHighContrastPalette.focusRing, lightHighContrastPalette.surface],
    ['high-contrast dark control border', darkHighContrastPalette.controlBorder, darkHighContrastPalette.surface],
    ['high-contrast dark focus ring', darkHighContrastPalette.focusRing, darkHighContrastPalette.surface]
  ])('%s has at least 3:1 non-text contrast', (_label, foreground, background) => {
    expect(contrastRatio(foreground, background)).toBeGreaterThanOrEqual(3);
  });

  it.each([
    ['light', lightPalette],
    ['dark', darkPalette]
  ])('%s keeps structural borders quieter than interactive control borders', (_label, palette) => {
    expect(contrastRatio(palette.border, palette.surface))
      .toBeLessThan(contrastRatio(palette.controlBorder, palette.surface));
  });

  it('resolves the requested appearance and increased-contrast variant', () => {
    expect(mobileColorPalette('light')).toBe(lightPalette);
    expect(mobileColorPalette('dark')).toBe(darkPalette);
    expect(mobileColorPalette('unspecified')).toBe(lightPalette);
    expect(mobileColorPalette('light', true)).toBe(lightHighContrastPalette);
    expect(mobileColorPalette('dark', true)).toBe(darkHighContrastPalette);
  });
});

function contrastRatio(foreground: string, background: string): number {
  const foregroundLuminance = relativeLuminance(foreground);
  const backgroundLuminance = relativeLuminance(background);
  const lighter = Math.max(foregroundLuminance, backgroundLuminance);
  const darker = Math.min(foregroundLuminance, backgroundLuminance);
  return (lighter + 0.05) / (darker + 0.05);
}

function compositeOverWhite(color: string): string {
  const match = color.match(/^rgba\((\d+),\s*(\d+),\s*(\d+),\s*([\d.]+)\)$/);
  if (!match) {
    return color;
  }
  const alpha = Number(match[4]);
  const channels = match.slice(1, 4).map((channel) => Math.round(
    (Number(channel) * alpha) + (255 * (1 - alpha))
  ));
  return `#${channels.map((channel) => channel.toString(16).padStart(2, '0')).join('')}`;
}

function relativeLuminance(color: string): number {
  const channels = [1, 3, 5].map((index) => Number.parseInt(color.slice(index, index + 2), 16) / 255);
  const [red, green, blue] = channels.map((channel) => channel <= 0.04045
    ? channel / 12.92
    : ((channel + 0.055) / 1.055) ** 2.4);
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}
