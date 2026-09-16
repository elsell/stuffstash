import { useRef } from 'react';
import { Host, TextField } from '@expo/ui/swift-ui';
import { accessibilityLabel, disabled, textFieldStyle } from '@expo/ui/swift-ui/modifiers';
import type { AddDraftNameFieldProps } from './AddDraftNameField.types';

export function AddDraftNameField({ value, editable, accessibilityLabel: label, placeholder, onChangeText }: AddDraftNameFieldProps) {
  const initialValue = useRef(value).current;
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 54 }}>
    <TextField defaultValue={initialValue} placeholder={placeholder} onValueChange={text => {
      if (editable !== false) onChangeText(text);
    }} modifiers={[accessibilityLabel(label), textFieldStyle('roundedBorder'), disabled(editable === false)]} />
  </Host>;
}
