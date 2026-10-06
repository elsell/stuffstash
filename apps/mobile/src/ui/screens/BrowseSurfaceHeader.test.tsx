import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';
import { resetNavigation, navigationOptions } from '../../test-support/navigation';
import { BrowseSurfaceHeader } from './BrowseSurfaceHeader';
import type { InventoryMapSurface } from './InventoryMapPresentation';

it('settles header feedback, keeps current view handlers, and retires them on teardown', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  const calls: string[] = [];
  const render = (label: string, surface: InventoryMapSurface) => <NavigationOptionFeedback render={() =>
    <BrowseSurfaceHeader surface={surface} onChange={next => calls.push(`${label}:${next}`)} />} />;
  let change!: () => void;
  try {
    await h.render(render('old', 'list'));
    await h.press(h.byLabel('Browse view: List'));
    const choice = h.all().find(n => n.props.accessibilityRole === 'menuitem' && n.queryAll(c => c.type === 'Text' && c.children.includes('List')).length > 0)!;
    change = choice.props.onPress;
    expect(choice.props.accessibilityState.selected).toBe(true);
    expect(h.byType('NativeSegmentedControl')).toBeUndefined();
    expect((navigationOptions().at(-1) as { headerTitle: string }).headerTitle).toBe('');
    await h.render(render('current', 'map'));
    expect(h.byLabel('Browse view: Map')).toBeDefined();
    await h.run(change);
    expect(calls).toEqual(['current:list']);
  } finally { await h.unmount(); resetNavigation(); }
  await h.run(change);
  expect(calls).toEqual(['current:list']);
});
