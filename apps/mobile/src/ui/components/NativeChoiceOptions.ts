export type NativeChoiceOption = { readonly value: string; readonly label: string };

export function nativeChoiceOptions(options: readonly NativeChoiceOption[], includeEmptyOption = true): readonly NativeChoiceOption[] {
  return includeEmptyOption && !options.some(option => option.value === '')
    ? [{ value: '', label: 'Choose' }, ...options]
    : options;
}
