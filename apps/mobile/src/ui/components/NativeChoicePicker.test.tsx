import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeChoicePicker } from './NativeChoicePicker';

it('keeps an open choice draft unchanged when editing locks, then accepts choices after unlocking', async () => {
  const h = new MobileRenderHarness(); const selected: string[] = [];
  const props = { label: 'Month', value: '1', options: [{ value: '1', label: 'January' }, { value: '2', label: 'February' }], onChange: (value: string) => selected.push(value) };
  try {
    await h.render(<NativeChoicePicker {...props} />);
    await h.press(h.byLabel('Month'));
    await h.render(<NativeChoicePicker {...props} disabled />);
    await h.run(() => h.byLabel('February')?.props.onPress());
    expect(selected).toEqual([]);
    await h.render(<NativeChoicePicker {...props} />);
    await h.press(h.byLabel('February'));
    expect(selected).toEqual(['2']);
  } finally { await h.unmount(); }
});
