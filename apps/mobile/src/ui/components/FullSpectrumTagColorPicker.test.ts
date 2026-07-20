import { describe, expect, it } from 'vitest';
import { adjustSpectrumValue, androidSpectrumAccessibility, expoUIColorPickerAvailable, fullSpectrumPickerKind, spectrumGestureOwnership } from './FullSpectrumTagColorPickerPresentation';

describe('full spectrum tag color picker', () => {
  it('uses the SwiftUI native picker only on iOS', () => {
    expect(fullSpectrumPickerKind('ios', true)).toBe('native-ios');
    expect(fullSpectrumPickerKind('ios', false)).toBe('project-spectrum');
    expect(fullSpectrumPickerKind('android', true)).toBe('project-spectrum');
  });

  it('requires both ExpoUI native views before selecting SwiftUI', () => {
    const supported = (moduleName: string, viewName: string) => moduleName === 'ExpoUI' && ['HostView', 'ColorPickerView'].includes(viewName) ? {} : undefined;
    const staleHost = (_moduleName: string, viewName: string) => viewName === 'ColorPickerView' ? {} : undefined;
    const stalePicker = (_moduleName: string, viewName: string) => viewName === 'HostView' ? {} : undefined;

    expect(expoUIColorPickerAvailable('ios', supported)).toBe(true);
    expect(expoUIColorPickerAvailable('ios', staleHost)).toBe(false);
    expect(expoUIColorPickerAvailable('ios', stalePicker)).toBe(false);
    expect(expoUIColorPickerAvailable('android', supported)).toBe(false);
  });

  it('fails closed when a stale binary throws while resolving native view config', () => {
    expect(expoUIColorPickerAvailable('ios', () => { throw new Error('Unimplemented component'); })).toBe(false);
  });

  it('keeps spectrum gestures in the fixed modal interaction surface', () => {
    expect(spectrumGestureOwnership.onPanResponderTerminationRequest()).toBe(false);
    expect(spectrumGestureOwnership.onShouldBlockNativeResponder()).toBe(true);
  });

  it('exposes both Android gesture surfaces as direct adjustable controls', () => {
    const semantics = androidSpectrumAccessibility({ hue: 210, saturation: 0.6, brightness: 0.8 }, false);

    expect(semantics.spectrum).toMatchObject({
      accessibilityRole: 'adjustable',
      accessibilityState: { disabled: false },
      accessibilityValue: { text: '60 percent saturation, 80 percent brightness' }
    });
    expect(semantics.hue).toMatchObject({
      accessibilityRole: 'adjustable',
      accessibilityValue: { min: 0, max: 360, now: 210, text: '210 degrees' }
    });
    expect(semantics.hue.accessibilityActions.map((action) => action.name)).toEqual(['increment', 'decrement']);
    expect(semantics.spectrum.accessibilityActions.map((action) => action.name)).toEqual(['increment', 'decrement', 'increaseBrightness', 'decreaseBrightness']);
  });

  it('makes Android accessibility actions change and clamp the same color value', () => {
    const value = { hue: 358, saturation: 0.98, brightness: 0.8 };
    expect(adjustSpectrumValue(value, 'hue', 'increment').hue).toBe(3);
    expect(adjustSpectrumValue(value, 'spectrum', 'increment').saturation).toBe(1);
    expect(adjustSpectrumValue(value, 'spectrum', 'decrement').saturation).toBeCloseTo(0.93);
    expect(adjustSpectrumValue(value, 'spectrum', 'increaseBrightness').brightness).toBeCloseTo(0.85);
    expect(adjustSpectrumValue(value, 'hue', 'unknown')).toBe(value);
  });
});
