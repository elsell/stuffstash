import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import type { VoiceActionPlanProposal } from '../../application/voice/RealtimeVoiceSession';
import { VoicePlanHistorySummary } from './VoicePlanHistorySummary';

it('keeps expiration visible when a review becomes saved history', async () => {
 const harness = new MobileRenderHarness();
 const plan: VoiceActionPlanProposal = { planId: 'plan', status: 'approved', confirmationSummary: 'Add bottle', risks: [], commands: [{ id: 'bottle', kind: 'create_asset', summary: 'Add bottle', title: 'Bottle', expiration: { date: '2028-02', precision: 'month' } }] };
 try {
  await harness.render(<VoicePlanHistorySummary plan={plan} />);
  expect(harness.allText().join(' ')).toContain('Bottle · Expires February 2028');
  await harness.render(<VoicePlanHistorySummary plan={{ ...plan, status: 'executed' }} />);
  expect(harness.allText().join(' ')).toContain('Bottle · Expires February 2028');
  expect(harness.allText()).toContain('Saved');
 } finally { await harness.unmount(); }
});

it('keeps explicit expiration removal visible in history', async () => {
 const harness = new MobileRenderHarness();
 const plan: VoiceActionPlanProposal = { planId: 'plan', status: 'executed', confirmationSummary: 'Remove date', risks: [], commands: [{ id: 'bottle', kind: 'update_asset', summary: 'Remove date', title: 'Bottle', expirationCleared: true }] };
 try {await harness.render(<VoicePlanHistorySummary plan={plan} />);expect(harness.allText().join(' ')).toContain('Bottle · Remove expiration date');} finally {await harness.unmount();}
});
