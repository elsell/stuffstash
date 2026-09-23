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

it('keeps Add header commands available on Android without changing the iOS sheet', () => {
  const android = createAssetNativeSheetOptions(colors, 'android').add;
  expect(android).toMatchObject({ presentation: 'card', headerShown: true, title: 'Add item' });
  expect(android).not.toHaveProperty('sheetAllowedDetents');
  expect(createAssetNativeSheetOptions(colors, 'ios').add).toMatchObject({
    presentation: 'formSheet', headerShown: true, title: 'Add item', sheetAllowedDetents: [1]
  });
});

it('keeps Android checkout history Close available while retaining the iOS detents', () => {
  expect(createAssetNativeSheetOptions(colors, 'android').checkoutHistory).toMatchObject({
    presentation: 'card', headerShown: true, title: 'Checkout history'
  });
  expect(createAssetNativeSheetOptions(colors, 'android').checkoutHistory).not.toHaveProperty('sheetAllowedDetents');
  expect(createAssetNativeSheetOptions(colors, 'ios').checkoutHistory).toMatchObject({
    presentation: 'formSheet', sheetAllowedDetents: [0.58, 0.92]
  });
});


it('opens Edit at full height with persistent native completion controls', () => {
  expect(createAssetNativeSheetOptions(colors, 'ios').edit).toMatchObject({
    presentation: 'formSheet', headerShown: true, title: 'Edit asset',
    sheetAllowedDetents: [1], gestureEnabled: false
  });
});


it('gives Move a full-height destination picker and persistent native commands', () => {
  expect(createAssetNativeSheetOptions(colors, 'ios').move).toMatchObject({
    presentation: 'formSheet', headerShown: true, title: 'Move asset', sheetAllowedDetents: [1]
  });
});
