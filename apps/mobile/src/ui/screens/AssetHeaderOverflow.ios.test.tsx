import { describe, expect, it } from 'vitest';
import { assetHeaderOverflowScreenOptions } from './AssetHeaderOverflow.ios';

describe('AssetHeaderOverflow iOS native header contract', () => {
  it('installs a visible native navigation menu with grouped actions', () => {
    const calls: string[] = [];
    const props = {
      asset: { title: 'Drill', canArchive: true, canRestore: false, canDeletePermanently: true },
      disabled: true,
      onCheckoutHistory: () => calls.push('checkout-history'),
      onHistory: () => calls.push('history'),
      onLifecycleAction: (action: 'archive' | 'restore' | 'delete') => calls.push(action)
    };

    const options = assetHeaderOverflowScreenOptions(props);
    expect(options.headerShown).toBe(true);
    expect('headerRight' in options).toBe(false);

    const headerItems = options.unstable_headerRightItems;
    expect(headerItems).toBeTypeOf('function');
    const items = headerItems?.({ canGoBack: true }) ?? [];
    expect(items).toHaveLength(1);
    expect(items[0]).toMatchObject({
      accessibilityLabel: 'More actions for Drill',
      disabled: true,
      icon: { type: 'sfSymbol', name: 'ellipsis' },
      label: '',
      type: 'menu'
    });

    const menu = items[0];
    if (menu?.type !== 'menu') {
      throw new Error('Expected the iOS overflow header item to be a native menu.');
    }
    const groups = menu.menu.items.filter((entry) => entry.type === 'submenu');
    expect(groups).toHaveLength(3);
    expect(groups.every((group) => group.type === 'submenu' && group.inline)).toBe(true);
    const actions = groups
      .flatMap((group) => group.items)
      .filter((entry) => entry.type === 'action');
    expect(actions.map((action) => action.label)).toEqual([
      'Checkout history', 'History', 'Archive', 'Delete permanently'
    ]);
    expect(actions.at(-1)?.destructive).toBe(true);
    actions[1]?.onPress();
    actions[2]?.onPress();
    expect(calls).toEqual(['history', 'archive']);
  });
});


it('places native Edit beside More and disables it during pending work', () => {
  const calls: string[] = [];
  const props = { asset: { title: 'Drill', canArchive: false, canRestore: false, canDeletePermanently: false },
    onEdit: () => calls.push('edit'), onHistory: () => {}, onCheckoutHistory: () => {}, onLifecycleAction: () => {} };
  const items = assetHeaderOverflowScreenOptions(props).unstable_headerRightItems?.({ canGoBack: true }) ?? [];
  expect(items.map(item => item.type)).toEqual(['button', 'menu']);
  const edit = items[0];
  if (edit.type !== 'button') throw new Error('Expected native Edit');
  expect(edit.accessibilityLabel).toBe('Edit');
  edit.onPress();
  const disabled = assetHeaderOverflowScreenOptions({ ...props, disabled: true }).unstable_headerRightItems?.({ canGoBack: true })?.[0];
  if (disabled?.type !== 'button') throw new Error('Expected disabled native Edit');
  expect(disabled.disabled).toBe(true);
  disabled.onPress();
  expect(calls).toEqual(['edit']);
});
