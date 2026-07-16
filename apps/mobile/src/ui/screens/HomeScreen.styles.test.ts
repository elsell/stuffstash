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
    expect('backgroundColor' in darkStyles.contextControl).toBe(false);
    expect(darkStyles.contextInventory.color).toBe(darkPalette.text);
    expect(darkStyles.returnSheetSaveText.color).toBe(darkPalette.onAction);
  });

  it('uses compact borderless navigation controls with Apple-sized touch targets', () => {
    const styles = createHomeScreenStyles(lightPalette);

    expect(styles.contextControl.minHeight).toBeGreaterThanOrEqual(44);
    expect(styles.contextControl.minHeight).toBeLessThanOrEqual(52);
    expect('borderWidth' in styles.contextControl).toBe(false);
    expect(styles.settingsButton.minHeight).toBeGreaterThanOrEqual(44);
    expect(styles.settingsButton.minWidth).toBeGreaterThanOrEqual(44);
    expect('borderWidth' in styles.settingsButton).toBe(false);
  });

  it('uses only modest fallback padding because Native Tabs owns the dynamic bottom inset', () => {
    const styles = createHomeScreenStyles(lightPalette);

    expect(styles.content.paddingBottom).toBeLessThanOrEqual(32);
  });

  it('uses a calmer type hierarchy instead of extra-bold supporting chrome', () => {
    const styles = createHomeScreenStyles(lightPalette);

    expect(Number(styles.contextInventory.fontWeight)).toBeLessThanOrEqual(700);
    expect(Number(styles.contextTenantPrefix.fontWeight)).toBeLessThanOrEqual(600);
    expect(Number(styles.sectionTitle.fontWeight)).toBeLessThanOrEqual(700);
    expect(Number(styles.sectionAction.fontWeight)).toBeLessThanOrEqual(600);
  });

  it('lets accessibility-sized context labels grow without colliding with toolbar actions', () => {
    const styles = createHomeScreenStyles(lightPalette);

    expect(styles.homeTopBar.alignItems).toBe('flex-start');
    expect(styles.contextControl.flex).toBe(1);
    expect(styles.contextText.minWidth).toBe(0);
    expect(styles.topBarActions.alignSelf).toBe('flex-start');
  });
});
