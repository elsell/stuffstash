import { t } from '../../presentation/localization';
import { forwardRef, useImperativeHandle, useRef } from 'react';
import { Host, TextField, type TextFieldRef } from '@expo/ui/swift-ui';
import { accessibilityLabel, autocorrectionDisabled, disabled, frame, keyboardType,
  textContentType, textFieldStyle, textInputAutocapitalization } from '@expo/ui/swift-ui/modifiers';
import type { InvitationEmailInputHandle, InvitationEmailInputProps } from './InvitationEmailInput.types';

export const InvitationEmailInput = forwardRef<InvitationEmailInputHandle, InvitationEmailInputProps>(function InvitationEmailInput({ email, editable, onChangeText }, ref) {
  const seed = useRef(email).current;
  const field = useRef<TextFieldRef>(null);
  useImperativeHandle(ref, () => ({ blur: () => field.current?.blur() }), []);
  return <Host matchContents={{ vertical: true }} style={{ width: '100%' }}>
    <TextField ref={field} defaultValue={seed} placeholder={t('mobile.InvitationEmailInputios.friendExampleCom')}
      onValueChange={value => { if (editable) onChangeText(value); }}
      modifiers={[accessibilityLabel(t('mobile.InvitationEmailInputios.inviteeEmail')), keyboardType('email-address'),
        textContentType('emailAddress'), autocorrectionDisabled(), textInputAutocapitalization('never'),
        textFieldStyle('roundedBorder'), disabled(!editable), frame({ minHeight: 54 })]} />
  </Host>;
});
