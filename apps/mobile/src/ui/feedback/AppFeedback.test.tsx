import { useLayoutEffect } from 'react';
import { Text } from 'react-native';
import { expect, it, vi } from 'vitest';
import { setScreenReaderEnabledForTest, accessibilityAnnouncements, setReduceMotionEnabledForTest, animationStartCount } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';
import { AppFeedbackProvider, useAppFeedback, type AppFeedbackContextValue } from './AppFeedback';

it('does not let a replaced notice dismiss its successor', async () => {
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Probe() { feedback = useAppFeedback(); return null; }
  try {
    await h.render(<AppFeedbackProvider><Probe /></AppFeedbackProvider>);
    await h.run(() => feedback.showNotice({ tone: 'info', title: 'First notice' }));
    const oldDismiss = h.byLabel('First notice. Dismiss message');
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


it('keeps an action available without a time limit', async () => {
  vi.useFakeTimers();
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Probe() { feedback = useAppFeedback(); return null; }
  let retries = 0;
  try {
    await h.render(<AppFeedbackProvider><Probe /></AppFeedbackProvider>);
    await h.run(() => feedback.showNotice({ tone: 'error', title: 'Upload failed', action: { label: 'Retry upload', onPress: () => { retries += 1; } } }));
    await h.run(() => vi.advanceTimersByTime(60000));
    expect(h.byText('Upload failed')).toBeDefined();
    await h.press(h.byLabel('Retry upload'));
    expect(retries).toBe(1);
    expect(h.byText('Upload failed')).toBeUndefined();
  } finally { await h.unmount(); vi.useRealTimers(); }
});

it('keeps and announces feedback when a screen reader is active', async () => {
  vi.useFakeTimers();
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Probe() { feedback = useAppFeedback(); return null; }
  try {
    setScreenReaderEnabledForTest(true);
    await h.render(<AppFeedbackProvider><Probe /></AppFeedbackProvider>);
    await h.run(() => feedback.showNotice({ tone: 'info', title: 'Saved', message: 'Your item is ready' }));
    await h.run(() => vi.advanceTimersByTime(60000));
    expect(h.byLabel('Saved. Your item is ready. Dismiss message')).toBeDefined();
    expect(accessibilityAnnouncements()).toContain('Saved. Your item is ready');
    await h.run(() => setScreenReaderEnabledForTest(false));
    await h.run(() => vi.advanceTimersByTime(4200));
    expect(h.byText('Saved')).toBeUndefined();
  } finally { await h.unmount(); setScreenReaderEnabledForTest(false); vi.useRealTimers(); }
});


it('honors Reduce Motion for entry, preference changes, recovery, and dismissal', async () => {
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Probe() { feedback = useAppFeedback(); return null; }
  try {
    setReduceMotionEnabledForTest(true);
    await h.render(<AppFeedbackProvider><Probe /></AppFeedbackProvider>);
    const initialStarts = animationStartCount();
    await h.run(() => feedback.showNotice({ tone: 'warning', title: 'Check this item' }));
    expect(animationStartCount()).toBe(initialStarts);
    await h.run(() => setReduceMotionEnabledForTest(false));
    expect(animationStartCount()).toBeGreaterThan(initialStarts);
    await h.run(() => setReduceMotionEnabledForTest(true));
    const starts = animationStartCount();
    const banner = h.byLabel('Check this item');
    await h.run(() => banner?.props.onPanResponderRelease({}, { dy: -10, vy: 0 }));
    await h.run(() => banner?.props.onPanResponderTerminate());
    await h.press(h.byLabel('Check this item. Dismiss message'));
    expect(animationStartCount()).toBe(starts);
    expect(h.byText('Check this item')).toBeUndefined();
  } finally { await h.unmount(); setReduceMotionEnabledForTest(false); }
});

it('invalidates old notice content, actions and publishers across service contexts', async () => {
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue; let actions = 0;
  function Probe() { feedback = useAppFeedback(); return null; }
  const view = (scopeKey: string) => <AppFeedbackProvider scopeKey={scopeKey}><Probe /></AppFeedbackProvider>;
  try {
    await h.render(view('server-a-session'));
    const previous = feedback;
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'Saved private item', action: { label: 'Undo private edit', onPress: () => { actions++; } } }));
    const oldAction = h.byLabel('Undo private edit');
    await h.render(view('disconnected'));
    expect(h.byText('Saved private item')).toBeUndefined();
    await h.press(oldAction); expect(actions).toBe(0);
    await h.run(() => previous.showNotice({ tone: 'error', title: 'Late old-server failure' }));
    expect(h.byText('Late old-server failure')).toBeUndefined();
    await h.render(view('server-b-session'));
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'Current item', action: { label: 'View current item', onPress: () => { actions++; } } }));
    expect(h.byText('Current item')).toBeDefined();
    await h.press(h.byLabel('View current item')); expect(actions).toBe(1);
    await h.render(view('server-a-session'));
    await h.run(() => previous.showNotice({ tone: 'error', title: 'Reused old publisher' }));
    expect(h.byText('Reused old publisher')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('preserves same-context navigation notices but invalidates actions when the provider unmounts', async () => {
  const h = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue; let actions = 0;
  function Probe() { feedback = useAppFeedback(); return null; }
  const view = (page: string) => <AppFeedbackProvider scopeKey="same-session"><Probe /><Text>{page}</Text></AppFeedbackProvider>;
  try {
    await h.render(view('Editor'));
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'Saved', action: { label: 'View saved item', onPress: () => { actions++; } } }));
    await h.render(view('Home'));
    expect(h.byText('Saved')).toBeDefined();
    await h.press(h.byLabel('View saved item')); expect(actions).toBe(1);
    await h.run(() => feedback.showNotice({ tone: 'success', title: 'Another saved item', action: { label: 'Old action', onPress: () => { actions++; } } }));
    const oldAction = h.byLabel('Old action');
    await h.render(<Text>Signed out</Text>);
    await h.press(oldAction); expect(actions).toBe(1);
  } finally { await h.unmount(); }
});


it('keeps a new context notice published while old notice state is being cleared', async () => {
  const h = new MobileRenderHarness();
  function Publisher({ label }: { label: string }) {
    const feedback = useAppFeedback();
    useLayoutEffect(() => { feedback.showNotice({ tone: 'warning', title: label }); }, [feedback, label]);
    return null;
  }
  try {
    await h.render(<AppFeedbackProvider scopeKey="old"><Publisher label="Old context" /></AppFeedbackProvider>);
    expect(h.byText('Old context')).toBeDefined();
    await h.render(<AppFeedbackProvider scopeKey="new"><Publisher label="New context" /></AppFeedbackProvider>);
    expect(h.byText('Old context')).toBeUndefined();
    expect(h.byText('New context')).toBeDefined();
  } finally { await h.unmount(); }
});
