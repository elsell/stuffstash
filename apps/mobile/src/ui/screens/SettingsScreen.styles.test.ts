import { describe, expect, it, vi } from 'vitest';
import { darkPalette, lightPalette } from '../theme/tokens';
import { createSettingsScreenStyles } from './SettingsScreen.styles';

vi.mock('react-native', () => ({
  StyleSheet: { create: (styles: unknown) => styles }
}));

describe('createSettingsScreenStyles', () => {
  it('keeps every navigable row and compact icon control at least 44 points tall', () => {
    const styles = createSettingsScreenStyles(lightPalette, { stacked: false });

    expect(styles.navigationRow.minHeight).toBeGreaterThanOrEqual(44);
    expect(styles.navigationRow.minWidth).toBeGreaterThanOrEqual(44);
    expect(styles.iconButton.minHeight).toBeGreaterThanOrEqual(44);
    expect(styles.iconButton.minWidth).toBeGreaterThanOrEqual(44);
    expect(styles.choiceRow.minHeight).toBeGreaterThanOrEqual(44);
    expect(styles.actionRow.minHeight).toBeGreaterThanOrEqual(44);
  });

  it('vertically centers navigation, choice, and action content in their row targets', () => {
    const styles = createSettingsScreenStyles(lightPalette, { stacked: false });

    expect(styles.navigationRow.justifyContent).toBe('center');
    expect(styles.choiceRow.justifyContent).toBe('center');
    expect(styles.actionRow.justifyContent).toBe('center');
  });

  it('uses grouped-list surfaces and separators rather than card borders', () => {
    const styles = createSettingsScreenStyles(lightPalette, { stacked: false });

    expect(styles.group.backgroundColor).toBe(lightPalette.surface);
    expect(styles.group).not.toHaveProperty('borderWidth');
    expect(styles.navigationRow).not.toHaveProperty('borderWidth');
    expect(styles.separator.backgroundColor).toBe(lightPalette.border);
  });

  it('uses the same semantic hierarchy in dark appearance', () => {
    const styles = createSettingsScreenStyles(darkPalette, { stacked: true });

    expect(styles.shell.backgroundColor).toBe(darkPalette.background);
    expect(styles.group.backgroundColor).toBe(darkPalette.surface);
    expect(styles.rowLabel.color).toBe(darkPalette.text);
    expect(styles.rowValue.color).toBe(darkPalette.textMuted);
  });

  it('reflows row content vertically at accessibility sizes', () => {
    const compact = createSettingsScreenStyles(lightPalette, { stacked: false });
    const accessible = createSettingsScreenStyles(lightPalette, { stacked: true });

    expect(compact.navigationRowContent.flexDirection).toBe('row');
    expect(accessible.navigationRowContent.flexDirection).toBe('column');
    expect(accessible.navigationRowContent.alignItems).toBe('flex-start');
    expect(accessible.actionGroup.flexDirection).toBe('column');
    expect(accessible.choiceGroup.flexDirection).toBe('column');
  });
});
