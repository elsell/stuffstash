import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation, setScreenFocused } from '../../test-support/navigation';
import { VoicePreviewRecovery } from './VoicePreviewRecovery';

it('locks native Retry until completion and retires callbacks when leaving', async () => {
  resetNavigation();
  const h = new MobileRenderHarness(); let calls = 0; let finish!: () => void;
  const retry = () => { calls++; return new Promise<void>(resolve => { finish = resolve; }); };
  try {
    await h.render(<VoicePreviewRecovery message="Context could not load" identity="inventory-one" onRetry={retry} />);
    expect(h.allText()).toContain('Context could not load');
    const action = h.byLabel('Retry conversation')!;
    const retainedPress = action.props.onPress;
    await h.press(action); await h.press(action);
    expect(calls).toBe(1);
    expect(h.byLabel('Retry conversation')!.props.accessibilityState.disabled).toBe(true);
    await h.run(() => finish());
    expect(h.byLabel('Retry conversation')!.props.accessibilityState.disabled).toBe(false);
    await h.run(() => setScreenFocused(false));
    await h.run(retainedPress); expect(calls).toBe(1);
    await h.run(() => setScreenFocused(true));
    await h.run(retainedPress); expect(calls).toBe(1);
    await h.press(h.byLabel('Retry conversation')); expect(calls).toBe(2);
    await h.run(() => finish());
  } finally { await h.unmount(); }
});
