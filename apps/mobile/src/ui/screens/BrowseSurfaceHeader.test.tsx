import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';
import { resetNavigation } from '../../test-support/navigation';
import { BrowseSurfaceHeader } from './BrowseSurfaceHeader';
import type { InventoryMapSurface } from './InventoryMapPresentation';

it('settles header feedback, keeps current view handlers, and retires them on teardown', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  const calls: string[] = [];
  const render = (label: string, surface: InventoryMapSurface) => <NavigationOptionFeedback render={() =>
    <BrowseSurfaceHeader surface={surface} onChange={next => calls.push(`${label}:${next}`)} />} />;
  let change!: (label: string) => void;
  try {
    await h.render(render('old', 'list'));
    const control = h.byType('NativeSegmentedControl');
    change = control!.props.onValueChange;
    await h.render(render('current', 'map'));
    expect(h.byType('NativeSegmentedControl')).toBe(control);
    expect(control?.props.selectedIndex).toBe(1);
    await h.run(() => change('List'));
    expect(calls).toEqual(['current:list']);
  } finally { await h.unmount(); resetNavigation(); }
  await h.run(() => change('Map'));
  expect(calls).toEqual(['current:list']);
});
