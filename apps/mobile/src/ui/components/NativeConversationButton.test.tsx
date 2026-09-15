import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeConversationButton } from './NativeConversationButton';

it('exposes a named cancellation command and rejects disabled activation', async () => {
  const h = new MobileRenderHarness(); let cancelled = 0;
  try {
    await h.render(<NativeConversationButton kind="cancel" label="Cancel request" onPress={() => { cancelled++; }} />);
    expect(h.byLabel('Cancel request')?.props.accessibilityRole).toBe('button');
    expect(h.allText()).not.toContain('Cancel');
    await h.press(h.byLabel('Cancel request'));
    expect(cancelled).toBe(1);
    await h.render(<NativeConversationButton kind="cancel" label="Cancel request" disabled onPress={() => { cancelled++; }} />);
    await h.press(h.byLabel('Cancel request'));
    expect(cancelled).toBe(1);
  } finally { await h.unmount(); }
});

import { NativeConversationButton as IOSButton } from './NativeConversationButton.ios';
it('uses native symbols and guards disabled native callbacks', async () => {
  const h = new MobileRenderHarness(); let sent = 0;
  try {
    for (const [kind, symbol] of [['record', 'mic.fill'], ['send', 'arrow.up'], ['cancel', 'stop.fill']] as const) {
      await h.render(<IOSButton kind={kind} label="Conversation action" disabled onPress={() => { sent++; }} />);
      const button = h.byType('SwiftUIButton');
      expect(button?.props.systemImage).toBe(symbol);
      expect(button?.props.modifiers).toContainEqual({ type: 'accessibilityLabel', value: 'Conversation action' });
      await h.run(() => button?.props.onPress());
    }
    expect(sent).toBe(0);
  } finally { await h.unmount(); }
});
