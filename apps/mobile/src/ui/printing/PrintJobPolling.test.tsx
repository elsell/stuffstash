import { expect, it, vi } from 'vitest';
import { PrintJobScreen } from './PrintJobScreen';
import { PrintingFake } from '../../test-support/PrintingFake';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { setAppStateForTest } from '../../test-support/react-native';
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
class ReachablePrinting extends PrintingFake {
  available = true; reads: number[] = [];
  override async job(selected: typeof scope, id: string) {
    this.reads.push(Date.now());
    if (!this.available) throw new Error('Network unavailable');
    return super.job(selected, id);
  }
}
async function setup() {
  vi.useFakeTimers(); vi.setSystemTime(0);
  const h = new MobileRenderHarness(); const fake = new ReachablePrinting();
  fake.submitted.set('job', { id: 'job', printerId: 'printer', status: 'queued', revision: 1, copies: 1, completedCopies: 0 });
  await h.render(<PrintJobScreen workspace={fake.workspace()} scope={scope} jobId="job" canPrint />);
  return { h, fake, advance: async (ms: number) => { await h.run(async () => { await vi.advanceTimersByTimeAsync(ms); }); } };
}
it('stops completed job reads and refreshes when returning to focus', async () => {
  const { h, fake, advance } = await setup();
  try {
    fake.submitted.set('job', { ...fake.submitted.get('job')!, status: 'completed' });
    await advance(5000); await advance(60000); expect(fake.reads).toEqual([0, 5000]);
    await h.run(() => setScreenFocused(false)); await advance(60000);
    await h.run(() => setScreenFocused(true)); expect(fake.reads).toHaveLength(3);
    await advance(60000); expect(fake.reads).toHaveLength(3);
  } finally { await h.unmount(); setScreenFocused(true); vi.useRealTimers(); }
});
it('backs off failed reads with a bounded delay, resets after success, and suspends in background', async () => {
  const { h, fake, advance } = await setup();
  try {
    fake.available = false;
    await advance(5000); await advance(10000); await advance(20000); await advance(40000); await advance(60000);
    expect(fake.reads).toEqual([0, 5000, 15000, 35000, 75000, 135000]);
    fake.available = true; await advance(60000); await advance(5000);
    expect(fake.reads.slice(-2)).toEqual([195000, 200000]);
    await h.run(() => setAppStateForTest('background')); await advance(60000); expect(fake.reads).toHaveLength(8);
    await h.run(() => setAppStateForTest('active')); expect(fake.reads).toHaveLength(9);
    await h.unmount(); await advance(60000); expect(fake.reads).toHaveLength(9);
  } finally { await h.unmount(); setAppStateForTest('active'); vi.useRealTimers(); }
});
it('stops scheduled polling after the user cancels a queued job', async () => {
  const { h, fake, advance } = await setup();
  try {
    await h.press(h.byLabel('Cancel print job')); await h.settle();
    expect(fake.submitted.get('job')?.status).toBe('canceled');
    await advance(60000); expect(fake.reads).toEqual([0]);
  } finally { await h.unmount(); vi.useRealTimers(); }
});
