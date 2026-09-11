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
    await harness.press(harness.byLabel('Before expiration'));
    await harness.changeText(harness.byLabel('Days before expiration'), '1.5');
    expect(harness.byLabel('Save reminder days')?.props.disabled).toBe(true);
    await harness.changeText(harness.byLabel('Days before expiration'), '14');
    await harness.press(harness.byLabel('Save reminder days'));
    expect(saved).toEqual([{ ...policy, advanceDays: 14 }]);
    expect(harness.byLabel('Days before expiration')?.props.value).toBe('14');
  } finally { await harness.unmount(); }
});
it('lets a type override disabled inventory defaults and return to inheritance', async () => {
  const harness = new MobileRenderHarness();
  const saved: unknown[] = [];
  try {
    await harness.render(<ExpirationReminderEditor initialPolicy={null} inheritedPolicy={{ ...policy, enabled: false }} onSave={async (value) => { saved.push(value); }} />);
    expect(harness.byLabel('Days before expiration')).toBeUndefined();
    await harness.change(harness.byType('NativeSegmentedControl'), 'Custom');
    expect(saved).toEqual([policy]);
    await harness.change(harness.byType('NativeSegmentedControl'), 'Use defaults');
    expect(saved).toEqual([policy, null]);
  } finally { await harness.unmount(); }
});
it('refreshes clean policy values while protecting unsaved days', async () => {
 const harness = new MobileRenderHarness();
 const save = async () => {};
 try {
  await harness.render(<ExpirationReminderEditor initialPolicy={policy} onSave={save} />);
  await harness.render(<ExpirationReminderEditor initialPolicy={{ ...policy, advanceDays: 7 }} onSave={save} />);
  await harness.press(harness.byLabel('Before expiration'));
  expect(harness.byLabel('Days before expiration')?.props.value).toBe('7');
  await harness.changeText(harness.byLabel('Days before expiration'), '14');
  await harness.render(<ExpirationReminderEditor initialPolicy={{ ...policy, advanceDays: 10 }} onSave={save} />);
  expect(harness.byLabel('Days before expiration')?.props.value).toBe('14');
 } finally { await harness.unmount(); }
});

it('keeps pending reminder days when another setting is saved', async () => {
 const harness = new MobileRenderHarness(); const saved: unknown[] = [];
 try {
  await harness.render(<ExpirationReminderEditor initialPolicy={policy} onSave={async value => { saved.push(value); }} />);
  await harness.press(harness.byLabel('Before expiration'));
  await harness.changeText(harness.byLabel('Days before expiration'), '14');
  await harness.press(harness.byLabel('Before expiration'));
  const control = harness.byLabel('When expired');
  await harness.run(() => control?.props.onValueChange(false));
  expect(saved).toEqual([{...policy, advanceDays: 14, expired: false}]);
 } finally { await harness.unmount(); }
});
