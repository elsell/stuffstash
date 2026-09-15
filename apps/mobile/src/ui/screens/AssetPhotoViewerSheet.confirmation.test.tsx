import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { latestAlert } from '../../test-support/react-native';
import { setScreenFocused } from '../../test-support/navigation';
import { assetPhotoViewerModel } from '../components/AssetPhotoWorkspacePresentation';
import { AssetPhotoViewerSheet } from './AssetPhotoViewerSheet';

it.each(['access', 'pending', 'selection', 'collection', 'collection-append', 'close', 'visit', 'unmount', 'current'] as const)('owns photo confirmation across %s', async change => {
  const h = new MobileRenderHarness(); const removed: string[] = [];
  const photos = [{ id: 'one', label: 'One', uri: 'https://photo.invalid/one' }, { id: 'two', label: 'Two', uri: 'https://photo.invalid/two' }];
  let changed = false;
  const render = () => h.render(<AssetPhotoViewerSheet canRemove={!(changed && change === 'access')} isRemoving={changed && change === 'pending'} photos={changed && change === 'collection' ? photos.slice(1) : changed && change === 'collection-append' ? [...photos, { id: 'three', label: 'Three', uri: 'https://photo.invalid/three' }] : photos} model={assetPhotoViewerModel(photos, changed && change === 'close' ? undefined : changed && change === 'selection' ? 'two' : 'one')} onClose={() => {}} onSelectPhoto={() => {}} onRemove={id => removed.push(id)} />);
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
    expect(removed).toEqual(change === 'current' ? ['one'] : []);
  } finally { await h.unmount(); setScreenFocused(true); }
});
