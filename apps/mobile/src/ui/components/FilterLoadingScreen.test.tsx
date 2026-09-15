import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { FilterLoadingScreen } from './FilterLoadingScreen';

it('allows leaving pending filters without a completed choices query', async () => {
  const h = new MobileRenderHarness(); let cancelled = 0;
  try {
    await h.render(<FilterLoadingScreen onCancel={() => cancelled++} />);
    expect(h.allText()).toContain('Loading filters');
    await h.press(h.byLabel('Cancel'));
    expect(cancelled).toBe(1);
    expect(h.allText()).toContain('Loading filters');
  } finally { await h.unmount(); }
});
