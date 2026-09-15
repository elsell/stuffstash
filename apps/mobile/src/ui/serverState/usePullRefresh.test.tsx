import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { usePullRefresh } from './usePullRefresh';

it('tracks explicit pulls, ignores duplicate pulls, and clears presentation across navigation', async () => {
  const h = new MobileRenderHarness(); let control!: ReturnType<typeof usePullRefresh>;
  const finishes: (() => void)[] = []; let calls = 0;
  function Screen() { control = usePullRefresh(async () => { calls++; await new Promise<void>(resolve => finishes.push(resolve)); }); return null; }
  try {
    await h.render(<Screen />); expect(control.refreshing).toBe(false);
    let first!: Promise<void>; let second!: Promise<void>;
    await h.run(() => { first = control.refresh(); void control.refresh(); });
    expect(control.refreshing).toBe(true); expect(calls).toBe(1);
    await h.run(() => setScreenFocused(false)); expect(control.refreshing).toBe(false);
    await h.run(() => { void control.refresh(); });
    expect(calls).toBe(1); expect(control.refreshing).toBe(false);
    await h.run(() => setScreenFocused(true)); expect(control.refreshing).toBe(false);
    await h.run(() => { second = control.refresh(); }); expect(calls).toBe(2);
    await h.run(async () => { finishes[0](); await first; }); expect(control.refreshing).toBe(true);
    await h.run(async () => { finishes[1](); await second; }); expect(control.refreshing).toBe(false);
  } finally { await h.unmount(); setScreenFocused(true); }
});
it('ends the pull control after a failed refresh', async () => {
  const h = new MobileRenderHarness(); let control!: ReturnType<typeof usePullRefresh>;
  function Screen() { control = usePullRefresh(async () => { throw new Error('Unavailable'); }); return null; }
  try {
    await h.render(<Screen />);
    await h.run(async () => { await expect(control.refresh()).rejects.toThrow('Unavailable'); });
    expect(control.refreshing).toBe(false);
  } finally { await h.unmount(); }
});
