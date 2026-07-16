import { describe, expect, it } from 'vitest';
import { expoSegmentedControlAvailable } from './SettingsSegmentedControlPresentation';

describe('settings segmented control runtime selection', () => {
  it('requires every SwiftUI view used by the iOS control', () => {
    const available = new Set(['HostView', 'PickerView', 'TextView']);
    expect(expoSegmentedControlAvailable('ios', (_module, view) => available.has(view))).toBe(true);
    available.delete('PickerView');
    expect(expoSegmentedControlAvailable('ios', (_module, view) => available.has(view))).toBe(false);
  });

  it('requires every Compose view used by the Android control', () => {
    const available = new Set(['HostView', 'SingleChoiceSegmentedButtonRowView', 'SegmentedButtonView', 'TextView']);
    expect(expoSegmentedControlAvailable('android', (_module, view) => available.has(view))).toBe(true);
    available.delete('SegmentedButtonView');
    expect(expoSegmentedControlAvailable('android', (_module, view) => available.has(view))).toBe(false);
  });

  it('uses the accessible fallback when capability lookup is unavailable or throws', () => {
    expect(expoSegmentedControlAvailable('web', () => ({}))).toBe(false);
    expect(expoSegmentedControlAvailable('ios', () => { throw new Error('missing native view'); })).toBe(false);
  });
});
