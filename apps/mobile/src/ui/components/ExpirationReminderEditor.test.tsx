import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationReminderEditor } from './ExpirationReminderEditor';
const policy = { enabled: true, upcoming: true, expired: true, advanceDays: 30 };
it('keeps invalid thresholds unsaved and retains the draft after a failure', async () => {
  const harness = new MobileRenderHarness();
  const saved: unknown[] = [];
  try {
    await harness.render(<ExpirationReminderEditor initialPolicy={policy} onSave={async (value) => { saved.push(value); throw new Error('private backend'); }} />);
    await harness.changeText(harness.byLabel('Days before expiration'), '1.5');
    expect(harness.byLabel('Save reminders')?.props.disabled).toBe(true);
    await harness.changeText(harness.byLabel('Days before expiration'), '14');
    await harness.press(harness.byLabel('Save reminders'));
    expect(saved).toEqual([{ ...policy, advanceDays: 14 }]);
    expect(harness.byLabel('Days before expiration')?.props.value).toBe('14');
  } finally { await harness.unmount(); }
});
it('lets a type override disabled inventory defaults and return to inheritance', async () => {
  const harness = new MobileRenderHarness();
  const saved: unknown[] = [];
  try {
    await harness.render(<ExpirationReminderEditor initialPolicy={null} inheritedPolicy={{ ...policy, enabled: false }} onSave={async (value) => { saved.push(value); }} />);
    expect(harness.byLabel('Enable expiration reminders')?.props.disabled).toBe(true);
    await harness.run(() => harness.byLabel('Use inventory defaults')?.props.onValueChange(false));
    await harness.run(() => harness.byLabel('Enable expiration reminders')?.props.onValueChange(true));
    await harness.press(harness.byLabel('Save reminders'));
    expect(saved).toEqual([policy]);
    await harness.run(() => harness.byLabel('Use inventory defaults')?.props.onValueChange(true));
    await harness.press(harness.byLabel('Save reminders'));
    expect(saved).toEqual([policy, null]);
  } finally { await harness.unmount(); }
});
