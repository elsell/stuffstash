import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeCommandButton as IOSCommand } from './NativeCommandButton.ios';
import { NativeCommandButton as AndroidCommand } from './NativeCommandButton.android';

it('keeps short native labels and descriptive accessible names with disabled guards', async () => {
  const h = new MobileRenderHarness(); let calls = 0;
  try {
    await h.render(<IOSCommand label="Retry" accessibilityLabel="Retry saving reminders" disabled onPress={() => calls++} />);
    const ios = h.byType('StuffStashCommandButton');
    expect(ios?.props.accessibilityLabel).toBe('Retry saving reminders');
    expect(ios?.props.label).toBe('Retry');
    await h.press(ios); expect(calls).toBe(0);
    await h.render(<AndroidCommand label="Retry" accessibilityLabel="Retry saving reminders" disabled onPress={() => calls++} />);
    const android = h.byType('ComposeOutlinedButton');
    expect(android?.props.modifiers).toContainEqual({$type:'contentDescription', contentDescription:'Retry saving reminders'});
    await h.run(() => android?.props.onClick()); expect(calls).toBe(0);
    await h.render(<AndroidCommand label="Retry" accessibilityLabel="Retry saving reminders" onPress={() => calls++} />);
    await h.run(() => h.byType('ComposeOutlinedButton')?.props.onClick()); expect(calls).toBe(1);
  } finally { await h.unmount(); }
});
