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
    const locked = h.byType('StuffStashColorWell');
    expect(locked?.props.selection).toBe('#2E7D32');
    expect(locked?.props.enabled).toBe(false);
    await h.run(() => locked?.props.onSelectionChange?.({ nativeEvent: { value: '#FF7D32' } }));
    expect(changes).toEqual([]);
    await render(false);
    const unlocked = h.byType('StuffStashColorWell');
    expect(unlocked?.props.selection).toBe('#2E7D32');
    expect(unlocked?.props.enabled).toBe(true);
    await h.run(() => unlocked?.props.onSelectionChange({ nativeEvent: { value: '#FF7D32' } }));
    expect(changes).toEqual(['#FF7D32']);
    await render(true);
    expect(h.byType('StuffStashColorWell')?.props.enabled).toBe(false);
  } finally { await h.unmount(); }
});


it('updates native selection for external edits and clearing without publishing user edits', async () => {
  const h = new MobileRenderHarness();
  const changes: string[] = [];
  try {
    for (const [value, expected] of [['#a1b2c3', '#A1B2C3'], ['', null], ['invalid', null], ['#2E7D32', '#2E7D32']]) {
      await h.render(<NativeTagColorPicker disabled={false} value={value!} onChange={v => changes.push(v)} />);
      expect(h.byType('StuffStashColorWell')?.props.selection).toBe(expected);
    }
    expect(changes).toEqual([]);
  } finally { await h.unmount(); }
});

it('routes retained native events to the current owner and retires them on lock or unmount', async () => {
  const h = new MobileRenderHarness();
  const oldChanges: string[] = [], currentChanges: string[] = [];
  const render = (disabled: boolean, onChange: (value: string) => void) =>
    h.render(<NativeTagColorPicker disabled={disabled} value="" onChange={onChange} />);
  try {
    await render(false, v => oldChanges.push(v));
    const retained = h.byType('StuffStashColorWell')!.props.onSelectionChange;
    const event = { nativeEvent: { value: '#123456' } };
    await render(false, v => currentChanges.push(v));
    await h.run(() => retained(event));
    expect(oldChanges).toEqual([]);
    expect(currentChanges).toEqual(['#123456']);
    await render(true, v => currentChanges.push(v));
    await h.run(() => retained(event));
    expect(currentChanges).toEqual(['#123456']);
    await h.unmount();
    retained(event);
    expect(currentChanges).toEqual(['#123456']);
  } finally { await h.unmount(); }
});
