import React from 'react';
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
