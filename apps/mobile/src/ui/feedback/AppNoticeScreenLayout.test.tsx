import { expect, it, vi } from 'vitest';
import { accessibilityAnnouncements, animationStartCount } from '../../test-support/react-native';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { setNativeHeaderHeight } from '../../test-support/react-navigation-elements';
import { AppFeedbackProvider, useAppFeedback, type AppFeedbackContextValue } from './AppFeedback';
import { AppNoticeScreenLayout } from './AppNoticeScreenLayout';

it('moves a retained action between focused content presenters without a root duplicate', async () => {
  const harness = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue; let actions = 0;
  function Source() { feedback = useAppFeedback(); return null; }
  const view = (name: string, transparent: boolean) => <AppFeedbackProvider noticePlacement="screen"><Source /><AppNoticeScreenLayout route={{ name }} options={{ headerTransparent: transparent }}><Source /></AppNoticeScreenLayout></AppFeedbackProvider>;
  try {
    await harness.render(view('details', true));
    await harness.run(() => feedback.showNotice({ tone: 'info', title: 'Saved item', action: { label: 'Undo', onPress: () => { actions++; } } }));
    expect(harness.all().filter(node => node.props.testID === 'app-notice-container')).toHaveLength(1);
    const layer = () => harness.all().find(node => node.props.testID === 'app-notice-layer');
    expect(layer()?.props.style).toContainEqual({ top: 154 });
    await harness.run(() => setNativeHeaderHeight(88));
    expect(layer()?.props.style).toContainEqual({ top: 98 });
    await harness.run(() => setScreenFocused(false));
    expect(harness.byLabel('Undo')).toBeUndefined();
    await harness.render(view('destination', false));
    await harness.run(() => setScreenFocused(true));
    expect(layer()?.props.style).toContainEqual({ top: 10 });
    await harness.press(harness.byLabel('Undo'));
    expect(actions).toBe(1);
    expect(harness.byText('Saved item')).toBeUndefined();
  } finally { await harness.unmount(); setScreenFocused(true); setNativeHeaderHeight(144); }
});

it('leaves nested tab containers to their leaf presenters', async () => {
  const harness = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Source() { feedback = useAppFeedback(); return null; }
  try {
    await harness.render(<AppFeedbackProvider noticePlacement="screen"><Source /><AppNoticeScreenLayout route={{ name: '(tabs)' }} options={{ headerShown: false }}><AppNoticeScreenLayout route={{ name: 'index' }} options={{ headerTransparent: true }}><Source /></AppNoticeScreenLayout></AppNoticeScreenLayout></AppFeedbackProvider>);
    await harness.run(() => feedback.showNotice({ tone: 'info', title: 'One notice' }));
    expect(harness.all().filter(node => node.props.testID === 'app-notice-container')).toHaveLength(1);
  } finally { await harness.unmount(); }
});

it('preserves a plain notice lifetime and announcement through focus handoff', async () => {
  vi.useFakeTimers();
  const harness = new MobileRenderHarness(); let feedback!: AppFeedbackContextValue;
  function Source() { feedback = useAppFeedback(); return null; }
  try {
    await harness.render(<AppFeedbackProvider noticePlacement="screen"><Source /><AppNoticeScreenLayout route={{ name: 'details' }} options={{}}><Source /></AppNoticeScreenLayout></AppFeedbackProvider>);
    await harness.run(() => feedback.showNotice({ tone: 'info', title: 'Handoff lifetime' }));
    const announcements = accessibilityAnnouncements().filter(value => value === 'Handoff lifetime').length;
    const animations = animationStartCount();
    await harness.run(() => vi.advanceTimersByTime(3000));
    await harness.run(() => setScreenFocused(false));
    await harness.run(() => setScreenFocused(true));
    expect(accessibilityAnnouncements().filter(value => value === 'Handoff lifetime')).toHaveLength(announcements);
    expect(animationStartCount()).toBe(animations);
    await harness.run(() => vi.advanceTimersByTime(1200));
    expect(harness.byText('Handoff lifetime')).toBeUndefined();
  } finally { await harness.unmount(); setScreenFocused(true); vi.useRealTimers(); }
});
