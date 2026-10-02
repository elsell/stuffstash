import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { t } from '../../presentation/localization';
import { AssetRegionRecovery } from './AssetRegionRecovery';

it.each(['contents', 'photos'] as const)('localizes complete %s recovery and preserves retry state', async region => {
  const h = new MobileRenderHarness();
  const keys = region === 'contents'
    ? { failed: 'mobile.AssetRegionRecovery.contentsFailed', loading: 'mobile.AssetRegionRecovery.contentsLoading', retry: 'mobile.AssetRegionRecovery.contentsRetry' } as const
    : { failed: 'mobile.AssetRegionRecovery.photosFailed', loading: 'mobile.AssetRegionRecovery.photosLoading', retry: 'mobile.AssetRegionRecovery.photosRetry' } as const;
  let calls = 0;
  try {
    await h.render(<AssetRegionRecovery region={region} isRetrying={false} onRetry={() => { calls++; }} />);
    expect(h.byText(t(keys.failed))).toBeDefined();
    await h.press(h.byLabel(t(keys.retry)));
    expect(calls).toBe(1);
    await h.render(<AssetRegionRecovery region={region} isRetrying onRetry={() => { calls++; }} />);
    expect(h.byText(t(keys.loading))).toBeDefined();
    await h.press(h.byLabel(t(keys.retry)));
    expect(calls).toBe(1);
  } finally { await h.unmount(); }
});
