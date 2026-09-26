import { expect, it } from 'vitest';
import { NativeChoicePicker } from './NativeChoicePicker.android';
import type { NativeActionMenuProps } from './NativeActionMenu.types';

it('provides the Android menu with current selection, accessible value and native disabled state', () => {
  const selected: string[] = [];
  const props = NativeChoicePicker({ label: 'Availability', accessibilityLabel: 'Choose availability', value: '',
    includeEmptyOption: false, disabled: true, options: [{ value: '', label: 'Any availability' }, { value: 'available', label: 'Available' }],
    onChange: value => selected.push(value) }).props as NativeActionMenuProps;
  expect(props.disabled).toBe(true);
  expect(props.trigger).toEqual({ kind: 'row', label: 'Availability', value: 'Any availability' });
  expect(props.accessibilityLabel).toBe('Choose availability, Any availability');
  expect(props.groups[0].items.map(item => [item.id, item.isSelected])).toEqual([['', true], ['available', false]]);
  // An open menu may deliver a selection after the parent locks editing.
  props.groups[0].items[1].onPress();
  expect(selected).toEqual([]);
  const enabled = NativeChoicePicker({ label: 'Availability', value: '', options: [{ value: 'available', label: 'Available' }], onChange: value => selected.push(value) }).props as NativeActionMenuProps;
  enabled.groups[0].items.find(item => item.id === 'available')?.onPress();
  expect(selected).toEqual(['available']);
});

it('offers the same empty month choice on Android', () => {
  const props = NativeChoicePicker({ label: 'Expiration month', value: '1', options: [{ value: '1', label: 'January' }], onChange: () => {} }).props as NativeActionMenuProps;
  expect(props.groups[0].items.map(item => [item.id, item.label, item.isSelected])).toEqual([['', 'Choose', false], ['1', 'January', true]]);
});
