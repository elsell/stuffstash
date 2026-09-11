import { Host, Picker, Text } from '@expo/ui/swift-ui';
import { disabled as disabledModifier, pickerStyle, tag } from '@expo/ui/swift-ui/modifiers';
import type { NativeChoicePickerProps } from './NativeChoicePicker';
export function NativeChoicePicker({ label, value, options, disabled, onChange }: NativeChoicePickerProps) {
  return <Host style={{ minHeight: 48 }} matchContents={{ vertical: true }}>
    <Picker label={label} selection={value} onSelectionChange={onChange} modifiers={[pickerStyle('menu'), disabledModifier(!!disabled)]}>
      <Text modifiers={[tag('')]}>Choose</Text>
      {options.map(option => <Text key={option.value} modifiers={[tag(option.value)]}>{option.label}</Text>)}
    </Picker>
  </Host>;
}
