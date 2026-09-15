import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeChoicePicker } from './NativeChoicePicker.ios';

it('keeps explicit Any distinct from the optional month placeholder and forwards selection', async () => {
  const h = new MobileRenderHarness(); const selected: string[] = [];
  try {
    await h.render(<NativeChoicePicker label="Availability" accessibilityLabel="Choose availability" value="" includeEmptyOption={false}
      options={[{ value: '', label: 'Any availability' }, { value: 'available', label: 'Available' }]} onChange={value => selected.push(value)} />);
    expect(h.allText()).toEqual(['Any availability', 'Available']);
    const picker = h.byType('SwiftUIPicker');
    expect(picker?.props.selection).toBe('');
    expect(picker?.props.modifiers).toContainEqual({ type: 'accessibilityLabel', value: 'Choose availability' });
    await h.run(() => picker?.props.onSelectionChange('available'));
    expect(selected).toEqual(['available']);
    await h.render(<NativeChoicePicker label="Expiration month" value="1" disabled options={[{ value: '1', label: 'January' }]} onChange={() => {}} />);
    expect(h.allText()).toEqual(['Choose', 'January']);
    expect(h.byType('SwiftUIPicker')?.props.modifiers).toContainEqual({ type: 'disabled', value: true });
  } finally { await h.unmount(); }
});

it('rejects native selection events while editing is locked and resumes after unlocking', async () => {
  const h = new MobileRenderHarness(); const selected: string[] = [];
  const props = { label: 'Month', value: '1', options: [{ value: '1', label: 'January' }, { value: '2', label: 'February' }], onChange: (value: string) => selected.push(value) };
  try {
    await h.render(<NativeChoicePicker {...props} />);
    await h.render(<NativeChoicePicker {...props} disabled />);
    await h.run(() => h.byType('SwiftUIPicker')?.props.onSelectionChange('2'));
    expect(selected).toEqual([]);
    await h.render(<NativeChoicePicker {...props} />);
    await h.run(() => h.byType('SwiftUIPicker')?.props.onSelectionChange('2'));
    expect(selected).toEqual(['2']);
  } finally { await h.unmount(); }
});
