import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeSheetActions } from './NativeSheetActions.ios';

it('lets each footer action adopt its measured padded native height', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<NativeSheetActions primaryLabel="Show results" secondaryLabel="Cancel" disabled={false} onApply={() => {}} onBack={() => {}} />);
    const buttons = h.allByType('StuffStashCommandButton');
    expect(buttons).toHaveLength(2);
    await h.run(() => buttons[0].props.onSizeChange({ nativeEvent: { height: 112 } }));
    await h.run(() => buttons[1].props.onSizeChange({ nativeEvent: { height: 84 } }));
    expect(h.allByType('StuffStashCommandButton').map(button => button.props.style.height)).toEqual([112, 84]);
    expect(h.allByType('StuffStashCommandButton').map(button => button.props.fullWidth)).toEqual([true, true]);
    expect(buttons.map(button => button.props.prominence)).toEqual(['primary', 'secondary']);
  } finally { await h.unmount(); }
});

it('retains current separately enabled footer decisions and retires events on removal', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  const render = (disabled: boolean, secondaryDisabled: boolean, value: string) => h.render(
    <NativeSheetActions primaryLabel="Save" primaryAccessibilityLabel="Save draft" secondaryLabel="Cancel" secondaryAccessibilityLabel="Cancel draft"
      disabled={disabled} secondaryDisabled={secondaryDisabled}
      onApply={() => calls.push('save ' + value)} onBack={() => calls.push('cancel ' + value)} />);
  try {
    await render(false, false, 'old');
    const [save, cancel] = h.allByType('StuffStashCommandButton').map(button => button.props.onPress);
    expect(h.allByType('StuffStashCommandButton').map(button => button.props.accessibilityLabel)).toEqual(['Save draft', 'Cancel draft']);
    await render(true, true, 'busy');
    await h.run(save); await h.run(cancel); expect(calls).toEqual([]);
    await render(true, false, 'invalid');
    await h.run(save); await h.run(cancel); expect(calls).toEqual(['cancel invalid']);
    await render(false, false, 'current');
    await h.run(save); expect(calls).toEqual(['cancel invalid', 'save current']);
    await h.unmount();
    await h.run(save); await h.run(cancel); expect(calls).toEqual(['cancel invalid', 'save current']);
  } finally { await h.unmount(); }
});
