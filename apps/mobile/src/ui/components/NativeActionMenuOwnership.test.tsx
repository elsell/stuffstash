import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeActionMenu as FallbackMenu } from './NativeActionMenu';
import { NativeActionMenu as AndroidMenu } from './NativeActionMenu.android';
import { NativeActionMenu as IOSMenu } from './NativeActionMenu.ios';

for (const [platform, Menu] of [['fallback', FallbackMenu], ['ios', IOSMenu], ['android', AndroidMenu]] as const) {
  it(`${platform} menu rejects retained actions after lock, removal and teardown, then uses current handlers`, async () => {
    const h = new MobileRenderHarness();
    const calls: string[] = [];
    const render = (owner: string, disabled = false, removed = false, itemDisabled = false, groupId = 'asset') =>
      h.render(<Menu accessibilityLabel="Actions" disabled={disabled} groups={[{ id: groupId, items: removed ? [] : [
        { id: 'archive', label: 'Archive', disabled: itemDisabled, onPress: () => calls.push(owner) }
      ] }]} />);
    const open = async () => { if (platform === 'fallback') await h.press(h.byLabel('Actions')); else if (platform === 'android') await h.run(h.byType('ComposeTextButton')!.props.onClick); };
    try {
      await render('old'); await open();
      const node = platform === 'ios' ? h.byType('SwiftUIButton') : platform === 'android' ? h.byType('ComposeDropdownMenuItem') : h.all().find(n => n.props.accessibilityRole === 'menuitem');
      expect(node).toBeDefined();
      const retained = platform === 'android' ? node!.props.onClick : node!.props.onPress;
      await render('locked', true);
      await h.run(retained);
      expect(calls).toEqual([]);
      await render('removed', false, true);
      await h.run(retained);
      expect(calls).toEqual([]);
      await render('item locked', false, false, true);
      await h.run(retained);
      expect(calls).toEqual([]);
      await render('different group', false, false, false, 'inventory');
      await h.run(retained);
      expect(calls).toEqual([]);
      await render('current');
      await h.run(retained);
      expect(calls).toEqual(['current']);
      await h.unmount();
      await h.run(retained);
      expect(calls).toEqual(['current']);
    } finally { await h.unmount(); }
  });
}

for (const [platform, Menu] of [['fallback', FallbackMenu], ['android', AndroidMenu]] as const) {
  it(`${platform} closes an open menu on lock and stays closed after unlock`, async () => {
    const h = new MobileRenderHarness();
    const render = (disabled: boolean) => h.render(<Menu accessibilityLabel="Actions" disabled={disabled}
      groups={[{ id: 'asset', items: [{ id: 'archive', label: 'Archive', onPress: () => undefined }] }]} />);
    const hasItem = () => platform === 'android'
      ? Boolean(h.byType('ComposeDropdownMenuItem'))
      : h.all().some(node => node.props.accessibilityRole === 'menuitem');
    try {
      await render(false);
      if (platform === 'android') await h.run(h.byType('ComposeTextButton')!.props.onClick);
      else await h.press(h.byLabel('Actions'));
      expect(hasItem()).toBe(true);
      const retainedTrigger = platform === 'android'
        ? h.byType('ComposeTextButton')!.props.onClick
        : h.byLabel('Actions')!.props.onPress;
      await render(true);
      await h.run(retainedTrigger);
      expect(hasItem()).toBe(false);
      await render(false);
      expect(hasItem()).toBe(false);
    } finally { await h.unmount(); }
  });
}
