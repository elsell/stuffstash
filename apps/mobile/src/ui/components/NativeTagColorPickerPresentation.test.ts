import { describe, expect, it } from 'vitest';
import { nativeTagColorInteraction, nativeTagColorSelection } from './NativeTagColorPickerPresentation';

describe('native tag color picker presentation', () => {
  it('represents empty and invalid optional colors as no native selection', () => {
    expect(nativeTagColorSelection('')).toBeNull();
    expect(nativeTagColorSelection('not-a-color')).toBeNull();
    expect(nativeTagColorSelection('#a1b2c3')).toBe('#A1B2C3');
  });

  it('removes native interaction when the form is read-only', () => {
    const onChange = (_value: string) => undefined;
    expect(nativeTagColorInteraction(true, onChange)).toEqual({ pointerEvents: 'none', onSelectionChange: undefined });
    expect(nativeTagColorInteraction(false, onChange)).toEqual({ pointerEvents: 'auto', onSelectionChange: onChange });
  });
});
