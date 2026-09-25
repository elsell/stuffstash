import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { FullScreenPhotoViewer } from './FullScreenPhotoViewer';

it('rejects queued selection events after dismissal, including before parent readback', async () => {
  const h = new MobileRenderHarness();
  const selected: number[] = [];
  let closed = 0;
  const photos = [{ id: 'one', label: 'One', uri: 'one' }, { id: 'two', label: 'Two', uri: 'two' }];
  const render = (index: number | undefined) => <FullScreenPhotoViewer canRemove={false} currentIndex={index}
    photos={photos} onClose={() => closed++} onSelectIndex={value => selected.push(value)} />;
  try {
    await h.render(render(0));
    const button = h.byLabel('Close photo viewer');
    await h.render(render(1));
    expect(h.byLabel('Close photo viewer')).toBe(button);
    const native = h.byType('ImageViewing')!;
    const selection = native.props.onImageIndexChange;
    await h.run(() => { native.props.onRequestClose(); selection(1); });
    expect(closed).toBe(1);
    expect(selected).toEqual([]);
    await h.render(render(undefined));
    await h.run(() => selection(0));
    expect(selected).toEqual([]);
    await h.render(render(1));
    await h.run(() => selection(0));
    expect(selected).toEqual([]);
    await h.run(() => h.byType('ImageViewing')!.props.onImageIndexChange(0));
    expect(selected).toEqual([0]);
  } finally { await h.unmount(); }
});
