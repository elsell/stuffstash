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

it('uses native primary emphasis while preserving disabled command behavior and explicit quiet actions', async () => {
  const h = new MobileRenderHarness();
  let calls = 0;
  try {
    await h.render(<NativeCommandButton label="Add item here" prominence="primary" disabled onPress={() => calls++} />);
    let button = h.byType('SwiftUIButton');
    expect(button?.props.modifiers).toContainEqual({ type: 'buttonStyle', value: 'borderedProminent' });
    await h.press(button); expect(calls).toBe(0);
    await h.render(<NativeCommandButton label="Edit" prominence="standard" onPress={() => calls++} />);
    button = h.byType('SwiftUIButton');
    expect(button?.props.modifiers).toContainEqual({ type: 'buttonStyle', value: 'borderless' });
    await h.press(button); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});

it('supports a bounded secondary recovery action with current and disabled handlers', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  try {
    await h.render(<NativeCommandButton label="Retry photos" onPress={() => calls.push('photos')} />);
    expect(h.byType('SwiftUIButton')?.props.modifiers).toContainEqual({ type: 'buttonStyle', value: 'bordered' });
    await h.press(h.byType('SwiftUIButton'));
    await h.render(<NativeCommandButton label="Retry contents" prominence="secondary" disabled onPress={() => calls.push('contents')} />);
    await h.press(h.byType('SwiftUIButton'));
    expect(calls).toEqual(['photos']);
    await h.render(<NativeCommandButton label="Retry contents" prominence="secondary" onPress={() => calls.push('contents')} />);
    await h.press(h.byType('SwiftUIButton'));
    expect(calls).toEqual(['photos', 'contents']);
  } finally { await h.unmount(); }
});
