import { describe, expect, it } from 'vitest';
import { tagColorModalLayout } from './TagColorPickerPresentation';

describe('tag color modal layout', () => {
  it('compacts the fixed picker for keyboard-shortened and large-text viewports', () => {
    expect(tagColorModalLayout({ availableHeight: 520, fontScale: 1 }).compactSpectrum).toBe(true);
    expect(tagColorModalLayout({ availableHeight: 800, fontScale: 1.5 }).compactSpectrum).toBe(true);
    expect(tagColorModalLayout({ availableHeight: 800, fontScale: 1 }).compactSpectrum).toBe(false);
  });
});
