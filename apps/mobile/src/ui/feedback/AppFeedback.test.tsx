import { expect, it, vi } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { AppFeedbackProvider, useAppFeedback, type AppFeedbackContextValue } from './AppFeedback';

it('does not let a replaced notice dismiss its successor', async () => {
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Probe() { feedback = useAppFeedback(); return null; }
  try {
    await h.render(<AppFeedbackProvider><Probe /></AppFeedbackProvider>);
    await h.run(() => feedback.showNotice({ tone: 'info', title: 'First notice' }));
    const oldDismiss = h.byLabel('Dismiss message');
    await h.run(() => feedback.showNotice({ tone: 'info', title: 'New notice' }));
    await h.press(oldDismiss);
    expect(h.byText('New notice')).toBeDefined();
  } finally { await h.unmount(); }
});

it('gives a replacement notice its full display interval', async () => {
  vi.useFakeTimers();
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Probe() { feedback = useAppFeedback(); return null; }
  try {
    await h.render(<AppFeedbackProvider><Probe /></AppFeedbackProvider>);
    await h.run(() => feedback.showNotice({ tone: 'info', title: 'First notice' }));
    await h.run(() => vi.advanceTimersByTime(4000));
    await h.run(() => feedback.showNotice({ tone: 'info', title: 'New notice' }));
    await h.run(() => vi.advanceTimersByTime(300));
    expect(h.byText('New notice')).toBeDefined();
    await h.run(() => vi.advanceTimersByTime(3900));
    expect(h.byText('New notice')).toBeUndefined();
  } finally { await h.unmount(); vi.useRealTimers(); }
});
