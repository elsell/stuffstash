import { t } from '../../presentation/localization';
import { useEffect, useRef } from 'react';
import { Host, TextField, type TextFieldRef } from '@expo/ui/swift-ui';
import { accessibilityLabel, autocorrectionDisabled, disabled, frame, keyboardType,
  textContentType, textFieldStyle, textInputAutocapitalization } from '@expo/ui/swift-ui/modifiers';
import type { InvitationEmailInputProps } from './InvitationEmailInput.types';

export function InvitationEmailInput({ email, editable, onChangeText }: InvitationEmailInputProps) {
  const seed = useRef(email).current;
  const field = useRef<TextFieldRef>(null);
  useEffect(() => {
    if (!editable) void field.current?.blur();
  }, [editable]);
  return <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
    <TextField ref={field} defaultValue={seed} placeholder={t('mobile.InvitationEmailInputios.friendExampleCom')}
      onValueChange={value => { if (editable) onChangeText(value); }}
      modifiers={[accessibilityLabel(t('mobile.InvitationEmailInputios.inviteeEmail')), keyboardType('email-address'),
        textContentType('emailAddress'), autocorrectionDisabled(), textInputAutocapitalization('never'),
        textFieldStyle('roundedBorder'), disabled(!editable), frame({ minHeight: 54 })]} />
  </Host>;
}
