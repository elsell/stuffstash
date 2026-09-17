import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeCommandButton } from './NativeCommandButton.ios';

it('preserves the native destructive role and prevents disabled destructive actions', async () => {
  const h = new MobileRenderHarness(); let calls = 0;
  try {
    await h.render(<NativeCommandButton label="Delete permanently" role="destructive" disabled onPress={() => calls++} />);
    expect(h.byType('SwiftUIButton')?.props.role).toBe('destructive');
    await h.press(h.byType('SwiftUIButton')); expect(calls).toBe(0);
    await h.render(<NativeCommandButton label="Delete permanently" role="destructive" onPress={() => calls++} />);
    await h.press(h.byType('SwiftUIButton')); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});

it('uses native primary emphasis while preserving disabled command behavior and quiet defaults', async () => {
  const h = new MobileRenderHarness();
  let calls = 0;
  try {
    await h.render(<NativeCommandButton label="Add item here" prominence="primary" disabled onPress={() => calls++} />);
    let button = h.byType('SwiftUIButton');
    expect(button?.props.modifiers).toContainEqual({ type: 'buttonStyle', value: 'borderedProminent' });
    await h.press(button); expect(calls).toBe(0);
    await h.render(<NativeCommandButton label="Edit" onPress={() => calls++} />);
    button = h.byType('SwiftUIButton');
    expect(button?.props.modifiers).toContainEqual({ type: 'buttonStyle', value: 'borderless' });
    await h.press(button); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});
