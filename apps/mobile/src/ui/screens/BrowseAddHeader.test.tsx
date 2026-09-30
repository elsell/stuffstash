import React from 'react';
import { lightPalette } from '../theme/tokens';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation, navigationOptions } from '../../test-support/navigation';
import { BrowseAddHeader } from './BrowseAddHeader';
it('installs Add in the navigation slot and removes it when permission is lost', async () => {
  resetNavigation(); const h = new MobileRenderHarness(); let adds = 0;
  try {
    await h.render(<BrowseAddHeader canAdd onAdd={() => adds++} />);
    await h.press(h.byLabel('Add an asset'));
    expect(adds).toBe(1);
    await h.render(<BrowseAddHeader canAdd={false} onAdd={() => adds++} />);
    expect(h.byLabel('Add an asset')).toBeUndefined();
  } finally { await h.unmount(); }
});
it('keeps a Browse title and clears the old leading inventory control', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<BrowseAddHeader canAdd onAdd={() => {}} />);
    expect(navigationOptions().at(-1)).toMatchObject({ title: 'Browse', headerLeft: undefined });
  } finally { await h.unmount(); }
});

it('settles navigation feedback and keeps a changed Add command live', async () => {
  resetNavigation(); const h = new MobileRenderHarness(); const calls: string[] = [];
  const render = (label: string, canAdd = true) => <NavigationOptionFeedback render={() => <BrowseAddHeader canAdd={canAdd} onAdd={() => calls.push(label)} />} />;
  try {
    await h.render(render('old'));
    const oldAction = h.byLabel('Add an asset')!.props.onPress;
    await h.render(render('current'));
    await h.run(oldAction);
    expect(calls).toEqual(['current']);
    await h.render(render('denied', false));
    await h.run(oldAction);
    expect(calls).toEqual(['current']);
    expect(h.byLabel('Add an asset')).toBeUndefined();
  } finally { await h.unmount(); resetNavigation(); }
});

it('places Filters after Add, updates its count and keeps it available without creation permission', async () => {
  resetNavigation(); const h = new MobileRenderHarness(); const calls: string[] = [];
  const render = (count: number, canAdd: boolean, value: string) => h.render(<BrowseAddHeader canAdd={canAdd}
    onAdd={() => {}} filterCount={count} onFilters={() => calls.push(value)} />);
  try {
    await render(0, true, 'old');
    expect(h.all().filter(n => n.props.accessibilityRole === 'button').map(n => n.props.accessibilityLabel)).toEqual(['Add an asset', 'Filters']);
    expect(h.all().find(n => n.type === 'ListFilterIcon')!.props.color).toBe(lightPalette.text);
    const retained = h.byLabel('Filters')!.props.onPress;
    await render(2, false, 'current');
    expect(h.byLabel('Add an asset')).toBeUndefined();
    expect(h.byLabel('Filters, 2 applied')).toBeDefined();
    expect(h.all().find(n => n.type === 'ListFilterIcon')!.props.color).toBe(lightPalette.action);
    await h.run(retained); expect(calls).toEqual(['current']);
    await render(0, false, 'clear');
    expect(h.all().find(n => n.type === 'ListFilterIcon')!.props.color).toBe(lightPalette.text);
  } finally { await h.unmount(); resetNavigation(); }
});
