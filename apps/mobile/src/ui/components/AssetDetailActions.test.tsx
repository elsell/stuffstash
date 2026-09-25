import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AssetDetailActions } from './AssetDetailActions';

it('groups permitted asset commands and switches checkout to return without exposing denied actions', async () => {
  const h = new MobileRenderHarness();
  const calls: string[] = [];
  const asset = { kind: 'item' as const, canEdit: true, canMove: true, canAddPhotos: true, canCheckout: true, canReturn: false };
  const render = (value = asset, pending = false) => <AssetDetailActions asset={value} isActionPending={pending}
    onAddPhotos={() => calls.push('photos')} onMove={() => calls.push('move')}
    onCheckout={() => calls.push('checkout')} onReturn={() => calls.push('return')} />;
  try {
    await h.render(render());
    expect(h.byLabel('Asset actions')).toBeDefined();
    for (const label of ['Add photos', 'Move', 'Check out']) await h.press(h.byLabel(label));
    expect(calls).toEqual(['photos', 'move', 'checkout']);
    await h.render(render({ ...asset, canCheckout: false, canReturn: true }));
    expect(h.byLabel('Check out')).toBeUndefined();
    await h.press(h.byLabel('Return'));
    expect(calls.at(-1)).toBe('return');
    await h.render(render(asset, true));
    await h.press(h.byLabel('Move'));
    expect(calls).toHaveLength(4);
    await h.render(render({ ...asset, canMove: false, canAddPhotos: false, canCheckout: false }));
    expect(h.byLabel('Asset actions')).toBeUndefined();
  } finally { await h.unmount(); }
});
