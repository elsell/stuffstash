import React from 'react';
import { describe, expect, it } from 'vitest';
import { AssetDetailPhotoGallery, assetDetailPhotoPages, assetDetailPhotoWidth } from './AssetDetailPhotoGallery';
import { MobileRenderHarness } from '../../test-support/render';
import type { AssetPhotoViewModel } from '../../application/assets/AssetViewModels';

const photos: readonly AssetPhotoViewModel[] = [
  { id: 'one', label: 'First photo', uri: 'https://example.invalid/one', heroUri: 'https://example.invalid/one-small', heroHeaders: { Authorization: 'synthetic' } },
  { id: 'two', label: 'Second photo', uri: 'https://example.invalid/two' }
];

describe('asset gallery', () => {
  it('explains a failed preview and keeps opening the original photo available', async () => {
    const harness = new MobileRenderHarness(); const opened: string[] = [];
    try {
      await harness.render(<AssetDetailPhotoGallery canAddPhotos imagePlaceholderLabel="Item" photos={photos}
        onAddPhotos={() => {}} onPhotoPress={id => opened.push(id)} />);
      await harness.run(() => harness.allByType('Image')[0]?.props.onError?.({ nativeEvent: { error: 'private URL' } }));
      expect(harness.allText()).toContain('Preview unavailable');
      expect(harness.byLabel('Open photo 1 of 2')?.props.accessibilityValue).toEqual({ text: 'Preview unavailable' });
      expect(harness.allText()).not.toContain('private URL');
      await harness.press(harness.byLabel('Open photo 1 of 2'));
      expect(opened).toEqual(['one']);
      expect(harness.allByType('Image')).toHaveLength(1);
      expect(harness.byLabel('Add photos')).toBeDefined();
    } finally { await harness.unmount(); }
  });

  it('starts fresh for replacement preview credentials and ignores obsolete failure callbacks', async () => {
    const harness = new MobileRenderHarness();
    const render = (currentPhotos: readonly AssetPhotoViewModel[]) => harness.render(
      <AssetDetailPhotoGallery canAddPhotos={false} imagePlaceholderLabel="Item" photos={currentPhotos} />);
    try {
      await render(photos);
      const oldFailure = harness.allByType('Image')[0]?.props.onError;
      await harness.run(() => oldFailure?.());
      expect(harness.allText()).toContain('Preview unavailable');
      await render([{ ...photos[0]!, heroHeaders: { Authorization: 'replacement synthetic' } }, photos[1]!]);
      await harness.run(() => oldFailure?.());
      expect(harness.allText()).not.toContain('Preview unavailable');
      expect(harness.byLabel('Open photo 1 of 2')?.props.accessibilityValue).toBeUndefined();
      expect(harness.allByType('Image')).toHaveLength(2);
      expect(harness.allByType('Image')[0]?.props.source.headers.Authorization).toBe('replacement synthetic');
    } finally { await harness.unmount(); }
  });

  it('names each photo by position and adapts page width to the viewport', () => {
    expect(assetDetailPhotoPages(photos)).toEqual([
      { accessibilityLabel: 'Open photo 1 of 2', positionLabel: '1 of 2' },
      { accessibilityLabel: 'Open photo 2 of 2', positionLabel: '2 of 2' }
    ]);
    expect(assetDetailPhotoWidth(390)).toBe(342);
    expect(assetDetailPhotoWidth(320)).toBe(272);
    expect(assetDetailPhotoWidth(40)).toBe(0);
  });

  it.each([{ photos: [] }, { photos }])('keeps a separate Add photos command for empty and populated galleries', async ({ photos: currentPhotos }) => {
    const harness = new MobileRenderHarness(); let added = 0;
    try {
      await harness.render(<AssetDetailPhotoGallery canAddPhotos imagePlaceholderLabel="Item" photos={currentPhotos}
        onAddPhotos={() => { added++; }} />);
      await harness.press(harness.byLabel('Add photos'));
      expect(added).toBe(1);
      expect(harness.all().filter(node => node.props.accessibilityLabel === 'Add photos')).toHaveLength(1);
      if (!currentPhotos.length) expect(harness.allText()).toContain('No photos');
    } finally { await harness.unmount(); }
  });

  it.each([false, true])('omits Add photos when permission or its handler is absent', async canAddPhotos => {
    const harness = new MobileRenderHarness();
    try {
      await harness.render(<AssetDetailPhotoGallery canAddPhotos={canAddPhotos} imagePlaceholderLabel="Item" photos={photos}
        {...(!canAddPhotos ? { onAddPhotos: () => {} } : {})} />);
      expect(harness.byLabel('Add photos')).toBeUndefined();
    } finally { await harness.unmount(); }
  });

  it('opens the selected photo and retains authenticated thumbnail sources', async () => {
    const harness = new MobileRenderHarness(); const opened: string[] = [];
    try {
      await harness.render(<AssetDetailPhotoGallery canAddPhotos={false} imagePlaceholderLabel="Item" photos={photos}
        onPhotoPress={id => opened.push(id)} />);
      await harness.press(harness.byLabel('Open photo 2 of 2'));
      expect(opened).toEqual(['two']);
      expect(harness.allByType('Image')[0]?.props.source).toEqual({ uri: photos[0]?.heroUri, headers: photos[0]?.heroHeaders });
    } finally { await harness.unmount(); }
  });
});
