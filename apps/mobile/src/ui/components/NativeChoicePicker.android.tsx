import { NativeActionMenu } from './NativeActionMenu';
import type { NativeChoicePickerProps } from './NativeChoicePicker';
import { nativeChoiceOptions } from './NativeChoiceOptions';

export function NativeChoicePicker({ label, accessibilityLabel, includeEmptyOption, value, options, disabled, onChange }: NativeChoicePickerProps) {
  const selected = options.find(option => option.value === value)?.label ?? 'Choose';
  return <NativeActionMenu accessibilityLabel={`${accessibilityLabel ?? label}, ${selected}`} disabled={disabled}
    trigger={{ kind: 'row', label, value: selected }} groups={[{
      id: 'choices', items: nativeChoiceOptions(options, includeEmptyOption).map(option => ({ id: option.value, label: option.label,
        isSelected: value === option.value, disabled, onPress: () => { if (!disabled) onChange(option.value); } }))
    }]} />;
}
