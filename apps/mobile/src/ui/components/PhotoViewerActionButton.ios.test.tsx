import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { PhotoViewerActionButton } from './PhotoViewerActionButton.ios';

it('dispatches current photo commands and rejects retained disabled or dismissed actions', async () => {
  const h = new MobileRenderHarness();
  const calls: string[] = [];
  try {
    await h.render(<PhotoViewerActionButton action="next" onPress={() => calls.push('old')} />);
    const retained = h.byType('SwiftUIButton')!;
    await h.render(<PhotoViewerActionButton action="next" onPress={() => calls.push('current')} />);
    await h.press(retained);
    expect(calls).toEqual(['current']);
    await h.render(<PhotoViewerActionButton action="next" disabled onPress={() => calls.push('disabled')} />);
    await h.press(retained);
    expect(calls).toEqual(['current']);
    await h.unmount();
    await h.press(retained);
    expect(calls).toEqual(['current']);
  } finally { await h.unmount(); }
});
