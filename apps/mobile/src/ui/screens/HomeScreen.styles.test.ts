import { describe, expect, it, vi } from 'vitest';
import { darkPalette, lightPalette } from '../theme/tokens';
import { createHomeScreenStyles } from './HomeScreen.styles';

vi.mock('react-native', () => ({
  StyleSheet: { create: (styles: unknown) => styles }
}));

describe('createHomeScreenStyles', () => {
  it('resolves semantic surfaces and text for light and dark appearances', () => {
    const lightStyles = createHomeScreenStyles(lightPalette);
    const darkStyles = createHomeScreenStyles(darkPalette);

    expect(lightStyles.shell.backgroundColor).toBe(lightPalette.background);
    expect(darkStyles.shell.backgroundColor).toBe(darkPalette.background);
    expect(darkStyles.contextControl.backgroundColor).toBe(darkPalette.surface);
    expect(darkStyles.contextInventory.color).toBe(darkPalette.text);
    expect(darkStyles.returnSheetSaveText.color).toBe(darkPalette.onAction);
  });
});
