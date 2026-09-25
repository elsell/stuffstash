import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeCommandButton } from './NativeCommandButton.ios';

it('preserves the native destructive role and prevents disabled destructive actions', async () => {
  const h = new MobileRenderHarness(); let calls = 0;
  try {
    await h.render(<NativeCommandButton label="Delete permanently" role="destructive" disabled onPress={() => calls++} />);
    expect(h.byType('StuffStashCommandButton')?.props.role).toBe('destructive');
    await h.press(h.byType('StuffStashCommandButton')); expect(calls).toBe(0);
    await h.render(<NativeCommandButton label="Delete permanently" role="destructive" onPress={() => calls++} />);
    await h.press(h.byType('StuffStashCommandButton')); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});

it('uses native primary emphasis while preserving disabled command behavior and explicit quiet actions', async () => {
  const h = new MobileRenderHarness();
  let calls = 0;
  try {
    await h.render(<NativeCommandButton label="Add item here" prominence="primary" disabled onPress={() => calls++} />);
    let button = h.byType('StuffStashCommandButton');
    expect(button?.props.prominence).toBe('primary');
    await h.press(button); expect(calls).toBe(0);
    await h.render(<NativeCommandButton label="Edit" prominence="standard" onPress={() => calls++} />);
    button = h.byType('StuffStashCommandButton');
    expect(button?.props.prominence).toBe('standard');
    await h.press(button); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});

it('supports a bounded secondary recovery action with current and disabled handlers', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  try {
    await h.render(<NativeCommandButton label="Retry photos" onPress={() => calls.push('photos')} />);
    expect(h.byType('StuffStashCommandButton')?.props.prominence).toBe('secondary');
    await h.press(h.byType('StuffStashCommandButton'));
    await h.render(<NativeCommandButton label="Retry contents" prominence="secondary" disabled onPress={() => calls.push('contents')} />);
    await h.press(h.byType('StuffStashCommandButton'));
    expect(calls).toEqual(['photos']);
    await h.render(<NativeCommandButton label="Retry contents" prominence="secondary" onPress={() => calls.push('contents')} />);
    await h.press(h.byType('StuffStashCommandButton'));
    expect(calls).toEqual(['photos', 'contents']);
  } finally { await h.unmount(); }
});

it('adopts complete native height and rejects invalid measurements', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<NativeCommandButton label="Cancel new tag" onPress={() => {}} />);
    await h.run(() => h.byType('StuffStashCommandButton')?.props.onSizeChange({ nativeEvent: { height: 132 } }));
    expect(h.byType('StuffStashCommandButton')?.props.style.height).toBe(132);
    await h.run(() => h.byType('StuffStashCommandButton')?.props.onSizeChange({ nativeEvent: { height: -1 } }));
    expect(h.byType('StuffStashCommandButton')?.props.style.height).toBe(132);
    await h.run(() => h.byType('StuffStashCommandButton')?.props.onSizeChange({ nativeEvent: { height: 48 } }));
    expect(h.byType('StuffStashCommandButton')?.props.style.height).toBe(48);
  } finally { await h.unmount(); }
});

it('dispatches retained native events only to the current mounted enabled action', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  await h.render(<NativeCommandButton label="Retry" onPress={() => calls.push('old')} />);
  const retained = h.byType('StuffStashCommandButton')?.props.onPress;
  await h.render(<NativeCommandButton label="Retry" onPress={() => calls.push('current')} />);
  await h.run(retained);
  expect(calls).toEqual(['current']);
  await h.render(<NativeCommandButton label="Retry" disabled onPress={() => calls.push('disabled')} />);
  await h.run(retained);
  await h.unmount();
  await h.run(retained);
  expect(calls).toEqual(['current']);
});
