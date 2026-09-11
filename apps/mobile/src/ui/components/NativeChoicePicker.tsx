import { useState } from 'react';
import { Pressable, Text } from 'react-native';
import { Check } from 'lucide-react-native';
import { SelectionRow } from './SelectionRow';
import { useAppearancePalette } from '../theme/AppearanceContext';
export type NativeChoicePickerProps = { readonly label: string; readonly value: string; readonly options: readonly { value: string; label: string }[]; readonly disabled?: boolean; readonly onChange: (value: string) => void };
export function NativeChoicePicker({ label, value, options, disabled, onChange }: NativeChoicePickerProps) {
  const colors = useAppearancePalette();
  const [open, setOpen] = useState(false);
  return <SelectionRow label={label} value={options.find(option => option.value === value)?.label ?? 'Choose'} disabled={disabled} expanded={open} onPress={() => setOpen(current => !current)}>
    {options.map(option => <Pressable key={option.value} accessibilityRole="radio" accessibilityLabel={option.label} accessibilityState={{ checked: value === option.value }} disabled={disabled} onPress={() => { onChange(option.value); setOpen(false); }} style={{ minHeight: 48, flexDirection: 'row', alignItems: 'center' }}><Text style={{ color: colors.text, flex: 1 }}>{option.label}</Text>{option.value === value ? <Check color={colors.action} size={20} /> : null}</Pressable>)}
  </SelectionRow>;
}
