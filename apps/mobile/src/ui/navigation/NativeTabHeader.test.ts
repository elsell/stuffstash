import { expect, it } from 'vitest';
import { nativeTabHeaderOptions } from './NativeTabHeader';
import { lightPalette } from '../theme/tokens';
it('lets iOS 26 supply the soft scroll edge without an opaque bar or double blur', () => {
  const options = nativeTabHeaderOptions(lightPalette, 'ios', '26.0');
  expect(options.headerTransparent).toBe(true);
  expect(options.scrollEdgeEffects).toEqual({ top: 'soft' });
  expect(options.headerStyle).toBeUndefined();
  expect(options.headerBlurEffect).toBeUndefined();
  expect(options.headerShadowVisible).toBe(false);
});
it('uses native material on older iOS and preserves opaque Android layout', () => {
  expect(nativeTabHeaderOptions(lightPalette, 'ios', '18.6')).toMatchObject({ headerTransparent: true, headerBlurEffect: 'systemMaterial' });
  const android = nativeTabHeaderOptions(lightPalette, 'android', 36);
  expect(android.headerTransparent).toBe(false);
  expect(android.headerStyle).toEqual({ backgroundColor: lightPalette.surface });
  expect(android.scrollEdgeEffects).toBeUndefined();
  expect(android.headerBlurEffect).toBeUndefined();
});
