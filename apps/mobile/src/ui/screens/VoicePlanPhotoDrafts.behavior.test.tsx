import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoicePlanPhotoDraftStrip } from './VoicePlanPhotoDrafts';
it('keeps staged photos visible without editing controls after review disconnects', async () => {
 const h = new MobileRenderHarness();
 try {
  await h.render(<VoicePlanPhotoDraftStrip readOnly commandKey="wipes" photos={[{id:'photo',uri:'file:///wipes.jpg',fileName:'wipes.jpg',contentType:'image/jpeg',sizeBytes:10}]} onAddPhotos={() => {throw new Error('read-only');}} onRemovePhoto={() => {throw new Error('read-only');}} />);
  expect(h.allByType('Image')).toHaveLength(1);
  expect(h.byLabel('Remove photo 1')).toBeUndefined();
  expect(h.byLabel('Add photos')).toBeUndefined();
  expect(h.allText()).toContain('Draft kept on this device.');
 } finally { await h.unmount(); }
});

it('removes only the chosen photo through its named native command', async () => {
 const h = new MobileRenderHarness(); const removed: string[][] = []; const added: string[] = [];
 const photos = ['one', 'two'].map(id => ({ id, uri: `file:///${id}.jpg`, fileName: `${id}.jpg`, contentType: 'image/jpeg' as const, sizeBytes: 10 }));
 try {
  await h.render(<VoicePlanPhotoDraftStrip commandKey="supplies" photos={photos}
    onAddPhotos={key => added.push(key)} onRemovePhoto={(key,id) => removed.push([key,id])} />);
  await h.press(h.byLabel('Remove photo 2'));
  expect(removed).toEqual([['supplies','two']]);
  await h.press(h.byLabel('Add photos'));
  expect(added).toEqual(['supplies']);
 } finally { await h.unmount(); }
});
