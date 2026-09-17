import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeTagColorPicker } from './NativeTagColorPicker.ios';

it('locks the native color well without losing selection and resumes editing after unlock', async () => {
  const h = new MobileRenderHarness();
  const changes: string[] = [];
  const render = (disabled: boolean) => h.render(<NativeTagColorPicker disabled={disabled}
    value="#2E7D32" onChange={value => changes.push(value)} />);
  try {
    await render(true);
    const locked = h.byType('SwiftUIColorPicker');
    expect(locked?.props.selection).toBe('#2E7D32');
    expect(locked?.props.modifiers).toContainEqual({ type: 'disabled', value: true });
    await h.run(() => locked?.props.onSelectionChange?.('#FF7D32'));
    expect(changes).toEqual([]);
    await render(false);
    const unlocked = h.byType('SwiftUIColorPicker');
    expect(unlocked?.props.selection).toBe('#2E7D32');
    expect(unlocked?.props.modifiers).toContainEqual({ type: 'disabled', value: false });
    await h.run(() => unlocked?.props.onSelectionChange('#FF7D32'));
    expect(changes).toEqual(['#FF7D32']);
    await render(true);
    expect(h.byType('SwiftUIColorPicker')?.props.modifiers).toContainEqual({ type: 'disabled', value: true });
  } finally { await h.unmount(); }
});
