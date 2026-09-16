import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeReadStateButton } from './NativeReadStateButton.ios';

it('updates the native read action and blocks a disabled callback', async () => {
  const h = new MobileRenderHarness(); let calls = 0;
  try {
    await h.render(<NativeReadStateButton read={false} label="Mark medicine read" disabled onPress={() => calls++} />);
    expect(h.byType('SwiftUIButton')?.props.systemImage).toBe('envelope.open');
    expect(h.byType('SwiftUIButton')?.props.modifiers).toContainEqual({ type: 'accessibilityLabel', value: 'Mark medicine read' });
    await h.press(h.byType('SwiftUIButton')); expect(calls).toBe(0);
    await h.render(<NativeReadStateButton read label="Mark medicine unread" onPress={() => calls++} />);
    expect(h.byType('SwiftUIButton')?.props.systemImage).toBe('envelope');
    await h.press(h.byType('SwiftUIButton')); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});
