import React from 'react';
import { afterEach, describe, expect, it } from 'vitest';
import type { TestInstance } from 'test-renderer';
import { MobileRenderHarness } from '../../test-support/render';
import { darkPalette } from '../theme/tokens';
import { FullSpectrumTagColorPicker } from './FullSpectrumTagColorPicker';
import { swatchForeground, TagColorPicker, tagColorName } from './TagColorPicker';

let harness: MobileRenderHarness | undefined;
afterEach(async () => { await harness?.unmount(); harness = undefined; });

async function renderPicker(value: string, changes: string[] = [], disabled = false) {
  harness = new MobileRenderHarness();
  await harness.render(<TagColorPicker disabled={disabled} onChange={(next) => changes.push(next)} palette={darkPalette} value={value} />);
  return { changes, screen: harness };
}

describe('TagColorPicker', () => {
  it('chooses checkmark contrast from the actual swatch color', () => {
    expect(swatchForeground('#2F80ED')).toBe('#000000');
    expect(swatchForeground('#2E7D32')).toBe('#FFFFFF');
    expect(swatchForeground('#D97706')).toBe('#000000');
  });

  it('offers swatches and clears the optional tag color', async () => {
    const { changes, screen } = await renderPicker('#2f80ed');
    expect(screen.byLabel('Choose Blue tag color')?.props.accessibilityState).toMatchObject({ selected: true });
    await screen.press(screen.byLabel('Choose Green tag color'));
    await screen.press(screen.byLabel('No tag color'));
    expect(changes).toEqual(['#2E7D32', '']);
  });

  it('distinguishes invalid typed colors from no color', async () => {
    const { screen } = await renderPicker('#oops');
    expect(screen.allText()).toContain('Choose Custom… to correct this color.');
    expect(screen.byLabel('No tag color')?.props.accessibilityState).toMatchObject({ selected: false });
    expect(tagColorName('#oops')).toBe('Invalid color');
  });

  it('marks disabled color choices unavailable', async () => {
    const { changes, screen } = await renderPicker('#2F80ED', [], true);
    for (const label of ['No tag color', 'Choose Green tag color', 'Choose a custom tag color']) {
      expect(screen.byLabel(label)?.props.accessibilityState).toMatchObject({ disabled: true });
      expect(screen.byLabel(label)?.props.disabled).toBe(true);
    }
    expect(changes).toEqual([]);
  });

  it('provides human color semantics for rows and assistive technology', () => {
    expect(tagColorName(undefined)).toBe('No color');
    expect(tagColorName('#2f80ed')).toBe('Blue');
    expect(tagColorName('#123456')).toBe('Custom color #123456');
  });

  it('uses the supplied appearance palette for neutral controls and labels', async () => {
    const { screen } = await renderPicker('');
    expect(screen.byLabel('No tag color')?.props.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ backgroundColor: darkPalette.surface, borderColor: darkPalette.border }),
      expect.objectContaining({ borderColor: darkPalette.action })
    ]));
    expect(screen.byLabel('Choose a custom tag color')?.props.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ backgroundColor: darkPalette.surface, borderColor: darkPalette.border })
    ]));
  });

  it('keeps the full-spectrum picker collapsed behind a clear Custom option', async () => {
    const { screen } = await renderPicker('');
    expect(screen.byLabel('Choose a custom tag color')).toBeDefined();
    expect(screen.byLabel('Saturation and brightness')).toBeUndefined();
    expect(screen.allText()).not.toContain('Custom color');
  });

  it('opens a dedicated custom-color modal with spectrum, hex, clear, cancel, and done', async () => {
    const { changes, screen } = await renderPicker('#123456');
    await screen.press(screen.byLabel('Choose a custom tag color'));
    expect(screen.allText()).toContain('Custom color');
    expect(screen.byLabel('Saturation and brightness')).toBeDefined();
    expect(screen.byLabel('Custom tag color hex value')).toBeDefined();
    expect(screen.allText()).toEqual(expect.arrayContaining(['Clear color', 'Cancel', 'Done']));
    expect(hasAncestorType(screen.byLabel('Saturation and brightness'), 'ScrollView')).toBe(false);
    expect(hasAncestorType(screen.byLabel('Custom tag color hex value'), 'ScrollView')).toBe(true);
    expect(hasAncestorType(screen.byText('Clear color'), 'ScrollView')).toBe(true);
    expect(hasAncestorType(screen.byText('Cancel'), 'ScrollView')).toBe(false);
    expect(hasAncestorType(screen.byText('Done'), 'ScrollView')).toBe(false);

    await screen.changeText(screen.byLabel('Custom tag color hex value'), '#A1B2C3');
    await screen.press(screen.byText('Done')?.parent ?? undefined);
    expect(changes).toEqual(['#A1B2C3']);
  });

  it('preserves an arbitrary existing custom color as the selected indicator', async () => {
    const { screen } = await renderPicker('#123456');
    expect(screen.byLabel('Choose a custom tag color')?.props.accessibilityState).toMatchObject({ selected: true });
    expect(screen.byTestId('custom-tag-color-indicator')?.props.style).toEqual(expect.arrayContaining([
      expect.objectContaining({ backgroundColor: '#123456' })
    ]));
  });

  it('cancels the custom modal without mutating the original color', async () => {
    const { changes, screen } = await renderPicker('#123456');
    await screen.press(screen.byLabel('Choose a custom tag color'));
    await screen.accessibilityAction(screen.byLabel('Saturation and brightness'), 'increment');
    await screen.press(screen.byText('Cancel')?.parent ?? undefined);
    expect(changes).toEqual([]);
    expect(screen.allText()).not.toContain('Custom color');
  });

  it('renders the project spectrum fallback and applies accessible HSV adjustments', async () => {
    const changes: string[] = [];
    harness = new MobileRenderHarness();
    await harness.render(<FullSpectrumTagColorPicker onChange={(value) => changes.push(value)} value="#2F80ED" />);
    const spectrum = harness.byLabel('Saturation and brightness');
    expect(spectrum?.props.accessibilityRole).toBe('adjustable');
    expect(spectrum?.props.accessibilityValue.text).toMatch(/saturation/);
    const before = spectrum?.props.accessibilityValue.text;
    await harness.accessibilityAction(spectrum, 'increment');
    expect(changes).toHaveLength(1);
    expect(changes[0]).toMatch(/^#[0-9A-F]{6}$/);
    expect(changes[0]).not.toBe('#2F80ED');
    expect(harness.byLabel('Saturation and brightness')?.props.accessibilityValue.text).not.toBe(before);
  });

  it('compacts the fixed fallback while keeping keyboard actions available in a short viewport', async () => {
    const { changes, screen } = await renderPicker('#123456');
    await screen.press(screen.byLabel('Choose a custom tag color'));
    const modalShell = screen.all().find((node) => node.props.accessibilityViewIsModal);
    await screen.run(() => modalShell?.props.onLayout?.({ nativeEvent: { layout: { height: 520 } } }));
    const spectrum = screen.byLabel('Saturation and brightness');
    expect(spectrum?.props.style).toEqual(expect.arrayContaining([expect.objectContaining({ height: 112 })]));
    expect(screen.allText()).toEqual(expect.arrayContaining(['Hex color', 'Clear color', 'Cancel', 'Done']));
    expect(screen.allText()).not.toContain('Brightness');
    await screen.run(() => spectrum?.props.onLayout?.({ nativeEvent: { layout: { width: 320, height: 112 } } }));
    await screen.run(() => screen.byLabel('Saturation and brightness')?.props.onPanResponderGrant?.({ nativeEvent: { locationX: 320, locationY: 0 } }));
    expect(screen.byLabel('Custom tag color hex value')?.props.value).not.toBe('#000000');
    await screen.run(() => screen.byLabel('Saturation and brightness')?.props.onPanResponderGrant?.({ nativeEvent: { locationX: 320, locationY: 112 } }));
    expect(screen.byLabel('Custom tag color hex value')?.props.value).toBe('#000000');
    await screen.press(screen.byText('Done')?.parent ?? undefined);
    expect(changes).toEqual(['#000000']);
  });
});

function hasAncestorType(node: TestInstance | undefined, type: string): boolean {
  let parent = node?.parent;
  while (parent) {
    if (parent.type === type) return true;
    parent = parent.parent;
  }
  return false;
}
