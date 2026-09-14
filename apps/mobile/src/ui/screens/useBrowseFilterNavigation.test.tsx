import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useBrowseFilterNavigation } from './useBrowseFilterNavigation';

it.each(['cancel', 'unmount', 'scope change'] as const)('ignores late verification after %s and rejects duplicate taps', async reason => {
  const h = new MobileRenderHarness();
  let resolve!: () => void;
  const pending = new Promise<void>(done => { resolve = done; });
  let signal: AbortSignal | undefined; let reads = 0; let navigations = 0;
  let controls!: ReturnType<typeof useBrowseFilterNavigation>;
  const verify = async (request: AbortSignal) => { signal = request; reads++; await pending; };
  function Screen({ scope = 'one' }: { scope?: string }) {
    controls = useBrowseFilterNavigation(scope, verify);
    return null;
  }
  await h.render(<Screen />);
  let first!: Promise<void>;
  await h.run(() => { first = controls.navigate(() => navigations++); void controls.navigate(() => navigations++); });
  expect(reads).toBe(1);
  if (reason === 'cancel') await h.run(() => controls.cancel());
  else if (reason === 'unmount') await h.unmount();
  else await h.render(<Screen scope="two" />);
  expect(signal?.aborted).toBe(true);
  await h.run(async () => { resolve(); await first; });
  expect(navigations).toBe(0);
  await h.unmount();
});
it('navigates once when current verification succeeds and exposes a verification failure', async () => {
  const h = new MobileRenderHarness(); let controls!: ReturnType<typeof useBrowseFilterNavigation>;
  let fail = false; let navigations = 0;
  function Screen() { controls = useBrowseFilterNavigation('one', async () => { if (fail) throw new Error('Denied'); }); return null; }
  await h.render(<Screen />);
  await h.run(() => controls.navigate(() => navigations++));
  expect(navigations).toBe(1);
  fail = true;
  await h.run(() => controls.navigate(() => navigations++));
  expect(navigations).toBe(1); expect(controls.error).toContain('Reopen Browse filters');
  await h.unmount();
});
