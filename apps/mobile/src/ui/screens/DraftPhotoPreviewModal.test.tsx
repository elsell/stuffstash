import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { latestAlert } from '../../test-support/react-native';
import { setScreenFocused } from '../../test-support/navigation';
import { DraftPhotoPreviewModal } from './DraftPhotoPreviewModal';

it.each(['pending', 'selection', 'collection', 'close', 'visit', 'unmount', 'current', 'only-photo'] as const)('owns draft photo removal across %s', async change => {
  const h = new MobileRenderHarness(); const removed: string[] = []; const indices: (number | undefined)[] = []; let closed = 0;
  const photo = (id: string) => ({ id, uri: `file:///${id}.jpg`, fileName: `${id}.jpg`, contentType: 'image/jpeg' as const, sizeBytes: 10 });
  const photos = change === 'only-photo' ? [photo('one')] : [photo('one'), photo('two')];
  let changed = false;
  const render = () => h.render(<DraftPhotoPreviewModal disabled={changed && change === 'pending'} currentIndex={changed && change === 'close' ? undefined : changed && change === 'selection' ? 1 : 0} photos={changed && change === 'collection' ? [...photos, photo('three')] : photos} onClose={() => { closed++; }} onSetIndex={index => indices.push(index)} onRemovePhoto={id => removed.push(id)} />);
  try {
    await render(); await h.press(h.byLabel('Remove photo'));
    const confirm = latestAlert()?.buttons.find(button => button.text === 'Remove')?.onPress;
    expect(confirm).toBeTypeOf('function');
    changed = true;
    if (change === 'unmount') await h.unmount();
    else if (change === 'visit') { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
    else await render();
    if (change === 'close') { changed = false; await render(); }
    await h.run(() => confirm?.()); await h.run(() => confirm?.());
    const valid = change === 'current' || change === 'only-photo';
    expect(removed).toEqual(valid ? ['one'] : []);
    expect(indices).toEqual(change === 'current' ? [0] : []);
    expect(closed).toBe(change === 'only-photo' ? 1 : 0);
  } finally { await h.unmount(); setScreenFocused(true); }
});
