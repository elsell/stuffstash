import { useRef } from 'react';
import { Host, TextField } from '@expo/ui/swift-ui';
import { accessibilityLabel, autocorrectionDisabled, disabled, frame, keyboardType,
  textContentType, textFieldStyle, textInputAutocapitalization } from '@expo/ui/swift-ui/modifiers';
import type { InvitationEmailInputProps } from './InvitationEmailInput.types';

export function InvitationEmailInput({ email, editable, onChangeText }: InvitationEmailInputProps) {
  const seed = useRef(email).current;
  return <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
    <TextField defaultValue={seed} placeholder="friend@example.com"
      onValueChange={value => { if (editable) onChangeText(value); }}
      modifiers={[accessibilityLabel('Invitee email'), keyboardType('email-address'),
        textContentType('emailAddress'), autocorrectionDisabled(), textInputAutocapitalization('never'),
        textFieldStyle('roundedBorder'), disabled(!editable), frame({ minHeight: 54 })]} />
  </Host>;
}
