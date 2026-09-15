import React from 'react';
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
