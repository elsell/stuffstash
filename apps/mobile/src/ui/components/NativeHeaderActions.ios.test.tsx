import { expect, it } from 'vitest';
import { nativeHeaderActionOptions } from './NativeHeaderActions.ios';

it('dispatches a native leading Back command with the backward system symbol', () => {
  let returned = false;
  const item = nativeHeaderActionOptions([{ kind: 'back', label: 'Back to settings collection', onPress: () => { returned = true; } }], 'left').unstable_headerLeftItems?.({ canGoBack: false })[0];
  expect(item).toMatchObject({ icon: { type: 'sfSymbol', name: 'chevron.backward' }, accessibilityLabel: 'Back to settings collection' });
  if (item?.type === 'button') item.onPress?.();
  expect(returned).toBe(true);
});

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
it('provides a system close action for native sheets', () => {
  let closed = false;
  const item = nativeHeaderActionOptions([{ kind: 'close', label: 'Close inventory switcher', onPress: () => { closed = true; } }]).unstable_headerRightItems?.({ canGoBack: false })[0];
  expect(item).toMatchObject({ accessibilityLabel: 'Close inventory switcher', icon: { type: 'sfSymbol', name: 'xmark' } });
  if (item?.type === 'button') item.onPress?.();
  expect(closed).toBe(true);
});

it('uses the system compose action for a new conversation', () => {
  let started = false;
  const item = nativeHeaderActionOptions([{ kind: 'compose', label: 'New conversation', onPress: () => { started = true; } }]).unstable_headerRightItems?.({ canGoBack: false })[0];
  expect(item).toMatchObject({ accessibilityLabel: 'New conversation', icon: { type: 'sfSymbol', name: 'square.and.pencil' } });
  if (item?.type === 'button') item.onPress?.();
  expect(started).toBe(true);
});


it('keeps invalid save commands disabled and exposes a leading close command', () => {
  const calls: string[] = [];
  const save = nativeHeaderActionOptions([{ kind: 'save', label: 'Save item', disabled: true, onPress: () => calls.push('save') }]).unstable_headerRightItems?.({ canGoBack: false })[0];
  expect(save).toMatchObject({ disabled: true, icon: { type: 'sfSymbol', name: 'checkmark' } });
  if (save?.type === 'button') save.onPress?.();
  expect(calls).toEqual([]);
  const close = nativeHeaderActionOptions([{ kind: 'close', label: 'Cancel Add', onPress: () => calls.push('close') }], 'left').unstable_headerLeftItems?.({ canGoBack: false })[0];
  if (close?.type === 'button') close.onPress?.();
  expect(calls).toEqual(['close']);
});
