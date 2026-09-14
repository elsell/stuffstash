import { Host, Picker, Text } from '@expo/ui/swift-ui';
import { accessibilityLabel as nativeAccessibilityLabel, disabled as disabledModifier, pickerStyle, tag } from '@expo/ui/swift-ui/modifiers';
import type { NativeChoicePickerProps } from './NativeChoicePicker';
import { nativeChoiceOptions } from './NativeChoiceOptions';
export function NativeChoicePicker({ label, accessibilityLabel, includeEmptyOption = true, value, options, disabled, onChange }: NativeChoicePickerProps) {
  return <Host style={{ minHeight: 48 }} matchContents={{ vertical: true }}>
    <Picker label={label} selection={value} onSelectionChange={onChange} modifiers={[pickerStyle('menu'), disabledModifier(!!disabled), nativeAccessibilityLabel(accessibilityLabel ?? label)]}>
      {nativeChoiceOptions(options, includeEmptyOption).map(option => <Text key={option.value} modifiers={[tag(option.value)]}>{option.label}</Text>)}
    </Picker>
  </Host>;
}
