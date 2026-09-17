import { expect, it } from 'vitest';
import { inventorySwitcherNativeOptions } from './InventorySwitcherNativeOptions';
import { colors } from '../theme/tokens';

it('exposes the inventory switcher header on Android and keeps the iOS sheet', () => {
  const android = inventorySwitcherNativeOptions(colors, 'android');
  expect(android).toMatchObject({ presentation: 'card', headerShown: true });
  expect(android).not.toHaveProperty('sheetAllowedDetents');
  expect(inventorySwitcherNativeOptions(colors, 'ios')).toMatchObject({
    presentation: 'formSheet', headerShown: true, sheetAllowedDetents: [0.5, 1]
  });
});
