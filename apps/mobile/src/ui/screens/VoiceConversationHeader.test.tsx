import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { latestAlert } from '../../test-support/react-native';
import { VoiceConversationHeader } from './VoiceConversationHeader';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
import { NavigationOptionFeedback } from '../../test-support/NavigationOptionFeedback';
import { resetNavigation } from '../../test-support/navigation';

it('settles navigation feedback while retaining current close and reset handlers', async () => {
  const h = new MobileRenderHarness(); resetNavigation(); const calls: string[] = [];
  const render = (name: string) => <NavigationOptionFeedback render={() => <VoiceConversationHeader realtime={null}
    photoDrafts={{}} commandDrafts={{}} onReset={() => calls.push(`reset-${name}`)} onClose={() => calls.push(`close-${name}`)} />} />;
  try {
    await h.render(render('old'));
    const close = h.byLabel('Close voice session')!.props.onPress;
    const reset = h.byLabel('New conversation')!.props.onPress;
    await h.render(render('current'));
    await h.run(close); await h.run(reset);
    expect(calls).toEqual(['close-current', 'reset-current']);
  } finally { await h.unmount(); resetNavigation(); }
});

it('protects retryable photos on native New conversation and keeps Close independent', async () => {
  const h = new MobileRenderHarness(); let resets = 0; let closes = 0;
  const realtime: VoiceRealtimeState = { status: 'completed', tenantName: 'Home', inventoryName: 'Main', debugEvents: [],
    actionPlan: { planId: 'plan', status: 'executed', confirmationSummary: 'Saved item', commands: [], risks: [] },
    photoAttachmentStatus: { status: 'partial_failed', message: 'Photo upload failed.', canRetry: true } };
  try {
    await h.render(<VoiceConversationHeader realtime={realtime} photoDrafts={{}} commandDrafts={{}} onReset={() => { resets++; }} onClose={() => { closes++; }} />);
    await h.press(h.byLabel('Close voice session')); expect(closes).toBe(1); expect(resets).toBe(0);
    await h.press(h.byLabel('New conversation')); expect(resets).toBe(0);
    const alert = latestAlert();
    expect(alert?.title).toBe('Start a new conversation?');
    await h.run(() => alert?.buttons?.find(button => button.text === 'New conversation')?.onPress?.());
    expect(resets).toBe(1);
  } finally { await h.unmount(); }
});
