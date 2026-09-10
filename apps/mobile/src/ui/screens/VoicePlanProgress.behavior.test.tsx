import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { VoicePlanProgress } from './VoicePlanProgress';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
it('exposes measured photo completion through native accessibility', async () => {
 const h = new MobileRenderHarness();
 const state: VoiceRealtimeState = { status: 'processing', tenantName: 'Home', inventoryName: 'Main', progressLabel: 'Adding photos', debugEvents: [], actionPlan: { planId: 'plan', status: 'executed', commands: [], risks: [], confirmationSummary: 'Add item' }, photoAttachmentStatus: { status: 'uploading', message: 'Adding photos', attachedCount: 1, totalCount: 3, failedCount: 0 } };
 try {
  await h.render(<VoicePlanProgress state={state} />);
  expect(h.allText()).toContain('Saved · adding photos');
  expect(h.allText()).toContain('33%');
  expect(h.all().find(node => node.props.accessibilityRole === 'progressbar')?.props.accessibilityValue).toMatchObject({ min: 0, max: 100, now: 33 });
 } finally { await h.unmount(); }
});
