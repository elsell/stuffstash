import { Host, LabeledContent, Picker, Text, VStack } from '@expo/ui/swift-ui';
import { useWindowDimensions } from 'react-native';
import { accessibilityLabel as nativeAccessibilityLabel, disabled as disabledModifier, fixedSize, labelsHidden, pickerStyle, tag } from '@expo/ui/swift-ui/modifiers';
import type { NativeChoicePickerProps } from './NativeChoicePicker';
import { nativeChoiceOptions } from './NativeChoiceOptions';

// Default AccessibilityMedium multiplier in pinned React Native 0.83.
const iosAccessibilityTextScale = 1.786;

export function NativeChoicePicker({ label, accessibilityLabel, includeEmptyOption = true, value, options, disabled, onChange }: NativeChoicePickerProps) {
  const stacked = useWindowDimensions().fontScale >= iosAccessibilityTextScale;
  const title = <Text modifiers={[fixedSize({ horizontal: false, vertical: true })]}>{label}</Text>;
  const picker = <Picker label={label} selection={value} onSelectionChange={next => { if (!disabled) onChange(next); }}
    modifiers={[pickerStyle('menu'), disabledModifier(!!disabled), nativeAccessibilityLabel(accessibilityLabel ?? label), ...(stacked ? [labelsHidden()] : [])]}>
    {nativeChoiceOptions(options, includeEmptyOption).map(option => <Text key={option.value} modifiers={[tag(option.value), fixedSize({ horizontal: false, vertical: true })]}>{option.label}</Text>)}
  </Picker>;
  return <Host style={{ minHeight: 48, width: '100%' }} matchContents={{ vertical: true }}>
    {stacked ? <VStack alignment="leading" spacing={8}>{title}{picker}</VStack> : <LabeledContent label={title}>{picker}</LabeledContent>}
  </Host>;
}
