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


it('keeps photo inspection clear and puts metadata and removal in More', async () => {
  const h = new MobileRenderHarness();
  const removed: string[] = []; const selected: number[] = [];
  const photos = [{ id: 'one', label: 'Kitchen.jpg', metadataLabel: '2 MB', uri: 'one' }, { id: 'two', label: 'Shelf.jpg', uri: 'two' }];
  try {
    await h.render(<FullScreenPhotoViewer canRemove currentIndex={0} photos={photos}
      onClose={() => {}} onRemove={photo => removed.push(photo.id!)} onSelectIndex={index => selected.push(index)} />);
    expect(h.byType('ImageViewing')!.props.FooterComponent).toBeUndefined();
    expect(h.byLabel('Next photo')).toBeUndefined();
    expect(h.byLabel('Previous photo')).toBeUndefined();
    expect(h.allText().join(' ')).not.toContain('Kitchen.jpg');
    await h.press(h.byLabel('Photo options'));
    expect(h.allText().join(' ')).toContain('Kitchen.jpg');
    expect(h.allText().join(' ')).toContain('2 MB');
    await h.press(h.byText('Remove photo')?.parent ?? undefined);
    expect(removed).toEqual(['one']);
    await h.run(() => h.byLabel('Photo, 1 of 2')!.props.onAccessibilityAction({ nativeEvent: { actionName: 'increment' } }));
    expect(selected).toEqual([1]);
  } finally { await h.unmount(); }
});
