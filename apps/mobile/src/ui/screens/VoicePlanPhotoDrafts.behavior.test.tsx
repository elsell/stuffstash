import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoicePlanPhotoDraftStrip } from './VoicePlanPhotoDrafts';
it('keeps staged photos visible without editing controls after review disconnects', async () => {
 const h = new MobileRenderHarness();
 try {
  await h.render(<VoicePlanPhotoDraftStrip readOnly commandKey="wipes" photos={[{id:'photo',uri:'file:///wipes.jpg',fileName:'wipes.jpg',contentType:'image/jpeg',sizeBytes:10}]} onAddPhotos={() => {throw new Error('read-only');}} onRemovePhoto={() => {throw new Error('read-only');}} />);
  expect(h.allByType('Image')).toHaveLength(1);
  expect(h.byLabel('Remove draft photo')).toBeUndefined();
  expect(h.byLabel('Stage draft photos for this planned item')).toBeUndefined();
  expect(h.allText()).toContain('Draft kept on this device.');
 } finally { await h.unmount(); }
});
