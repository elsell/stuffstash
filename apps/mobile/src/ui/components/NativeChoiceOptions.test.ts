import { expect, it } from 'vitest';
import { nativeChoiceOptions } from './NativeChoiceOptions';

it('offers an empty month choice without duplicating an explicit Any choice', () => {
  expect(nativeChoiceOptions([{ value: '1', label: 'January' }])).toEqual([
    { value: '', label: 'Choose' }, { value: '1', label: 'January' }
  ]);
  const any = [{ value: '', label: 'Any availability' }, { value: 'available', label: 'Available' }];
  expect(nativeChoiceOptions(any)).toEqual(any);
  expect(nativeChoiceOptions(any, false)).toEqual(any);
});

it('does not offer an invalid empty choice for required filter values', () => {
  const required = [{ value: 'active', label: 'Active' }];
  expect(nativeChoiceOptions(required, false)).toEqual(required);
});
