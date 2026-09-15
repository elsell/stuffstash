import React from 'react';
import { Platform } from 'react-native';
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
    await harness.press(harness.byLabel('Expiration'));
    await harness.change(harness.byType('NativeSegmentedControl'), 'Month and year');
    await harness.press(harness.byLabel('Expiration month'));
    await harness.press(harness.byLabel('February'));
    expect(changes.at(-1)?.valid).toBe(false);
    await harness.changeText(harness.byLabel('Expiration year'), '2028');
    expect(changes.at(-1)).toEqual({ value: { date: '2028-02', precision: 'month' }, valid: true });
    await harness.changeText(harness.byLabel('Expiration year'), '20');
    expect(changes.at(-1)?.valid).toBe(false);
    await harness.press(harness.byLabel('Clear expiration'));
    expect(changes.at(-1)).toEqual({ value: undefined, valid: true });
  } finally { await harness.unmount(); }
});

it('does not change the Android date when the system picker is cancelled', async () => {
  const harness = new MobileRenderHarness();
  const changes: unknown[] = [];
  const platform = Platform.OS; Platform.OS = 'android';
  try {
    await harness.render(<ExpirationField initialValue={{ date: '2028-02-29', precision: 'day' }} initialPickerDate={new Date(2028, 0, 1)} onChange={(value) => changes.push(value)} />);
    await harness.press(harness.byLabel('Expiration'));
    await harness.press(harness.byLabel('Choose expiration date'));
    await harness.run(() => harness.byType('NativeDatePicker')?.props.onChange({ type: 'dismissed' }));
    expect(changes).toEqual([]);
  } finally { await harness.unmount(); Platform.OS = platform; }
});

it('edits the iOS date draft directly through the native picker and retains precision drafts', async () => {
  const harness = new MobileRenderHarness();
  const changes: unknown[] = [];
  try {
    await harness.render(<ExpirationField initialPickerDate={new Date(2028, 0, 1)} onChange={(value) => changes.push(value)} />);
    await harness.press(harness.byLabel('Expiration'));
    expect(changes).toEqual([]);
    await harness.press(harness.byLabel('Add expiration date'));
    expect(changes.at(-1)).toEqual({ date: '2028-01-01', precision: 'day' });
    await harness.run(() => harness.byType('NativeDatePicker')?.props.onChange({ type: 'set' }, new Date(2028, 1, 29, 12)));
    expect(changes.at(-1)).toEqual({ date: '2028-02-29', precision: 'day' });
    expect(harness.byLabel('Use expiration date')).toBeUndefined();
    expect(harness.byType('NativeDatePicker')).toBeDefined();
    await harness.change(harness.byType('NativeSegmentedControl'), 'Month and year');
    expect(changes.at(-1)).toEqual({ date: '2028-02', precision: 'month' });
    await harness.change(harness.byType('NativeSegmentedControl'), 'Exact date');
    expect(changes.at(-1)).toBeUndefined();

  } finally { await harness.unmount(); }
});
it('does not revive a removed expiration when changing precision', async () => {
  const harness = new MobileRenderHarness();
  let value: AssetExpiration | undefined = { date: '2028-02-29', precision: 'day' };
  try {
    await harness.render(<ExpirationField initialValue={value} initialPickerDate={new Date(2028, 0, 1)} onChange={next => { value = next; }} />);
    await harness.press(harness.byLabel('Expiration'));
    await harness.change(harness.byType('NativeSegmentedControl'), 'Month and year');
    expect(value).toEqual({ date: '2028-02', precision: 'month' });
    await harness.press(harness.byLabel('Clear expiration'));
    await harness.press(harness.byLabel('Expiration'));
    await harness.change(harness.byType('NativeSegmentedControl'), 'Exact date');
    expect(value).toBeUndefined();
  } finally { await harness.unmount(); }
});

it('ignores a native date change arriving after the editor becomes disabled', async () => {
  const harness = new MobileRenderHarness(); const changes: unknown[] = [];
  const form = (disabled: boolean) => <ExpirationField disabled={disabled}
    initialValue={{ date: '2028-02-29', precision: 'day' }} initialPickerDate={new Date(2028, 0, 1)}
    onChange={value => changes.push(value)} />;
  try {
    await harness.render(form(false));
    await harness.press(harness.byLabel('Expiration'));
    await harness.render(form(true));
    await harness.run(() => harness.byType('NativeDatePicker')?.props.onChange({ type: 'set' }, new Date(2028, 2, 1)));
    expect(changes).toEqual([]);
  } finally { await harness.unmount(); }
});

it('ignores a year event delivered while disabled and retains the draft after re-enabling', async () => {
  const h = new MobileRenderHarness(); const changes: unknown[] = [];
  const form = (disabled: boolean) => <ExpirationField disabled={disabled}
    initialValue={{ date: '2028-02', precision: 'month' }} initialPickerDate={new Date(2028, 0, 1)}
    onChange={(value, valid) => changes.push({ value, valid })} />;
  try {
    await h.render(form(false));
    await h.press(h.byLabel('Expiration'));
    await h.render(form(true));
    await h.changeText(h.byLabel('Expiration year'), '2030');
    expect(changes).toEqual([]);
    await h.render(form(false));
    expect(h.byLabel('Expiration year')?.props.value).toBe('2028');
    await h.changeText(h.byLabel('Expiration year'), '2029');
    expect(changes).toEqual([{ value: { date: '2029-02', precision: 'month' }, valid: true }]);
  } finally { await h.unmount(); }
});
