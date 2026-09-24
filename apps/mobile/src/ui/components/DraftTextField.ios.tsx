import { Host, TextField } from '@expo/ui/swift-ui';
import { accessibilityHint, accessibilityLabel, disabled, textFieldStyle } from '@expo/ui/swift-ui/modifiers';
import type { DraftTextFieldProps } from './DraftTextField.types';

export function DraftTextField({ value, editable, accessibilityLabel: label, accessibilityHint: hint, placeholder, onChangeText }: DraftTextFieldProps) {
  // Expo seeds native text on every appearance; keep that seed current without
  // issuing setText commands during editing.
  return <Host matchContents={{ vertical: true }} style={{ width: '100%', minHeight: 54 }}>
    <TextField defaultValue={value} placeholder={placeholder} onValueChange={text => {
      if (editable !== false) onChangeText(text);
    }} modifiers={[accessibilityLabel(label), accessibilityHint(hint ?? ''), textFieldStyle('roundedBorder'), disabled(editable === false)]} />
  </Host>;
}
