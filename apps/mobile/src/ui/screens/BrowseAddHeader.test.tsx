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
it('owns inventory context in the native header and clears it on loading or failure', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<BrowseAddHeader inventoryContext="Garage inventory" canAdd onAdd={() => {}} />);
    const options = navigationOptions().at(-1) as { title: string; headerLeft: () => React.ReactElement };
    expect(options.title).toBe('');
    await h.render(options.headerLeft());
    expect(h.byText('Garage inventory')).toBeDefined();
    await h.render(<BrowseAddHeader inventoryContextStatus="loading" canAdd={false} onAdd={() => {}} />);
    const loading = navigationOptions().at(-1) as { headerLeft: () => React.ReactElement };
    await h.render(loading.headerLeft());
    expect(h.byText('Garage inventory')).toBeUndefined();
    expect(h.byText('Loading inventory…')).toBeDefined();
    let retries = 0;
    await h.render(<BrowseAddHeader inventoryContext="Stale inventory" inventoryContextStatus="error" onRetryInventoryContext={() => retries++} canAdd={false} onAdd={() => {}} />);
    const failed = navigationOptions().at(-1) as { headerLeft: () => React.ReactElement };
    await h.render(failed.headerLeft());
    expect(h.byText('Stale inventory')).toBeUndefined();
    await h.press(h.byLabel('Retry inventory context'));
    expect(retries).toBe(1);
  } finally { await h.unmount(); }
});
