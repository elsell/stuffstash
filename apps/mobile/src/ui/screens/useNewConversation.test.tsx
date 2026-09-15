import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { latestAlert } from '../../test-support/react-native';
import { Pressable } from 'react-native';
import { setScreenFocused } from '../../test-support/navigation';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
import { useNewConversation } from './useNewConversation';

it.each(['plan', 'status', 'photos', 'edits', 'visit', 'unmount', 'meter', 'current', 'empty'] as const)('owns New conversation across %s', async change => {
  const h = new MobileRenderHarness(); let resets = 0; let changed = false;
  function Screen() {
    const realtime: VoiceRealtimeState | null = change === 'empty' ? null : { status: 'processing', tenantName: 'Home', inventoryName: 'Main', debugEvents: [], actionPlan: { planId: changed && change === 'plan' ? 'second' : 'first', status: 'proposed', confirmationSummary: 'Save item', commands: [], risks: [] }, recordingLevel: changed && change === 'meter' ? 0.9 : 0.1, photoAttachmentStatus: { status: changed && change === 'status' ? 'attached' : 'uploading', message: '' } };
    const start = useNewConversation(realtime, changed && change === 'photos' ? { one: [{ id: 'one', uri: 'file:///one.jpg', fileName: 'one.jpg', contentType: 'image/jpeg', sizeBytes: 1 }] } : {}, changed && change === 'edits' ? { one: { title: 'New title' } } : {}, () => { resets++; });
    return <Pressable accessibilityLabel="New conversation" onPress={start} />;
  }
  try {
    await h.render(<Screen />); await h.press(h.byLabel('New conversation'));
    if (change === 'empty') { expect(resets).toBe(1); return; }
    expect(resets).toBe(0);
    const confirm = latestAlert()?.buttons.find(button => button.text === 'New conversation')?.onPress;
    expect(confirm).toBeTypeOf('function');
    changed = true;
    if (change === 'unmount') await h.unmount();
    else if (change === 'visit') { await h.run(() => setScreenFocused(false)); await h.run(() => setScreenFocused(true)); }
    else await h.render(<Screen />);
    await h.run(() => confirm?.()); await h.run(() => confirm?.());
    expect(resets).toBe(change === 'current' || change === 'meter' ? 1 : 0);
  } finally { await h.unmount(); setScreenFocused(true); }
});
