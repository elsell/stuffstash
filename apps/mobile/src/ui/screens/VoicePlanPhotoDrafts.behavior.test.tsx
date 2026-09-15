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

it('rejects departed photo source callbacks and suppresses an obsolete selection error', async () => {
  const { showVoicePlanPhotoSourceChooser } = await import('./VoicePlanPhotoDrafts');
  const { latestActionSheetCallback, alertCount } = await import('../../test-support/react-native');
  let current = true; let starts = 0;
  let rejectSelection!: (error: Error) => void;
  const select = () => { starts++; return new Promise<void>((_resolve, reject) => { rejectSelection = reject; }); };
  const options = { isCurrent: () => current, onCamera: select, onLibrary: select };
  showVoicePlanPhotoSourceChooser(options);
  const retained = latestActionSheetCallback();
  current = false; retained?.(0);
  expect(starts).toBe(0);
  current = true; showVoicePlanPhotoSourceChooser(options);
  latestActionSheetCallback()?.(1); expect(starts).toBe(1);
  const before = alertCount(); current = false;
  rejectSelection(new Error('old permission error'));
  await Promise.resolve(); await Promise.resolve();
  expect(alertCount()).toBe(before);
  current = true; showVoicePlanPhotoSourceChooser(options);
  latestActionSheetCallback()?.(0);
  rejectSelection(new Error('current permission error'));
  await Promise.resolve(); await Promise.resolve();
  expect(alertCount()).toBe(before + 1);
});

it('keeps a source choice invalid after returning to a visit or replacing the plan', async () => {
  const { useTaskPresentation } = await import('../navigation/useTaskPresentation');
  const { setScreenFocused } = await import('../../test-support/navigation');
  const { latestActionSheetCallback } = await import('../../test-support/react-native');
  const { showVoicePlanPhotoSourceChooser } = await import('./VoicePlanPhotoDrafts');
  const h = new MobileRenderHarness(); let open!: () => void; let starts = 0;
  function Probe({ plan }: { plan: string }) {
    const begin = useTaskPresentation(undefined, plan);
    open = () => showVoicePlanPhotoSourceChooser({ isCurrent: begin(),
      onCamera: async () => { starts++; }, onLibrary: async () => { starts++; } });
    return null;
  }
  try {
    await h.render(<Probe plan="one" />);
    open(); const oldVisit = latestActionSheetCallback();
    await h.run(() => setScreenFocused(false));
    await h.run(() => setScreenFocused(true));
    oldVisit?.(0); expect(starts).toBe(0);
    open(); const oldPlan = latestActionSheetCallback();
    await h.render(<Probe plan="two" />);
    oldPlan?.(1); expect(starts).toBe(0);
    open(); latestActionSheetCallback()?.(1); expect(starts).toBe(1);
  } finally { await h.unmount(); setScreenFocused(true); }
});
