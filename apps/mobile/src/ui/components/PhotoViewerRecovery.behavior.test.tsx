import React from 'react';
import { expect, it } from 'vitest';
import { PhotoViewerLoadError } from './FullScreenPhotoViewer';
import { AssetPhotoViewerSheet } from '../screens/AssetPhotoViewerSheet';
import { assetPhotoViewerModel } from './AssetPhotoWorkspacePresentation';
import { MobileRenderHarness } from '../../test-support/render';

it('explains a failed photo and offers an explicit retry command', async () => {
  const h = new MobileRenderHarness(); let retries = 0;
  try {
    await h.render(<PhotoViewerLoadError onRetry={() => { retries++; }} />);
    expect(h.allText()).toContain('Photo unavailable');
    expect(retries).toBe(0);
    await h.press(h.byLabel('Retry photo'));
    expect(retries).toBe(1);
  } finally { await h.unmount(); }
});

it('preserves image source identity while unrelated removal status changes', async () => {
  const h = new MobileRenderHarness();
  const photos = [{ id: 'one', label: 'One', uri: 'https://photo.invalid/one', fileName: 'One', headers: { Authorization: 'synthetic' } }];
  const render = (isRemoving: boolean) => <AssetPhotoViewerSheet canRemove isRemoving={isRemoving} model={assetPhotoViewerModel(photos, 'one')}
    onClose={() => {}} onSelectPhoto={() => {}} onRemove={() => {}} photos={photos} />;
  try {
    await h.render(render(false));
    const images = h.byType('ImageViewing')?.props.images;
    expect(images).toEqual([{ uri: photos[0].uri, headers: photos[0].headers }]);
    await h.render(render(true));
    expect(h.byType('ImageViewing')?.props.images).toBe(images);
  } finally { await h.unmount(); }
});
