import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ExpirationField } from './ExpirationField';
import type { AssetExpiration } from '../../domain/assets/AssetSummary';

it('preserves month precision, reports incomplete input, and clears explicitly', async () => {
  const harness = new MobileRenderHarness();
  const changes: Array<{ value: AssetExpiration | undefined; valid: boolean }> = [];
  try {
    await harness.render(<ExpirationField initialPickerDate={new Date(2028, 0, 1)} onChange={(value, valid) => changes.push({ value, valid })} />);
    expect(changes).toEqual([]);
    await harness.change(harness.byType('NativeSegmentedControl'), 'Month and year');
    await harness.changeText(harness.byLabel('Expiration month'), '2');
    expect(changes.at(-1)?.valid).toBe(false);
    await harness.changeText(harness.byLabel('Expiration year'), '2028');
    expect(changes.at(-1)).toEqual({ value: { date: '2028-02', precision: 'month' }, valid: true });
    await harness.changeText(harness.byLabel('Expiration month'), '13');
    expect(changes.at(-1)?.valid).toBe(false);
    await harness.press(harness.byLabel('Clear expiration'));
    expect(changes.at(-1)).toEqual({ value: undefined, valid: true });
  } finally { await harness.unmount(); }
});

it('does not change the date when the picker is cancelled', async () => {
  const harness = new MobileRenderHarness();
  const changes: unknown[] = [];
  try {
    await harness.render(<ExpirationField initialValue={{ date: '2028-02-29', precision: 'day' }} initialPickerDate={new Date(2028, 0, 1)} onChange={(value) => changes.push(value)} />);
    await harness.press(harness.byLabel('Choose expiration date'));
    await harness.run(() => harness.byType('NativeDatePicker')?.props.onChange({ type: 'dismissed' }));
    expect(changes).toEqual([]);
  } finally { await harness.unmount(); }
});

it('commits the chosen local calendar day only after confirmation and retains precision drafts', async () => {
  const harness = new MobileRenderHarness();
  const changes: unknown[] = [];
  try {
    await harness.render(<ExpirationField initialPickerDate={new Date(2028, 0, 1)} onChange={(value) => changes.push(value)} />);
    await harness.press(harness.byLabel('Choose expiration date'));
    await harness.run(() => harness.byType('NativeDatePicker')?.props.onChange({ type: 'set' }, new Date(2028, 1, 29, 12)));
    expect(changes).toEqual([]);
    await harness.press(harness.byLabel('Use expiration date'));
    expect(changes.at(-1)).toEqual({ date: '2028-02-29', precision: 'day' });
    expect(harness.byLabel('Choose expiration date')?.props.accessibilityValue).toEqual({ text: '2028-02-29' });
    await harness.change(harness.byType('NativeSegmentedControl'), 'Month and year');
    await harness.changeText(harness.byLabel('Expiration month'), '3');
    await harness.changeText(harness.byLabel('Expiration year'), '2029');
    await harness.change(harness.byType('NativeSegmentedControl'), 'Exact date');
    expect(changes.at(-1)).toEqual({ date: '2028-02-29', precision: 'day' });
    await harness.change(harness.byType('NativeSegmentedControl'), 'Month and year');
    expect(changes.at(-1)).toEqual({ date: '2029-03', precision: 'month' });
  } finally { await harness.unmount(); }
});
