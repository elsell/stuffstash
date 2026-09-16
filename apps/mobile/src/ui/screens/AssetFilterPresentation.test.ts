import { expect, it } from 'vitest';
import { createAssetNativeSheetOptions } from './AssetNativeSheetOptions';
import { colors } from '../theme/tokens';

it('uses an Android native stack filter route without sheet footer lifecycle', () => {
  const filters = createAssetNativeSheetOptions(colors, 'android').filters;
  expect(filters.presentation).toBe('card');
  expect(filters.headerShown).toBe(true);
  expect(filters).not.toHaveProperty('sheetAllowedDetents');
});

it('retains the iOS filter sheet', () => {
  expect(createAssetNativeSheetOptions(colors, 'ios').filters).toMatchObject({
    presentation: 'formSheet', headerShown: true, sheetAllowedDetents: [0.7, 1]
  });
});

it('keeps Android asset actions out of partial sheet geometry', () => {
  const options = createAssetNativeSheetOptions(colors, 'android');
  for (const action of [options.edit, options.move, options.moveHere]) {
    expect(action.presentation).toBe('card');
    expect(action.headerShown).toBe(true);
    expect(action).not.toHaveProperty('sheetAllowedDetents');
  }
  expect(options.edit.gestureEnabled).toBe(false);
});
