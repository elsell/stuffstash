import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetDetailRouteErrorState } from './AssetDetailRouteErrorState';

it('offers retry only for recoverable asset failures and retains a scrollable explanation', async () => {
 const h = new MobileRenderHarness(); let retries = 0;
 const render = (canRetry: boolean) => h.render(<AssetDetailRouteErrorState canRetry={canRetry}
  title="Unable to load this place" message="A long explanation that can wrap across several lines."
  onRetry={()=>retries++} />);
 try {
  await render(true);
  expect(h.byLabel('Asset error')?.type).toBe('ScrollView');
  expect(h.byLabel('Asset error')?.props.contentContainerStyle).toMatchObject({flexGrow:1});
  expect(h.allText()).toContain('A long explanation that can wrap across several lines.');
  await h.press(h.byLabel('Retry asset')); expect(retries).toBe(1);
  await render(false); expect(h.byLabel('Retry asset')).toBeUndefined();
 } finally { await h.unmount(); }
});
