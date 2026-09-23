import { describe, expect, it } from 'vitest';
import { nativeTagColorSelection } from './NativeTagColorPickerPresentation';

describe('native tag color picker presentation', () => {
  it('represents empty and invalid optional colors as no native selection', () => {
    expect(nativeTagColorSelection('')).toBeNull();
    expect(nativeTagColorSelection('not-a-color')).toBeNull();
    expect(nativeTagColorSelection('#a1b2c3')).toBe('#A1B2C3');
  });

});
