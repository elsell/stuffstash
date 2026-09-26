import { expect, it } from 'vitest';
import { assetOverflowMenuGroups } from './AssetOverflowMenu';

it('keeps secondary commands discoverable in More and respects unavailable actions', () => {
  const calls: string[] = [];
  const base = { asset: { title: 'Tent', canArchive: false, canRestore: false, canDeletePermanently: false },
    onHistory() {}, onCheckoutHistory() {}, onLifecycleAction() {},
    onMove: () => calls.push('move'), onAddPhotos: () => calls.push('photos'), onCheckout: () => calls.push('checkout'), photosDisabled: true };
  const items = assetOverflowMenuGroups(base).flatMap(g => g.items);
  expect(items.map(i => i.label)).toEqual(['Add photos', 'Move', 'Check out', 'Checkout history', 'History']);
  expect(items.find(i => i.label === 'Add photos')?.disabled).toBe(true);
  items.find(i => i.label === 'Move')!.onPress();
  expect(calls).toEqual(['move']);
  expect(assetOverflowMenuGroups({ ...base, onMove: undefined }).flatMap(g => g.items).some(i => i.label === 'Move')).toBe(false);
});
