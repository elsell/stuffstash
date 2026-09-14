import { expect, it } from 'vitest';
import { nativeHeaderActionOptions } from './NativeHeaderActions.ios';

it('installs real native bar items with system symbols, badges and actions', () => {
  const pressed: string[] = [];
  const options = nativeHeaderActionOptions([
    { kind: 'notifications', label: 'Notifications, 3 unread', badgeCount: 3, onPress: () => pressed.push('notifications') },
    { kind: 'add', label: 'Add an asset', onPress: () => pressed.push('add') },
    { kind: 'account', label: 'Open account and settings', onPress: () => pressed.push('account') }
  ]);
  const items = options.unstable_headerRightItems?.({ canGoBack: false });
  expect(items?.map(item => item.type)).toEqual(['button', 'button', 'button']);
  expect(items?.map(item => item.type === 'button' ? item.width : undefined)).toEqual([44, 44, 44]);
  expect(items?.[0]).toMatchObject({ accessibilityLabel: 'Notifications, 3 unread', badge: { value: 3 }, icon: { type: 'sfSymbol', name: 'bell' } });
  expect(items?.[1]).toMatchObject({ icon: { type: 'sfSymbol', name: 'plus' } });
  items?.forEach(item => { if (item.type === 'button') item.onPress?.(); });
  expect(pressed).toEqual(['notifications', 'add', 'account']);
});
it('removes stale native actions when permission or scope changes', () => {
  expect(nativeHeaderActionOptions([]).unstable_headerRightItems?.({ canGoBack: false })).toEqual([]);
  const item = nativeHeaderActionOptions([{ kind: 'notifications', label: 'Notifications, loading unread count', onPress: () => {} }]).unstable_headerRightItems?.({ canGoBack: false })[0];
  expect(item).not.toHaveProperty('badge');
});
